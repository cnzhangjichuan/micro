package xwsk

import (
	"encoding/binary"
	"errors"
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"github.com/cnzhangjichuan/micro/xutils"
	"io"
	"mime/multipart"
	"net"
	"sync"
	"time"
)

type wsDpo struct {
	its.BaseDpo

	data []byte
	user types.User
	resp interface{}
	room string

	wClose  bool
	payload [8]byte
	packet  *xutils.Packet
	rspmu   sync.RWMutex
	rspch   chan []byte
}

func (w *wsDpo) Request(v interface{}) error {
	return xutils.UnmarshalJson(w.data, v)
}

func (w *wsDpo) Response(resp interface{}) {
	w.resp = resp
}

func (w *wsDpo) GetUser() types.User {
	return w.user
}

func (w *wsDpo) BindUser(u types.User) {
	w.user = u
	env.sds.Set(w)
}

func (w *wsDpo) UnBindUser() {
	if w.user == nil {
		return
	}
	if env.sds.Del(w) {
		w.user.OnLogout()
	}
	w.user = nil
}

func (w *wsDpo) MoveFileTo(name, dstName string) (string, error) {
	return "", nil
}

func (w *wsDpo) ProcessFile(name string, f func(multipart.File, *multipart.FileHeader) error) error {
	return nil
}

var (
	xErrCloseByClient = errors.New("connection closed by client")
	xErrSendSize      = errors.New("writted data size not eq data len")
)

// load data from connection
func (w *wsDpo) loadDataFromConn(conn net.Conn) (ok bool, err error) {
	const TIMEOUT = time.Minute * 5

	w.packet.Reset()
	for {
		if err = conn.SetReadDeadline(time.Now().Add(TIMEOUT)); err != nil {
			return
		}
		if _, err = conn.Read(w.payload[:1]); err != nil {
			if w.wClose {
				return
			}
			if e, ok := err.(net.Error); ok && e.Timeout() {
				w.wClose = true
				// ping
				conn.Write([]byte{0x89, 0x00})
				continue
			}
			return
		}
		w.wClose = false
		isReadComplete := (w.payload[0] >> 7) > 0
		msgType := w.payload[0] & 0xf
		if _, err = conn.Read(w.payload[:1]); err != nil {
			return
		}
		hasMask := (w.payload[0] >> 7) > 0
		size := int(w.payload[0] & 0x7f)
		switch size {
		case 126:
			if _, err = io.ReadFull(conn, w.payload[:2]); err != nil {
				return
			}
			size = int(binary.BigEndian.Uint16(w.payload[:]))
		case 127:
			if _, err = io.ReadFull(conn, w.payload[:8]); err != nil {
				return
			}
			size = int(binary.BigEndian.Uint64(w.payload[:]))
		}
		// has mask?
		if hasMask {
			if _, err = io.ReadFull(conn, w.payload[:4]); err != nil {
				return
			}
		}
		// has data?
		if size > 0 {
			data := w.packet.Allocate(size)
			if _, err = io.ReadFull(conn, data); err != nil {
				return
			}
			if hasMask {
				for i := 0; i < size; i++ {
					data[i] = data[i] ^ w.payload[i%4]
				}
			}
		}
		switch msgType {
		case 8: // close conn
			err = xErrCloseByClient
			return
		case 9: // ping
			conn.Write([]byte{0x8a, 0x00})
			return
		case 10: //pong
			return
		}
		// data complete?
		if isReadComplete {
			break
		}
	}
	ok = true
	return
}

// split path & data.
func (w *wsDpo) saxPathAndData() string {
	w.data = nil
	data := w.packet.Data()
	for i := 0; i < len(data); i++ {
		if data[i] == '{' {
			w.data = data[i:]
			return xutils.UnsafeBytesToString(data[:i])
		}
	}
	return ""
}

// encoding response data for send.
func (w *wsDpo) encodingResponseData(data []byte) []byte {
	const MaxUInt16 = 1<<16 - 1
	var (
		pack []byte
		size = len(data)
	)
	switch {
	case size < 126:
		pack = make([]byte, 2+size)
		pack[0] = 0x81
		pack[1] = byte(size)
		copy(pack[2:], data)
	case size < MaxUInt16:
		pack = make([]byte, 4+size)
		pack[0] = 0x81
		pack[1] = 126
		binary.BigEndian.PutUint16(pack[2:], uint16(size))
		copy(pack[4:], data)
	default:
		pack = make([]byte, 6+size)
		pack[0] = 0x81
		pack[1] = 127
		binary.BigEndian.PutUint32(pack[2:], uint32(size))
		copy(pack[6:], data)
	}
	return pack
}

// put data into channel
func (w *wsDpo) putResponseData(data []byte) {
	w.rspmu.RLock()
	if w.rspch != nil {
		select {
		default:
		case w.rspch <- data:
		}
	}
	w.rspmu.RUnlock()
}

// send & loop
func (w *wsDpo) sendLoop(conn net.Conn, timeout time.Duration) error {
	for msg := range w.rspch {
		conn.SetWriteDeadline(time.Now().Add(timeout))
		n, err := conn.Write(msg)
		if err != nil {
			return err
		}
		if n != len(msg) {
			return xErrSendSize
		}
	}
	return nil
}

// send message to users
func (w *wsDpo) SendMessage(message interface{}, userId ...string) {
	data, err := xutils.MarshalJson(message)
	if err != nil {
		return
	}
	data = w.encodingResponseData(data)
	if len(userId) == 0 {
		env.sds.Foreach(func(dpo *wsDpo) {
			dpo.putResponseData(data)
		})
	} else {
		for _, uid := range userId {
			if d, ok := env.sds.Get(uid); ok {
				d.putResponseData(data)
			}
		}
	}
}

func (w *wsDpo) SendRoomMessage(message interface{}) {
	data, err := xutils.MarshalJson(message)
	if err != nil {
		return
	}
	data = w.encodingResponseData(data)
	room := w.room
	env.sds.Foreach(func(dpo *wsDpo) {
		if dpo.room == room {
			dpo.putResponseData(data)
		}
	})
}

func SendMessage(message interface{}, userId ...string) {
	data, err := xutils.MarshalJson(message)
	if err != nil {
		return
	}
	encoded := false
	if len(userId) == 0 {
		env.sds.Foreach(func(dpo *wsDpo) {
			if !encoded {
				encoded = true
				data = dpo.encodingResponseData(data)
			}
			dpo.putResponseData(data)
		})
	} else {
		for _, uid := range userId {
			if d, ok := env.sds.Get(uid); ok {
				if !encoded {
					encoded = true
					data = d.encodingResponseData(data)
				}
				d.putResponseData(data)
			}
		}
	}
}

func SendRoomMessage(message interface{}, room string) {
	if room == "" {
		return
	}
	data, err := xutils.MarshalJson(message)
	if err != nil {
		return
	}
	encoded := false
	env.sds.Foreach(func(dpo *wsDpo) {
		if dpo.room == room {
			if !encoded {
				encoded = true
				data = dpo.encodingResponseData(data)
			}
			dpo.putResponseData(data)
		}
	})
}

// close
func (w *wsDpo) Close() {
	w.rspmu.Lock()
	if w.rspch != nil {
		w.UnBindUser()
		close(w.rspch)
		w.rspch = nil
	}
	w.rspmu.Unlock()
}

// join room
func (w *wsDpo) SetRoom(room string) {
	w.room = room
}

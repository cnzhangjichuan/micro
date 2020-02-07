package xwsk

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"github.com/cnzhangjichuan/micro/xutils"
	"net"
	"net/http"
	"time"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	const (
		RESPONSE = "" +
			"HTTP/1.1 101 Web Socket Protocol Handshake\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: %s\r\n" +
			"\r\n"

		TOM = time.Second * 6
	)

	// handshake
	secWebSocket := r.Header.Get("Sec-WebSocket-Key")
	if secWebSocket == "" {
		logError("Sec-WebSocket-Key not set value")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		logError("ws hijack error: %v", err)
		return
	}
	defer conn.Close()

	rsh1 := sha1.Sum([]byte(secWebSocket + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	resp := fmt.Sprintf(RESPONSE, base64.StdEncoding.EncodeToString(rsh1[:]))
	conn.SetWriteDeadline(time.Now().Add(TOM))
	_, err = conn.Write(xutils.UnsafeStringToBytes(resp))
	if err != nil {
		logError("response error: %v", err)
		return
	}

	dpo := &wsDpo{
		packet: xutils.NewPacket(1024),
		rspch:  make(chan []byte, 32),
	}

	// read data
	go func(dpo *wsDpo, conn net.Conn) {
		var (
			ok   bool
			err  error
			path string
		)
		for {
			ok, err = dpo.loadDataFromConn(conn)
			if err != nil {
				break
			}
			if !ok {
				continue
			}
			path = dpo.saxPathAndData()
			if path == "" {
				continue
			}
			handler, ok := env.handlers[path]
			// permission check
			if !ok {
				continue
			}
			if handler.Permit != "" {
				user := dpo.GetUser()
				if user == nil || !user.Access(handler.Permit) {
					continue
				}
			}

			// call logic codes
			dpo.resp = nil
			err = handler.Func(dpo)

			// send data
			if err != nil {
				continue
			}
			if dpo.resp == nil {
				continue
			}
			resp, err := xutils.MarshalJson(dpo.resp)
			if err != nil {
				continue
			}
			dpo.putResponseData(dpo.encodingResponseData(resp))
		}
		dpo.Close()
		if err != nil {
			logError("read data complete: %v", err)
		}
	}(dpo, conn)

	// send data
	err = dpo.sendLoop(conn, TOM)
	if err != nil {
		logError("websocket sendLoop error: %v", err)
	}
	dpo.Close()
}

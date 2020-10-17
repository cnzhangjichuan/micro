package micro

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type lgOutterCaller func(dpo Dpo) error

// SetLogin 设置登入回调
func SetLogin(f lgOutterCaller) {
	var callError = errors.New("login error")
	env.onLogin = func(dpo Dpo) (err error) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			pack := packet.New(1024)
			buf := pack.Allocate(1024)
			buf = buf[:runtime.Stack(buf, false)]
			Debug("\n call login error: %v\n%s\n\n", err, buf)
			packet.Free(pack)
			err = callError
		}()
		err = f(dpo)
		return
	}
}

// SetLogout 设置登出回调
func SetLogout(f lgOutterCaller) {
	var callError = errors.New("logout error")
	env.onLogout = func(dpo Dpo) (err error) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			pack := packet.New(1024)
			buf := pack.Allocate(1024)
			buf = buf[:runtime.Stack(buf, false)]
			Debug("\n call logout error: %v\n%s\n\n", err, buf)
			packet.Free(pack)
			err = callError
		}()
		err = f(dpo)
		return
	}
}

// 会话块的数量
const chunkSize = 16

type websocket struct {
	baseChain

	sChunks  [chunkSize]wsConnChunk
	dpoPool  sync.Pool
	connPool sync.Pool
}

// Init 初始化
func (w *websocket) Init() {
	for i := 0; i < chunkSize; i++ {
		w.sChunks[i].Init()
	}
	w.dpoPool.New = func() interface{} {
		return &wsDpo{}
	}
	w.connPool.New = func() interface{} {
		return &wConn{}
	}
}

// Handle 处理Conn
func (w *websocket) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	if name != "websocket" {
		return false
	}

	var (
		wc  = w.createWConn()
		cac = createDpoCache()
	)

	// 获取远端地址
	remote := pack.HTTPHeaderValue(httpRemoteAddress)
	if remote == "" {
		remote = conn.RemoteAddr().String()
	}

	// 处理握手数据
	uid, isCompress, err := w.handshake(conn, pack, func(protocols []string) (uid string, err error) {
		if env.onLogin == nil {
			return
		}

		// 没有UID
		if len(protocols) < 2 {
			err = errWSInvalidToken
			return
		}

		// 校验身份
		var ok bool
		uid, ok = env.authorize.Check(strings.TrimSpace(protocols[1]))
		if !ok {
			err = errWSInvalidToken
			return
		}

		// 调用登入
		dpo := w.createDpo()
		dpo.uid = uid
		dpo.pack = pack
		dpo.cache = cac
		dpo.group = &wc.group
		err = env.onLogin(dpo)
		w.freeDpo(dpo)
		return
	})
	if err != nil {
		freeDpoCache(cac)
		w.freeWConn(wc)
		Debug("handshake error %v", err)
		return true
	}

	// 将自身注册到会话中(初始化)
	if uid != "" {
		wc.uid = uid
		wc.conn = conn
		wc.isCompressed = isCompress
		w.RegisterConn(wc)
	}

	// 处理数据
	var payload = make([]byte, 8)
	for {
		// 读取数据
		err = w.decodeWebsocket(conn, pack, payload)
		if err != nil {
			break
		}
		api := xutils.UnsafeBytesToString(pack.ReadWhen('{'))
		dpo := w.createDpo()
		dpo.uid = uid
		dpo.cache = cac
		dpo.pack = pack
		dpo.group = &wc.group
		dpo.SetRemote(remote)

		// 调用业务接口
		if resp, _ := w.callAPI(conn, dpo, api); resp != nil {
			w.encodingResponseData(dpo.pack, api, resp, isCompress)
			ad := env.sender.NewAutoData(dpo.pack.Copy())
			env.sender.AddConnData(conn, ad)
			env.sender.FreeAutoData(ad)
		}

		// 将自身注册到会话中(切换账号)
		if dpo.uid != "" && dpo.uid != uid {
			uid = dpo.uid
			if wc.uid != "" {
				w.UnRegisterConn(wc)
			}
			wc.uid = uid
			wc.conn = conn
			wc.isCompressed = isCompress
			w.RegisterConn(wc)
		}
		w.freeDpo(dpo)
	}

	// 登出
	if env.onLogout != nil {
		dpo := w.createDpo()
		dpo.uid = uid
		dpo.cache = cac
		dpo.pack = pack
		dpo.SetRemote(remote)
		err = env.onLogout(dpo)
		if err != nil {
			Debug("logout error %v", err)
		}
		w.freeDpo(dpo)
	}

	// 释放资源
	freeDpoCache(cac)
	w.UnRegisterConn(wc)
	w.freeWConn(wc)

	return true
}

// Close 关闭
func (w *websocket) Close() {
	for i := 0; i < chunkSize; i++ {
		w.sChunks[i].Close()
	}
}

// callAPI 调用业务接口
func (w *websocket) callAPI(conn net.Conn, dpo *wsDpo, api string) (interface{}, bool) {
	// 没有登录
	if !env.authorize.CheckAPI(dpo.uid, api) {
		return noLoginError, false
	}

	// 没有发现业务接口
	bis, ok := findBis(api)
	if !ok {
		return apiNotFoundError, false
	}

	resp, errCode := bis(dpo)
	if errCode != "" {
		// 业务发生错误
		return &errBisResp{ErrCode: errCode}, false
	}

	// 返回业务数据
	return resp, true
}

// handshake 处理握手
func (w *websocket) handshake(conn net.Conn, pack *packet.Packet, onLogin func([]string) (string, error)) (uid string, isCompress bool, err error) {
	protocols := strings.Split(pack.HTTPHeaderValue(wsProtocol), ",")
	isCompress = strings.TrimSpace(protocols[0]) == "compress"
	uid, err = onLogin(protocols)
	if err != nil {
		return
	}
	secWK := pack.HTTPHeaderValue(wsKey)
	rsh1 := sha1.Sum(xutils.UnsafeStringToBytes(secWK + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	secWK = base64.StdEncoding.EncodeToString(rsh1[:])

	pack.ReadHTTPBody(conn)
	pack.Reset()
	pack.Write(wsRespOK)

	// sec-websocket-accept
	pack.Write(wsAccept)
	pack.Write(xutils.UnsafeStringToBytes(secWK))
	pack.Write(httpRowAt)
	pack.Write(httpRowAt)

	_, err = pack.FlushToConn(conn)
	return
}

// encodingResponseData 将要发送的数据进行编码，使之适合websocket协议
func (w *websocket) encodingResponseData(pack *packet.Packet, api string, v interface{}, isCompress bool) {
	pack.Reset()
	pack.Allocate(10)
	pack.EncodeJSONWithAPI(v, isCompress, isCompress, xutils.UnsafeStringToBytes(api))
	size := pack.Size() - 10
	prefix := pack.Slice(0, 10)

	offset := int(0)
	switch {
	case size < 126:
		offset = 8
		prefix[offset] = 0x81
		prefix[offset+1] = byte(size)
	case size < math.MaxUint16:
		offset = 6
		prefix[offset] = 0x81
		prefix[offset+1] = 126
		binary.BigEndian.PutUint16(prefix[offset+2:], uint16(size))
	default:
		offset = 0
		prefix[offset] = 0x81
		prefix[offset+1] = 127
		binary.BigEndian.PutUint64(prefix[offset+2:], uint64(size))
	}
	pack.Seek(offset, -1)
}

// decodeWebsocket 从流中读出数据
func (w *websocket) decodeWebsocket(conn net.Conn, pack *packet.Packet, payload []byte) (err error) {
	const RT = time.Minute

	wClose := false
	pack.Reset()

	for {
		// FIN/RSV1/RSV2/RSV3/OPCODE(4bits)
		conn.SetReadDeadline(time.Now().Add(RT))
		_, err = io.ReadFull(conn, payload[:2])
		if err != nil {
			if wClose {
				return
			}
			if e, ok := err.(net.Error); ok && e.Timeout() {
				wClose = true
				// ping
				conn.SetWriteDeadline(time.Now().Add(time.Second))
				conn.Write(wsPing)
				continue
			}
			return
		}
		wClose = false
		fin := payload[0]>>7 == 1
		//rv1 := (payload[0]>>6)&1 == 1
		opCode := payload[0] & 0xf

		// MASH/Size(7bits)
		hasMask := payload[1]>>7 == 1
		size := int(payload[1] & 0x7f)
		switch size {
		case 126:
			_, err = io.ReadFull(conn, payload[:2])
			if err != nil {
				return
			}
			size = int(binary.BigEndian.Uint16(payload))
		case 127:
			_, err = io.ReadFull(conn, payload[:8])
			if err != nil {
				return
			}
			size = int(binary.BigEndian.Uint64(payload))
		}

		// 读取掩码
		if hasMask {
			_, err = io.ReadFull(conn, payload[:4])
			if err != nil {
				return
			}
		}

		// 读入数据
		if size > 0 {
			s := pack.Size()
			data := pack.Allocate(size)
			if _, err = io.ReadFull(conn, data); err != nil {
				return
			}
			if hasMask {
				pack.Mask(payload[:4], s, pack.Size())
			}
		}

		// 关闭连接
		if opCode == 8 {
			err = io.EOF
			return
		}

		// Ping
		if opCode == 9 {
			conn.SetWriteDeadline(time.Now().Add(time.Second))
			conn.Write(wsPong)
			pack.Seek(0, pack.Size()-size)
			continue
		}

		// Pong
		if opCode == 10 {
			pack.Seek(0, pack.Size()-size)
			continue
		}

		if fin {
			break
		}
	}
	return
}

type wsDpo struct {
	baseDpo

	pack *packet.Packet
}

// Parse 获取客户端参数
func (w *wsDpo) Parse(v interface{}) {
	if w.pack == nil {
		return
	}
	err := w.pack.DecodeJSON(v)
	if err != nil {
		Debug("websocket dpo parse data error: %v", err)
	}
}

// CreateDpo 创建处理对象
func (w *websocket) createDpo() *wsDpo {
	return w.dpoPool.Get().(*wsDpo)
}

// FreeDpo 释放处理对象
func (w *websocket) freeDpo(dpo *wsDpo) {
	dpo.pack = nil
	dpo.release()
	w.dpoPool.Put(dpo)
}

// wsConnChunk conn会话单元
type wsConnChunk struct {
	sync.RWMutex

	m      map[string]*wConn
	worker chan func()
}

type wConn struct {
	conn         net.Conn
	uid          string
	isCompressed bool
	group        tUserDpoGroup
}

// createWConn 创建Conn
func (w *websocket) createWConn() *wConn {
	return w.connPool.Get().(*wConn)
}

// freeWConn 释放Conn
func (w *websocket) freeWConn(c *wConn) {
	if c == nil {
		return
	}
	c.conn = nil
	c.uid = ""
	c.isCompressed = false
	c.group.clear()
	w.connPool.Put(c)
}

// Init 初始化
func (w *wsConnChunk) Init() {
	const CHUNKSIZE = 128

	w.m = make(map[string]*wConn, CHUNKSIZE)
	w.worker = make(chan func(), 512)
	go func(wk <-chan func()) {
		for f := range w.worker {
			w.RLock()
			f()
			w.RUnlock()
		}
	}(w.worker)
}

// Close 关闭会话
func (w *wsConnChunk) Close() {
	w.Lock()
	if w.worker != nil {
		close(w.worker)
		w.worker = nil
	}
	w.Unlock()
}

// RegisterConn 注册到会话中
func (w *websocket) RegisterConn(c *wConn) {
	if c == nil {
		return
	}
	idx := xutils.HashCode32(c.uid) % chunkSize
	w.sChunks[idx].Register(c)
}

// Register 注册
func (w *wsConnChunk) Register(c *wConn) {
	w.Lock()
	if c1, ok := w.m[c.uid]; ok && c1 != c {
		c1.conn.Close()
	}
	w.m[c.uid] = c
	w.Unlock()
}

// UnRegisterConn 从会话中注销
func (w *websocket) UnRegisterConn(c *wConn) {
	if c == nil {
		return
	}
	idx := xutils.HashCode32(c.uid) % chunkSize
	w.sChunks[idx].UnRegister(c)
}

// UnRegister 注销
func (w *wsConnChunk) UnRegister(c *wConn) {
	w.Lock()
	if c1, ok := w.m[c.uid]; ok && c1 == c {
		delete(w.m, c.uid)
	}
	w.Unlock()
}

// SendData 发送数据
func (w *websocket) SendData(v interface{}, api string, uids []string) {
	// 创建数据包
	crGzd := func(isCompressed bool) *wkAutoData {
		pack := packet.New(2048)
		w.encodingResponseData(pack, api, v, isCompressed)
		return env.sender.NewAutoData(pack)
	}

	// 添加到worker中
	for i := 0; i < chunkSize; i++ {
		w.sChunks[i].SendData(crGzd, uids)
	}
}

// SendData 发送数据
func (w *wsConnChunk) SendData(crGzd func(bool) *wkAutoData, uids []string) {
	w.RLock()
	if w.worker != nil {
		w.worker <- func() {
			var gzd, nzd *wkAutoData
			if len(uids) > 0 {
				// 按用户发送
				for _, uid := range uids {
					if m, ok := w.m[uid]; ok {
						if m.isCompressed {
							if gzd == nil {
								gzd = crGzd(true)
							}
							env.sender.AddConnData(m.conn, gzd)
						} else {
							if nzd == nil {
								nzd = crGzd(false)
							}
							env.sender.AddConnData(m.conn, nzd)
						}
					}
				}
			} else {
				// 全部发送
				for _, m := range w.m {
					if m.isCompressed {
						if gzd == nil {
							gzd = crGzd(true)
						}
						env.sender.AddConnData(m.conn, gzd)
					} else {
						if nzd == nil {
							nzd = crGzd(false)
						}
						env.sender.AddConnData(m.conn, nzd)
					}
				}
			}
			if gzd != nil {
				env.sender.FreeAutoData(gzd)
			}
			if nzd != nil {
				env.sender.FreeAutoData(nzd)
			}
		}
	}
	w.RUnlock()
}

// SendGroup 按组发送数据
func (w *websocket) SendGroup(v interface{}, api string, flag uint8, group string) {
	// 创建数据包
	crGzd := func(isCompressed bool) *wkAutoData {
		pack := packet.New(2048)
		w.encodingResponseData(pack, api, v, isCompressed)
		return env.sender.NewAutoData(pack)
	}

	// 添加到worker中
	for i := 0; i < chunkSize; i++ {
		w.sChunks[i].SendGroup(crGzd, flag, group)
	}
}

// SendGroup 按组发送数据
func (w *wsConnChunk) SendGroup(crGzd func(bool) *wkAutoData, flag uint8, group string) {
	w.RLock()
	if w.worker != nil {
		w.worker <- func() {
			var gzd, nzd *wkAutoData
			for _, m := range w.m {
				if !m.group.Match(flag, group) {
					continue
				}
				if m.isCompressed {
					if gzd == nil {
						gzd = crGzd(true)
					}
					env.sender.AddConnData(m.conn, gzd)
				} else {
					if nzd == nil {
						nzd = crGzd(false)
					}
					env.sender.AddConnData(m.conn, nzd)
				}
			}
			if gzd != nil {
				env.sender.FreeAutoData(gzd)
			}
			if nzd != nil {
				env.sender.FreeAutoData(nzd)
			}
		}
	}
	w.RUnlock()
}

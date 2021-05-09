package micro

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"io"
	"math"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type lgCaller func(dpo Dpo)

// SetLogin 设置登入回调
func SetLogin(f lgCaller) {
	env.onLogin = func(dpo Dpo) {
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
		}()
		f(dpo)
	}
}

// SetLogout 设置登出回调
func SetLogout(f lgCaller) {
	env.onLogout = func(dpo Dpo) {
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
		}()
		f(dpo)
	}
}

// 会话块的数量
const (
	chunkSize  = 16
	workerSize = 16
)

type websocket struct {
	baseChain

	dpoPool sync.Pool

	// 会话
	session struct {
		chunks [chunkSize]struct {
			sync.RWMutex
			m map[string]*wConn
		}
		pool sync.Pool
	}

	// 数据发送器
	sender struct {
		sync.RWMutex
		seq uint32
		wgo [workerSize]chan *wkConnData
		adp sync.Pool
		cdp sync.Pool
	}
}

type wConn struct {
	conn         net.Conn
	uid          string
	isCompressed bool
	group        tUserDpoGroup
}

// Init 初始化
func (w *websocket) Init() {
	for i := 0; i < chunkSize; i++ {
		w.session.chunks[i].m = make(map[string]*wConn, 256)
	}
	w.dpoPool.New = func() interface{} {
		return &wsDpo{}
	}
	w.session.pool.New = func() interface{} {
		return &wConn{}
	}
	w.sender.adp.New = func() interface{} {
		return &wkAutoData{}
	}
	w.sender.cdp.New = func() interface{} {
		return &wkConnData{}
	}
	for i := 0; i < workerSize; i++ {
		w.sender.wgo[i] = make(chan *wkConnData, 512)
		go func(c <-chan *wkConnData) {
			const WT = time.Second * 10

			for wcd := range c {
				wcd.ad.pack.SetTimeout(0, WT)
				wcd.ad.pack.FlushToConn(wcd.conn)
				w.freeConnData(wcd)
			}
		}(w.sender.wgo[i])
	}
}

// Handle 处理Conn
func (w *websocket) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	if name != "websocket" {
		return false
	}

	var (
		wc          *wConn
		cac         dpoCache
		senderIndex uint32 = workerSize
	)

	// 获取远端地址
	remote := pack.HTTPHeaderValue(httpRemoteAddress)
	if remote == "" {
		remote = conn.RemoteAddr().String()
	}

	if env.onLogin != nil {
		// 如果设置了登入函数，需要校验登入Token

		// 处理握手数据
		uid, isCompress, err := w.handshake(conn, pack)
		if err != nil || uid == "" {
			return true
		}
		cac := createDpoCache()

		// 将自身注册到会话中
		wc := w.createWConn()
		wc.uid = uid
		wc.conn = conn
		wc.isCompressed = isCompress
		if w.RegisterConn(wc) && env.onLogin != nil {
			// 调用登入
			dpo := w.createDpo()
			dpo.uid = uid
			dpo.pack = pack
			dpo.cache = cac
			dpo.group = &wc.group
			env.onLogin(dpo)
			w.freeDpo(dpo)
		}

		// 处理数据
		var payload = make([]byte, 8)
		for {
			// 组装业务参数
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
			if resp := w.callAPI(dpo, api); resp != nil {
				w.encodingResponseData(dpo.pack, api, resp, isCompress)
				ad := w.NewRespAutoData(dpo.pack.Copy())
				senderIndex = w.AddRespConnData(conn, ad, senderIndex)
				w.freeAutoData(ad)
			}
			w.freeDpo(dpo)
		}
	} else {
		// 不需要登入Token, 一般用于网页端的直接接入

		// 处理握手数据
		uid, isCompress, err := w.handshake(conn, pack)
		if err != nil {
			return true
		}
		cac := createDpoCache()
		wc := w.createWConn()

		// 处理数据
		var payload = make([]byte, 8)
		for {
			// 组装业务参数
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

			// 校验登入状态
			if !env.authorize.CheckAPI(dpo.uid, api) {
				w.encodingResponseData(dpo.pack, api, apiNotFoundError, isCompress)
				ad := w.NewRespAutoData(dpo.pack.Copy())
				senderIndex = w.AddRespConnData(conn, ad, senderIndex)
				w.freeAutoData(ad)
			} else {
				// 调用业务接口
				resp := w.callAPI(dpo, api)

				// 将自身注册到会话中
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

				// 发送响应数据
				if resp != nil {
					w.encodingResponseData(dpo.pack, api, resp, isCompress)
					ad := w.NewRespAutoData(dpo.pack.Copy())
					senderIndex = w.AddRespConnData(conn, ad, senderIndex)
					w.freeAutoData(ad)
				}
			}
			w.freeDpo(dpo)
		}
	}

	// 释放资源
	if w.UnRegisterConn(wc) && env.onLogout != nil {
		dpo := w.createDpo()
		dpo.uid = wc.uid
		dpo.cache = cac
		dpo.pack = pack
		dpo.group = &wc.group
		dpo.SetRemote(remote)
		env.onLogout(dpo)
		w.freeDpo(dpo)
	}
	w.freeWConn(wc)
	freeDpoCache(cac)

	return true
}

// Close 关闭
func (w *websocket) Close() {
	w.sender.Lock()
	for i := 0; i < workerSize; i++ {
		close(w.sender.wgo[i])
	}
	w.sender.Unlock()
}

// callAPI 调用业务接口
func (w *websocket) callAPI(dpo *wsDpo, api string) interface{} {
	// 没有发现业务接口
	bis, ok := findBis(api)
	if !ok {
		return apiNotFoundError
	}

	resp, errCode := bis(dpo)
	// 业务发生错误
	if errCode != "" {
		return &errBisResp{ErrCode: errCode}
	}
	// 返回业务数据
	return resp
}

// handshake 处理握手
func (w *websocket) handshake(conn net.Conn, pack *packet.Packet) (uid string, isCompress bool, err error) {
	protocols := strings.Split(pack.HTTPHeaderValue(wsProtocol), ",")
	isCompress = strings.TrimSpace(protocols[0]) == "compress"
	if env.onLogin != nil {
		// 没有设置token
		if len(protocols) < 2 {
			err = errWSInvalidToken
			return
		}
		// 校对Token值
		var ok bool
		uid, ok = env.authorize.CheckToken(strings.TrimSpace(protocols[1]))
		if !ok {
			err = errWSInvalidToken
			return
		}
	} else {
		if len(protocols) > 1 {
			uid = strings.TrimSpace(protocols[1])
		}
	}

	secWK := pack.HTTPHeaderValue(wsKey)
	rsh1 := sha1.Sum(xutils.UnsafeStringToBytes(secWK + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	secWK = base64.StdEncoding.EncodeToString(rsh1[:])

	pack.ReadHTTPBody(conn)
	pack.Reset()
	pack.Write(wsRespOK)
	if env.onLogin != nil {
		pack.Write(wsProtocol)
		pack.Write(xutils.UnsafeStringToBytes(protocols[0]))
		pack.Write(httpRowAt)
	}
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
	pack.EncodeJSONApi(v, isCompress, isCompress, xutils.UnsafeStringToBytes(api))
	size := pack.Size() - 10
	prefix := pack.Slice(0, 10)

	offset := 0
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
	dpo.group = nil
	dpo.release()
	w.dpoPool.Put(dpo)
}

// createWConn 创建Conn
func (w *websocket) createWConn() *wConn {
	return w.session.pool.Get().(*wConn)
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
	w.session.pool.Put(c)
}

type wkAutoData struct {
	ref  int64
	pack *packet.Packet
}

// freeAutoData 释放AutoData
func (w *websocket) freeAutoData(ad *wkAutoData) {
	if ad == nil {
		return
	}
	if atomic.AddInt64(&ad.ref, -1) == 0 {
		packet.Free(ad.pack)
		w.sender.adp.Put(ad)
	}
}

type wkConnData struct {
	conn net.Conn
	ad   *wkAutoData
}

// freeConnData 释放ConnData
func (w *websocket) freeConnData(wc *wkConnData) {
	w.freeAutoData(wc.ad)
	wc.conn, wc.ad = nil, nil
	w.sender.cdp.Put(wc)
}

// NewAutoData 创建释放器
func (w *websocket) NewRespAutoData(pack *packet.Packet) *wkAutoData {
	ad := w.sender.adp.Get().(*wkAutoData)
	ad.pack, ad.ref = pack, 1
	return ad
}

// AddConnData 添加数据
func (w *websocket) AddRespConnData(conn net.Conn, ad *wkAutoData, idx uint32) uint32 {
	w.sender.RLock()
	wk := w.sender.cdp.Get().(*wkConnData)
	wk.conn, wk.ad = conn, ad
	atomic.AddInt64(&ad.ref, 1)
	if idx >= workerSize {
		idx = atomic.AddUint32(&w.sender.seq, 1) % workerSize
	}
	w.sender.wgo[idx] <- wk
	w.sender.RUnlock()
	return idx
}

// RegisterConn 注册到会话中
func (w *websocket) RegisterConn(c *wConn) (isNewConn bool) {
	if c == nil {
		return
	}
	idx := xutils.HashCode32(c.uid) % chunkSize
	w.session.chunks[idx].Lock()
	if c1, ok := w.session.chunks[idx].m[c.uid]; ok {
		if c1 != c {
			c1.conn.Close()
		}
	} else {
		isNewConn = true
	}
	w.session.chunks[idx].m[c.uid] = c
	w.session.chunks[idx].Unlock()
	return
}

// UnRegisterConn 从会话中注销
func (w *websocket) UnRegisterConn(c *wConn) (isMine bool) {
	if c == nil || c.uid == "" {
		return
	}
	idx := xutils.HashCode32(c.uid) % chunkSize
	w.session.chunks[idx].Lock()
	if c1, ok := w.session.chunks[idx].m[c.uid]; ok && c1 == c {
		isMine = true
		delete(w.session.chunks[idx].m, c.uid)
	}
	w.session.chunks[idx].Unlock()
	return
}

// SendData 发送数据
func (w *websocket) SendData(v interface{}, api string, uis []string) {
	var gzd, nzd *wkAutoData

	if len(uis) > 0 {
		// 按用户发送
		for i := 0; i < chunkSize; i++ {
			w.session.chunks[i].RLock()
			for _, uid := range uis {
				if m, ok := w.session.chunks[i].m[uid]; ok {
					if m.isCompressed {
						if gzd == nil {
							pack := packet.New(2048)
							w.encodingResponseData(pack, api, v, true)
							gzd = w.NewRespAutoData(pack)
						}
						w.AddRespConnData(m.conn, gzd, workerSize)
					} else {
						if nzd == nil {
							pack := packet.New(2048)
							w.encodingResponseData(pack, api, v, false)
							nzd = w.NewRespAutoData(pack)
						}
						w.AddRespConnData(m.conn, nzd, workerSize)
					}
				}
			}
			w.session.chunks[i].RUnlock()
		}
	} else {
		// 全部发送
		for i := 0; i < chunkSize; i++ {
			w.session.chunks[i].RLock()
			for _, m := range w.session.chunks[i].m {
				if m.isCompressed {
					if gzd == nil {
						pack := packet.New(2048)
						w.encodingResponseData(pack, api, v, true)
						gzd = w.NewRespAutoData(pack)
					}
					w.AddRespConnData(m.conn, gzd, workerSize)
				} else {
					if nzd == nil {
						pack := packet.New(2048)
						w.encodingResponseData(pack, api, v, false)
						nzd = w.NewRespAutoData(pack)
					}
					w.AddRespConnData(m.conn, nzd, workerSize)
				}
			}
			w.session.chunks[i].RUnlock()
		}
	}

	// 释放资源
	if gzd != nil {
		w.freeAutoData(gzd)
	}
	if nzd != nil {
		w.freeAutoData(nzd)
	}
}

// SendGroup 按组发送数据
func (w *websocket) SendGroup(v interface{}, api string, flag uint8, group string) {
	var gzd, nzd *wkAutoData

	// 按组发送数据
	for i := 0; i < chunkSize; i++ {
		w.session.chunks[i].RLock()
		for _, m := range w.session.chunks[i].m {
			if !m.group.Match(flag, group) {
				continue
			}
			if m.isCompressed {
				if gzd == nil {
					pack := packet.New(2048)
					w.encodingResponseData(pack, api, v, true)
					gzd = w.NewRespAutoData(pack)
				}
				w.AddRespConnData(m.conn, gzd, workerSize)
			} else {
				if nzd == nil {
					pack := packet.New(2048)
					w.encodingResponseData(pack, api, v, false)
					nzd = w.NewRespAutoData(pack)
				}
				w.AddRespConnData(m.conn, nzd, workerSize)
			}
		}
		w.session.chunks[i].RUnlock()
	}

	// 释放资源
	if gzd != nil {
		w.freeAutoData(gzd)
	}
	if nzd != nil {
		w.freeAutoData(nzd)
	}
}

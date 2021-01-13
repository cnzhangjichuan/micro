package micro

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type rpc struct {
	baseChain

	msgID          uint64
	com            sync.RWMutex
	cos            map[string]net.Conn
	rem            sync.RWMutex
	resp           map[uint64]chan *packet.Packet
	apiWorkerM     sync.RWMutex
	apiWorkerValid bool
	apiWorker      [rpcWorkerSize]chan *rpcPackConn
	dpoPool        sync.Pool
	packConnPool   sync.Pool
}

const (
	rpcCodeRequest  = 11
	rpcCodeResponse = 12
	rpcCodeDataOK   = 21
	rpcCodeDataERR  = 22
	rpcCodeDataNil  = 23
)

var (
	errRPCTimeout = errors.New("rpc call timeout")
)

const rpcWorkerSize = 16

// Init 初始化
func (r *rpc) Init() {
	r.cos = make(map[string]net.Conn, 16)
	r.resp = make(map[uint64]chan *packet.Packet, 16)
	for i := 0; i < rpcWorkerSize; i++ {
		r.apiWorker[i] = make(chan *rpcPackConn, 512)
		go func(c <-chan *rpcPackConn) {
			for p := range c {
				r.handAPIData(p)
				r.freePackConn(p)
			}
		}(r.apiWorker[i])
	}
	r.apiWorkerValid = true
	r.dpoPool.New = func() interface{} {
		return &rpcDpo{}
	}
	r.packConnPool.New = func() interface{} {
		return &rpcPackConn{}
	}
}

// Handle 处理Conn
func (r *rpc) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	if name != "rpc" {
		return false
	}

	// 获取服务名称
	srvName := pack.HTTPHeaderValue(httpAuthorize)
	if srvName == "" {
		return true
	}
	var ok bool
	if srvName, ok = env.authorize.Check(srvName); !ok {
		return true
	}

	// 获取服务地址
	address := pack.HTTPHeaderValue(httpRemoteAddress)
	if address == "" {
		address = conn.RemoteAddr().String()
	}

	// 发送确认信号
	pack.BeginWrite()
	pack.WriteI32(1)
	pack.EndWrite()
	_, err := pack.FlushToConn(conn)
	if err != nil {
		return true
	}

	// 注册连接
	r.com.Lock()
	r.cos[address] = conn
	r.com.Unlock()

	// 处理数据
	r.receive(conn)

	// 注销连接
	r.com.Lock()
	delete(r.cos, address)
	r.com.Unlock()

	return true
}

// Close 关闭
func (r *rpc) Close() {
	r.apiWorkerM.Lock()
	r.apiWorkerValid = false
	for i := 0; i < rpcWorkerSize; i++ {
		close(r.apiWorker[i])
	}
	r.apiWorkerM.Unlock()
}

// Call 远程调用
func (r *rpc) Call(out, in interface{}, adr, api string) error {
	const (
		RT = time.Second * 3
		WT = time.Second * 3
	)

	// 获取连接
	conn, err := r.GetConn(adr)
	if err != nil {
		return err
	}

	// 组装数据
	pack := packet.New(1024)
	pack.SetTimeout(RT, WT)
	pack.BeginWrite()
	msgID := atomic.AddUint64(&r.msgID, 1)
	pack.WriteU32(rpcCodeRequest)
	pack.WriteU64(msgID)
	pack.WriteString(api)
	if i, ok := in.(packet.Encoder); ok {
		i.Encode(pack)
	} else {
		pack.EncodeJSON(in, true, true)
	}
	pack.EndWrite()

	// 注册接收器
	resp := make(chan *packet.Packet, 1)
	r.rem.Lock()
	r.resp[msgID] = resp
	r.rem.Unlock()

	// 发送数据
	_, err = pack.FlushToConn(conn)
	packet.Free(pack)

	// 接收数据
	if err == nil {
		t := time.NewTimer(RT)
		select {
		case rsp := <-resp:
			t.Stop()
			switch rsp.ReadU32() {
			case rpcCodeDataOK:
				if out != nil {
					if d, ok := out.(packet.Decoder); ok {
						d.Decode(rsp)
					} else {
						err = rsp.DecodeJSON(out)
					}
				}
			case rpcCodeDataERR:
				err = errors.New(rsp.ReadString())
			case rpcCodeDataNil:
			}
			packet.Free(rsp)
		case <-t.C:
			err = errRPCTimeout
		}
	}

	// 清理资源
	r.rem.Lock()
	delete(r.resp, msgID)
	close(resp)
	r.rem.Unlock()

	return err
}

// GetConn 获取连接
func (r *rpc) GetConn(adr string) (conn net.Conn, err error) {
	const TIMEOUT = time.Second * 3

	ok := false
	r.com.RLock()
	conn, ok = r.cos[adr]
	r.com.RUnlock()
	if ok {
		return
	}

	// 创建连接
	r.com.Lock()
	conn, ok = r.cos[adr]
	if ok {
		r.com.Unlock()
		return
	}
	conn, err = net.DialTimeout("tcp", adr, TIMEOUT)
	if err != nil {
		r.com.Unlock()
		return
	}

	// 建立通道
	pack := packet.New(512)
	pack.SetTimeout(TIMEOUT, TIMEOUT)
	pack.Write(httpRPCUpgrade)
	pack.Write(httpRowAt)
	pack.Write(httpAuthorize)
	pack.Write(xutils.UnsafeStringToBytes(env.authorize.NewCode(env.config.Name)))
	pack.Write(httpRowAt)
	pack.Write(httpRowAt)
	_, err = pack.FlushToConn(conn)
	if err != nil {
		r.com.Unlock()
		packet.Free(pack)
		conn.Close()
		return
	}

	// 获取连接状态
	err = pack.ReadConn(conn)
	if err != nil || pack.ReadI32() != 1 {
		r.com.Unlock()
		packet.Free(pack)
		conn.Close()
		err = errRPCTimeout
		return
	}

	// 缓存本次连接
	r.cos[adr] = conn
	r.com.Unlock()

	// 接收数据
	go func(conn net.Conn, adr string) {
		defer func() {
			r.com.Lock()
			delete(r.cos, adr)
			r.com.Unlock()
			recover()
		}()
		r.receive(conn)
	}(conn, adr)

	return
}

// receive 接收数据
func (r *rpc) receive(conn net.Conn) {
	const (
		RT = time.Minute
		WT = time.Second * 3
	)

	pack := packet.New(1024)
	pack.SetTimeout(RT, WT)
	for r.apiWorkerValid {
		code, err := pack.ReadConnWithKeepAlive(conn)
		if err != nil {
			break
		}
		switch code {
		case rpcCodeResponse:
			// response
			r.rem.RLock()
			msgID := pack.ReadU64()
			if c, ok := r.resp[msgID]; ok {
				select {
				case c <- pack.Copy():
				default:
				}
			}
			r.rem.RUnlock()
		case rpcCodeRequest:
			// request
			r.apiWorkerM.RLock()
			if r.apiWorkerValid {
				pc := r.createPackConn()
				pc.pack = pack.Copy()
				pc.pack.SetTimeout(RT, WT)
				pc.conn = conn
				select {
				case r.apiWorker[0] <- pc:
				case r.apiWorker[1] <- pc:
				case r.apiWorker[2] <- pc:
				case r.apiWorker[3] <- pc:
				case r.apiWorker[4] <- pc:
				case r.apiWorker[5] <- pc:
				case r.apiWorker[6] <- pc:
				case r.apiWorker[7] <- pc:
				case r.apiWorker[8] <- pc:
				case r.apiWorker[9] <- pc:
				case r.apiWorker[10] <- pc:
				case r.apiWorker[11] <- pc:
				case r.apiWorker[12] <- pc:
				case r.apiWorker[13] <- pc:
				case r.apiWorker[14] <- pc:
				case r.apiWorker[15] <- pc:
				}
			}
			r.apiWorkerM.RUnlock()
		}
	}
	packet.Free(pack)
}

// handAPIData 处理数据
func (r *rpc) handAPIData(pc *rpcPackConn) {
	var (
		resp    interface{}
		errCode string
		msgID   = pc.pack.ReadU64()
		api     = pc.pack.ReadString()
	)

	f, ok := findBis(api)
	if !ok {
		errCode = apiNotFoundError.ErrCode
	} else {
		dpo := r.createDpo()
		dpo.pack = pc.pack
		resp, errCode = f(dpo)
		r.freeDpo(dpo)
	}

	// response
	pc.pack.BeginWrite()
	pc.pack.WriteU32(rpcCodeResponse)
	pc.pack.WriteU64(msgID)
	if errCode != "" {
		// 发送错误码
		pc.pack.WriteU32(rpcCodeDataERR)
		pc.pack.WriteString(errCode)
	} else if resp != nil {
		// 发送数据
		pc.pack.WriteU32(rpcCodeDataOK)
		if e, ok := resp.(packet.Encoder); ok {
			e.Encode(pc.pack)
		} else {
			pc.pack.EncodeJSON(resp, true, true)
		}
	} else {
		// 空响应
		pc.pack.WriteU32(rpcCodeDataNil)
	}
	pc.pack.EndWrite()
	// 发送数据
	if pc.conn != nil {
		pc.pack.FlushToConn(pc.conn)
	}
}

type rpcPackConn struct {
	conn net.Conn
	pack *packet.Packet
}

// createPackConn 创建连接包
func (r *rpc) createPackConn() *rpcPackConn {
	return r.packConnPool.Get().(*rpcPackConn)
}

// freePackConn 释放连接包
func (r *rpc) freePackConn(p *rpcPackConn) {
	if p == nil {
		return
	}
	packet.Free(p.pack)
	p.conn, p.pack = nil, nil
	r.packConnPool.Put(p)
}

type rpcDpo struct {
	baseDpo

	pack *packet.Packet
}

// Parse 获取客户端参数
func (r *rpcDpo) Parse(v interface{}) {
	if r.pack == nil {
		return
	}
	if d, ok := v.(packet.Decoder); ok {
		d.Decode(r.pack)
	} else {
		err := r.pack.DecodeJSON(v)
		if err != nil {
			Debug("rpc dpo parse data error: %v", err)
		}
	}
}

// createDpo 创建处理对象
func (r *rpc) createDpo() *rpcDpo {
	return r.dpoPool.Get().(*rpcDpo)
}

// freeDpo 释放处理对象
func (r *rpc) freeDpo(dpo *rpcDpo) {
	if dpo == nil {
		return
	}
	dpo.pack = nil
	dpo.release()
	r.dpoPool.Put(dpo)
}

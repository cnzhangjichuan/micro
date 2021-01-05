package micro

import (
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

const (
	registryBisInit   = 11
	registryBisAdd    = 12
	registryBisRemove = 13
)

// registry 注册表
type registry struct {
	baseChain
	sync.RWMutex

	running bool
	client  net.Conn
	remotes []net.Conn

	// 地址映射表
	addresses map[string]*addr
}

type addr struct {
	found uint32
	ads   []string
}

// Init 初始化
func (r *registry) Init() {
	r.running = true
	r.remotes = make([]net.Conn, 0, 16)
	r.addresses = make(map[string]*addr, 16)
	if env.config.Registry != "" {
		go r.Register(env.config.Registry)
	}
}

// Handle 处理Conn
func (r *registry) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	const TIMEOUT = time.Second * 10

	if name != "registry" {
		return false
	}
	if !r.running {
		return true
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
	hps := strings.Index(address, ":")
	if hps >= 0 {
		port := pack.HTTPHeaderValue(httpRegistryPort)
		address = address[:hps+1] + port
	}

	// 清除发送缓冲区
	pack.ReadHTTPBody(conn)

	// 设置同步超时时长
	pack.SetTimeout(TIMEOUT, TIMEOUT)

	// 广播加入事件
	r.Broadcast(pack, registryBisAdd, conn, srvName, address)

	// 保活连接
	pack.ReadConnWithKeepAlive(conn)

	// 广播离开事件
	r.Broadcast(pack, registryBisRemove, conn, srvName, address)

	return true
}

// 获取服务地址
func (r *registry) ServerAddress(name string) string {
	r.RLock()
	as, ok := r.addresses[name]
	if !ok {
		r.RUnlock()
		// 如果没有找到注册的服务
		// 使用注册服的地址
		return env.config.Registry
	}
	l := len(as.ads)
	if l == 0 {
		r.RUnlock()
		return ""
	}
	found := atomic.AddUint32(&as.found, 1)
	adr := as.ads[found%uint32(l)]
	r.RUnlock()

	return adr
}

// Close 关闭
func (r *registry) Close() {
	r.running = false
	if r.client != nil {
		r.client.Close()
		r.client = nil
	}

	r.Lock()
	if len(r.remotes) > 0 {
		for i := 0; i < len(r.remotes); i++ {
			r.remotes[i].Close()
		}
		r.remotes = r.remotes[:0]
	}
	r.Unlock()
}

// Register 注册到指定位置
func (r *registry) Register(remote string) {
	const (
		TIMEOUT  = time.Second * 30
		ERRDELAY = time.Second * 3
	)

	var err error
	for r.running {
		r.client, err = net.DialTimeout("tcp", remote, TIMEOUT)
		if err != nil {
			time.Sleep(ERRDELAY)
			continue
		}
		conn := r.client

		// 连接协商
		pack := packet.New(512)
		pack.SetTimeout(TIMEOUT, TIMEOUT)
		pack.Write(httpRegistryUpgrade)
		pack.Write(httpRowAt)
		// 服务名称
		pack.Write(httpAuthorize)
		pack.Write(xutils.UnsafeStringToBytes(env.authorize.NewCode(env.config.Name)))
		pack.Write(httpRowAt)
		// 服务端口
		pack.Write(httpRegistryPort)
		port := env.config.Address
		idx := strings.Index(port, ":")
		if idx >= 0 {
			port = port[idx+1:]
		}
		pack.Write(xutils.UnsafeStringToBytes(port))
		pack.Write(httpRowAt)
		pack.Write(httpRowAt)

		if _, err = pack.FlushToConn(conn); err != nil {
			Debug("register handshake error %v", err)
			packet.Free(pack)
			conn.Close()
			time.Sleep(ERRDELAY)
			continue
		}

		// 交换数据
		var msgCode uint32
		for {
			msgCode, err = pack.ReadConnWithKeepAlive(conn)
			if err != nil {
				Debug("register read state error %v", err)
				break
			}
			switch msgCode {
			case registryBisInit:
				// 初始化
				r.FillSet(pack)
			case registryBisAdd:
				// 添加name=address
				r.Add(pack)
			case registryBisRemove:
				// 移除name-address
				r.Remove(pack)
			}
		}
		packet.Free(pack)
		conn.Close()
		time.Sleep(ERRDELAY)
	}
}

// addRemote 添加remote
func (r *registry) addRemote(conn net.Conn) {
	r.remotes = append(r.remotes, conn)
}

// removeRemote 移除remote
func (r *registry) removeRemote(conn net.Conn) {
	for i := 0; i < len(r.remotes); i++ {
		if r.remotes[i] == conn {
			copy(r.remotes[i:], r.remotes[i+1:])
			r.remotes = r.remotes[:len(r.remotes)-1]
			break
		}
	}
}

// Broadcast 广播事件
func (r *registry) Broadcast(pack *packet.Packet, event uint32, conn net.Conn, name, address string) {
	r.Lock()

	// 加入/移除注册表信息
	if r.addOrRemove(name, address, event) {
		r.removeRemote(conn)
	} else {
		pack.BeginWrite()
		pack.WriteU32(registryBisInit)
		r.encode(pack)
		pack.EndWrite()
		pack.FlushToConn(conn)
	}

	// 广播事件
	pack.BeginWrite()
	pack.WriteU32(event)
	pack.WriteString(name)
	pack.WriteString(address)
	pack.EndWrite()
	for i := 0; i < len(r.remotes); i++ {
		pack.FlushToConn(r.remotes[i])
	}

	// 注册连接
	switch event {
	case registryBisAdd:
		r.addRemote(conn)
	}

	r.Unlock()
}

// Encode 序列化
func (r *registry) encode(pack *packet.Packet) {
	s := int(0)
	for _, as := range r.addresses {
		s += len(as.ads)
	}
	pack.WriteU32(uint32(s))
	for name, as := range r.addresses {
		for _, a := range as.ads {
			pack.WriteString(name)
			pack.WriteString(a)
		}
	}
}

// 填充注册表
func (r *registry) FillSet(pack *packet.Packet) {
	r.Lock()

	// clear
	for k := range r.addresses {
		delete(r.addresses, k)
	}

	// fill
	s := pack.ReadU32()
	for i := uint32(0); i < s; i++ {
		name := pack.ReadString()
		address := pack.ReadString()
		r.add(name, address)
	}
	r.Unlock()
}

// Remove 移除name-address
func (r *registry) Remove(pack *packet.Packet) {
	name := pack.ReadString()
	address := pack.ReadString()
	r.Lock()
	r.remove(name, address)
	r.Unlock()
}

// remove 移除name-address
func (r *registry) remove(name, address string) {
	as, ok := r.addresses[name]
	if !ok {
		return
	}
	as.ads = xutils.RemoveSS(as.ads, address)
}

// Add 添加name-address
func (r *registry) Add(pack *packet.Packet) {
	name := pack.ReadString()
	address := pack.ReadString()
	r.Lock()
	r.add(name, address)
	r.Unlock()
}

// add 添加name-address
func (r *registry) add(name, address string) {
	if name == env.config.Name {
		return
	}
	as, ok := r.addresses[name]
	if !ok {
		as = &addr{
			ads: make([]string, 0, 16),
		}
		r.addresses[name] = as
	}
	as.ads = xutils.AddNoRepeatItem(as.ads, address)
}

// addOrRemove 添加/移除注册表信息
func (r *registry) addOrRemove(name, address string, event uint32) (isRemoved bool) {
	switch event {
	case registryBisAdd:
		r.add(name, address)
	case registryBisRemove:
		isRemoved = true
		r.remove(name, address)
	}
	return
}

package xrpc

import (
	"errors"
	"github.com/cnzhangjichuan/micro/xutils"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

func Load(result interface{}, id, api string, request interface{}) error {
	var (
		c        *client
		err      error
		isNetErr bool
	)
	for i := 0; i < 2; i++ {
		c, err = env.csg.FindClient(id)
		if err != nil {
			return err
		}
		err, isNetErr = c.LoadData(result, api, request)
		if isNetErr {
			continue
		}
	}
	return err
}

func LoadByAddress(result interface{}, id, address, api string, request interface{}) error {
	var (
		c        *client
		err      error
		isNetErr bool
	)
	for i := 0; i < 2; i++ {
		c, err = env.csg.FindClientByAddress(id, address)
		if err != nil {
			return err
		}
		err, isNetErr = c.LoadData(result, api, request)
		if isNetErr {
			continue
		}
	}
	return err
}

type clientsGroup struct {
	sync.RWMutex
	climap        map[string]*clients
	connCacheSize int
}

func (c *clientsGroup) FindClient(id string) (cli *client, err error) {
	c.RLock()
	if cs, ok := c.climap[id]; ok {
		cli, err = cs.Find()
	} else {
		err = env.aut.errServiceNotFound
	}
	c.RUnlock()
	return
}

func (c *clientsGroup) FindClientByAddress(id, address string) (cli *client, err error) {
	c.RLock()
	if cs, ok := c.climap[id]; ok {
		cli, err = cs.FindByAddress(address)
	} else {
		err = env.aut.errServiceNotFound
	}
	c.RUnlock()
	return
}

func (c *clientsGroup) SyncAllMicroService(packet *xutils.Packet) {
	count := packet.ReadI64()
	if count == 0 {
		return
	}
	c.Lock()
	exists := make(map[string]string)
	// add new
	for i := int64(0); i < count; i++ {
		id := packet.ReadString()
		adr := packet.ReadString()
		logInfo("\tservice id: %s, address: %s", id, adr)
		if s, ok := c.climap[id]; ok {
			s.Add(adr)
		} else {
			s = &clients{
				idx: -1,
			}
			s.Add(adr)
			c.climap[id] = s
		}
		exists[adr] = id
	}
	// close not in list micro
	for id, s := range c.climap {
		for _, c := range s.cs {
			if eid, ok := exists[c.adr]; !ok || eid != id {
				c.SetEnable(false)
			}
		}
	}
	c.Unlock()
}

func (c *clientsGroup) AddMicroService(id, address string) {
	if id != env.id {
		c.Lock()
		if s, ok := c.climap[id]; ok {
			s.Add(address)
		} else {
			s = &clients{
				idx: -1,
			}
			s.Add(address)
			c.climap[id] = s
		}
		c.Unlock()
	}
}

func (c *clientsGroup) DelMicroService(id, address string) {
	if env.id != id {
		c.Lock()
		if s, ok := c.climap[id]; ok {
			s.Del(address)
		}
		c.Unlock()
	}
}

// ********************************************************************************************
// clients
type clients struct {
	sync.RWMutex
	cs  []*client
	idx int32
}

func (s *clients) FindByAddress(address string) (c *client, err error) {
	s.RLock()
	csl := len(s.cs)
	if csl == 0 {
		s.RUnlock()
		err = env.aut.errServiceNotFound
		return
	}
	for l := 0; l < csl; l++ {
		c = s.cs[l]
		if c.Equals(address) {
			if c.Enable() {
				s.RUnlock()
				return
			}
			break
		}
	}
	s.RUnlock()
	c = nil
	err = env.aut.errServiceNotFound
	return
}

func (s *clients) Find() (c *client, err error) {
	s.RLock()
	csl := len(s.cs)
	if csl == 0 {
		s.RUnlock()
		err = env.aut.errServiceNotFound
		return
	}

	for l := 0; l < csl; l++ {
		i := atomic.AddInt32(&s.idx, 1)
		if int(i) >= csl {
			i = 0
			atomic.StoreInt32(&s.idx, i)
		}
		c = s.cs[i]
		if c.Enable() {
			s.RUnlock()
			return
		}
	}
	s.RUnlock()
	c = nil
	err = env.aut.errServiceNotFound
	return
}

func (s *clients) Add(address string) {
	s.Lock()
	for _, c := range s.cs {
		if c.Equals(address) {
			c.SetEnable(true)
			s.Unlock()
			return
		}
	}
	s.cs = append(s.cs, s.NewClient(address))
	s.Unlock()
}

func (s *clients) Del(address string) {
	s.Lock()
	csl := len(s.cs)
	for i := 0; i < csl; i++ {
		if s.cs[i].Equals(address) {
			s.cs[i].SetEnable(false)
			break
		}
	}
	s.Unlock()
}

// *******************************************************************************************************
// new client
func (s *clients) NewClient(address string) *client {
	const (
		defaultConnCacheSize = 16
		maxBuffSize          = 1024
	)

	var c client
	c.adr = address
	c.enb = 1
	ccs := env.csg.connCacheSize
	if ccs <= 0 {
		ccs = defaultConnCacheSize
	}
	c.ccs = make(chan *con, ccs)
	for i := 0; i < ccs; i++ {
		c.ccs <- &con{
			conn:   nil,
			packet: xutils.NewPacket(maxBuffSize),
			used:   0,
		}
	}
	c.hsk = "GET / HTTP/1.1\r\nHost: " + address + "\r\nUpgrade: rpc\r\n\r\n"

	return &c
}

type client struct {
	adr string
	hsk string
	ccs chan *con
	enb int32
}

func (c *client) LoadData(result interface{}, api string, request interface{}) (err error, isNetErr bool) {
	cn, err := c.Get()
	if err != nil {
		return
	}

	// send data
	dat, err := xutils.MarshalJson(request)
	if err != nil {
		c.Put(cn)
		return
	}
	err = cn.Write(api, dat)
	if err != nil {
		// retry write
		err = c.BindConn(cn)
		if err == nil {
			err = cn.Write(api, dat)
		}
		if err != nil {
			c.Put(cn)
			isNetErr = true
			return
		}
	}

	// get response
	p, err := cn.Read()
	if err != nil {
		c.Put(cn)
		return
	}
	f, _ := p.ReadByte()
	switch f {
	default:
		err = xutils.UnmarshalJson(p.Data(), result)
	case xPREFIX_ERR:
		err = errors.New(p.ReadString())
	}
	c.Put(cn)

	return
}

func (c *client) SetEnable(enable bool) {
	if enable {
		atomic.StoreInt32(&c.enb, 1)
	} else {
		atomic.StoreInt32(&c.enb, 0)
		running := true
		for i := 0; running && i < cap(c.ccs); i++ {
			select {
			default:
				running = false
			case cn := <-c.ccs:
				cn.SetConn(nil)
				select {
				default:
				case c.ccs <- cn:
				}
			}
		}
	}
}

func (c *client) Enable() bool {
	return atomic.LoadInt32(&c.enb) == 1
}

func (c *client) Equals(address string) bool {
	return c.adr == address
}

func (c *client) Put(cn *con) {
	if !c.Enable() {
		cn.SetConn(nil)
	}
	cn.Unused()
	select {
	default:
		cn.SetConn(nil)
	case c.ccs <- cn:
	}
}

func (c *client) Get() (cn *con, err error) {
	if !c.Enable() {
		err = env.aut.errServiceClosed
		return
	}

	// get from ccs
	cn = <-c.ccs

	if cn.HasConn() {
		return
	}

	err = c.BindConn(cn)
	if err != nil {
		c.Put(cn)
	}

	return
}

func (c *client) BindConn(cn *con) (err error) {
	// create new conn
	conn, err := net.Dial("tcp", c.adr)
	if err != nil {
		return
	}
	_, err = conn.Write(xutils.UnsafeStringToBytes(c.hsk))
	if err != nil {
		return
	}

	// bind conn
	cn.SetConn(conn)

	// read codes from server
	p, err := cn.Read()
	if err != nil {
		return
	}

	// send encoded codes to server
	err = cn.Write("", env.aut.Encoding(p.Data()))
	if err != nil {
		return
	}

	// is server return ok?
	p, err = cn.Read()
	if err != nil {
		return
	}
	f, _ := p.ReadByte()
	switch f {
	default:
		if p.ReadString() != `OK` {
			err = env.aut.errServiceAuthFailed
			cn.SetConn(nil)
		}

	case xPREFIX_ERR:
		err = errors.New(p.ReadString())
		cn.SetConn(nil)

	}
	return
}

type con struct {
	conn   net.Conn
	packet *xutils.Packet
	used   int32
}

func (c *con) Used() bool {
	return atomic.CompareAndSwapInt32(&c.used, 0, 1)
}

func (c *con) Unused() {
	atomic.StoreInt32(&c.used, 0)
}

func (c *con) HasConn() bool {
	return c.conn != nil
}

func (c *con) SetConn(conn net.Conn) {
	if c.conn != nil && c.conn != conn {
		c.conn.Close()
	}
	c.conn = conn
}

// send data
func (c *con) Write(uri string, data []byte) error {
	const TOM = time.Second * 3

	// setter data
	c.packet.BeginWrite()
	if uri != "" {
		c.packet.WriteString(uri)
	}
	c.packet.Write(data)
	c.packet.EndWrite()

	// send data
	c.conn.SetWriteDeadline(time.Now().Add(TOM))
	_, err := c.packet.FlushToConn(c.conn)
	if err != nil {
		c.conn.Close()
		c.conn = nil
	}
	return err
}

// read data
func (c *con) Read() (p *xutils.Packet, err error) {
	const TOM = time.Second * 6

	c.conn.SetReadDeadline(time.Now().Add(TOM))
	err = c.packet.ReadConn(c.conn)
	if err != nil {
		c.conn.Close()
		c.conn = nil
	}
	p = c.packet
	return
}

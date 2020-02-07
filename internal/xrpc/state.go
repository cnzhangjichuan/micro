package xrpc

import (
	"github.com/cnzhangjichuan/micro/xutils"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	start     = `Z`
	syncMicro = `S`
	addMicro  = `A`
	delMicro  = `D`
	ping      = `Q`
	pong      = `R`
)

func HandleState(w http.ResponseWriter, r *http.Request) {
	const (
		TOM = time.Minute * 1
		ROM = time.Second * 1
	)

	host := r.Header.Get("X-Real-IP")
	if host == "" {
		host = r.RemoteAddr
	}
	host = strings.Split(host, ":")[0]
	conn, _, err := w.(http.Hijacker).Hijack()

	if err != nil {
		logError("rpc state hijacking [%s] error: %v", r.RemoteAddr, err)
		return
	}
	defer conn.Close()

	packet := xutils.NewPacket(1024)

	// request remote info
	packet.BeginWrite()
	packet.WriteString(start)
	packet.EndWrite()
	conn.SetWriteDeadline(time.Now().Add(ROM))
	_, err = packet.FlushToConn(conn)
	if err != nil {
		logError("rpc state request remote info error: %v", err)
		return
	}

	// register remote
	conn.SetReadDeadline(time.Now().Add(ROM))
	err = packet.ReadConn(conn)
	if err != nil {
		logError("rpc state get remote info error: %v", err)
		return
	}
	id := packet.ReadString()
	port := packet.ReadString()
	adr := host + ":" + port

	// sync micro address
	packet.BeginWrite()
	env.sia.Fill(packet)
	packet.EndWrite()
	conn.SetWriteDeadline(time.Now().Add(ROM))
	_, err = packet.FlushToConn(conn)
	if err != nil {
		logError("rpc state response all micro info error: %v", err)
		return
	}

	env.sia.Add(adr, id)
	env.csg.AddMicroService(id, adr)

	// broadcast
	packet.BeginWrite()
	packet.WriteString(addMicro)
	packet.WriteString(id)
	packet.WriteString(adr)
	packet.EndWrite()
	env.scp.Broadcast(packet.Data())

	// add notice list
	env.scp.RegisterStaConn(conn)

	// keep-alive
	willClose := false
	for {
		conn.SetReadDeadline(time.Now().Add(TOM))
		err = packet.ReadConn(conn)
		if err != nil {
			if willClose {
				break
			}
			if e, ok := err.(net.Error); ok && e.Timeout() {
				packet.BeginWrite()
				packet.WriteString(ping)
				packet.EndWrite()
				conn.SetWriteDeadline(time.Now().Add(ROM))
				packet.FlushToConn(conn)
				willClose = true
				continue
			}
			break
		}
		switch packet.ReadString() {
		case ping:
			packet.BeginWrite()
			packet.WriteString(pong)
			packet.EndWrite()
			conn.SetWriteDeadline(time.Now().Add(ROM))
			packet.FlushToConn(conn)
		case pong:
			willClose = false
		}
	}

	// remove conn
	env.scp.UnregisterStaConn(conn)

	// remove this
	env.sia.Del(adr)
	env.csg.DelMicroService(id, adr)

	packet.BeginWrite()
	packet.WriteString(delMicro)
	packet.WriteString(id)
	packet.WriteString(adr)
	packet.EndWrite()
	env.scp.Broadcast(packet.Data())
}

func syncState(address string) {
	const retry = time.Second * 3

	hsd := []byte("GET / HTTP/1.1\r\nHost: " + address + "\r\nUpgrade: sta\r\n\r\n")

	f := func(hsd []byte) error {
		const (
			TOM = time.Minute * 1
			WOM = time.Second * 1
		)

		conn, err := net.Dial("tcp", address)
		if err != nil {
			return err
		}
		_, err = conn.Write(hsd)
		if err != nil {
			conn.Close()
			return err
		}

		packet := xutils.NewPacket(1024)

		// read ok?
		conn.SetReadDeadline(time.Now().Add(WOM))
		err = packet.ReadConn(conn)
		if err != nil || packet.ReadString() != start {
			conn.Close()
			return err
		}

		// register
		packet.BeginWrite()
		packet.WriteString(env.id)
		packet.WriteString(env.port)
		packet.EndWrite()
		conn.SetWriteDeadline(time.Now().Add(WOM))
		_, err = packet.FlushToConn(conn)
		if err != nil {
			conn.Close()
			return err
		}

		// sync
		willClose := false
		for {
			conn.SetReadDeadline(time.Now().Add(TOM))
			err = packet.ReadConn(conn)
			if err != nil {
				if willClose {
					break
				}
				if e, ok := err.(net.Error); ok && e.Timeout() {
					packet.BeginWrite()
					packet.WriteString(ping)
					packet.EndWrite()
					conn.SetWriteDeadline(time.Now().Add(WOM))
					packet.FlushToConn(conn)
					willClose = true
					continue
				}
				break
			}
			switch packet.ReadString() {
			case ping:
				packet.BeginWrite()
				packet.WriteString(pong)
				packet.EndWrite()
				conn.SetWriteDeadline(time.Now().Add(WOM))
				packet.FlushToConn(conn)
			case pong:
				willClose = false
			case syncMicro:
				env.csg.SyncAllMicroService(packet)

			case addMicro:
				id := packet.ReadString()
				adr := packet.ReadString()
				env.csg.AddMicroService(id, adr)

			case delMicro:
				id := packet.ReadString()
				adr := packet.ReadString()
				env.csg.DelMicroService(id, adr)

			}
		}
		conn.Close()
		return err
	}

	for {
		err := f(hsd)
		if err != nil {
			logError("sync state error: %v", err)
		}
		time.Sleep(retry)
	}
}

// state id & address map
type staIdAddress struct {
	sync.RWMutex
	stamap map[string]string
}

func (s *staIdAddress) Fill(packet *xutils.Packet) {
	s.RLock()
	packet.WriteString(syncMicro)
	packet.WriteI64(int64(len(s.stamap)))
	for adr, id := range s.stamap {
		packet.WriteString(id)
		packet.WriteString(adr)
	}
	s.RUnlock()
}

func (s *staIdAddress) Add(address, id string) {
	s.Lock()
	s.stamap[address] = id
	s.Unlock()
}

func (s *staIdAddress) Del(address string) {
	s.Lock()
	delete(s.stamap, address)
	s.Unlock()
}

// state connection pool
type staConnPool struct {
	sync.RWMutex
	cons []net.Conn
}

func (s *staConnPool) Broadcast(data []byte) {
	const TOM = time.Second * 1
	at := time.Now().Add(TOM)
	s.RLock()
	for _, c := range s.cons {
		c.SetWriteDeadline(at)
		c.Write(data)
	}
	s.RUnlock()
}

func (s *staConnPool) RegisterStaConn(c net.Conn) {
	s.Lock()
	s.cons = append(s.cons, c)
	s.Unlock()
}

func (s *staConnPool) UnregisterStaConn(c net.Conn) {
	s.Lock()
	csl := len(s.cons)
	for i := 0; i < csl; i++ {
		if s.cons[i] == c {
			for j := i + 1; j < csl; j++ {
				s.cons[j-1] = s.cons[j]
			}
			s.cons[csl-1] = nil
			s.cons = s.cons[:csl-1]
			break
		}
	}
	s.Unlock()
}

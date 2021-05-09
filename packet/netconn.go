package packet

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"time"
)

// FlushToConn 将数据发送到网络连接上
func (p *Packet) FlushToConn(conn net.Conn) (n int, err error) {
	if p.r >= p.w {
		return
	}
	if p.wt > 0 {
		err = conn.SetWriteDeadline(time.Now().Add(p.wt))
		if err != nil {
			return
		}
	}
	n, err = conn.Write(p.buf[p.r:p.w])
	return
}

// ReadConn 从网络连接中读取数据
func (p *Packet) ReadConn(conn net.Conn) (err error) {
	if p.rt > 0 {
		err = conn.SetReadDeadline(time.Now().Add(p.rt))
		if err != nil {
			return
		}
	}
	p.Reset()
	hd := p.Allocate(4)
	_, err = io.ReadFull(conn, hd)
	if err != nil {
		return
	}
	s := binary.LittleEndian.Uint32(hd)
	p.Reset()
	_, err = io.ReadFull(conn, p.Allocate(int(s)))
	return
}

// BeginWrite 开始写入数据
func (p *Packet) BeginWrite() {
	p.r = 0
	p.w = 0
	p.Allocate(4)
}

// EndWrite 结束写入数据
func (p *Packet) EndWrite() {
	binary.LittleEndian.PutUint32(p.buf[:4], uint32(p.w-4))
}

// ReadConnWithKeepAlive 从网络连接中读取数据，并维持心跳
func (p *Packet) ReadConnWithKeepAlive(conn net.Conn) (msgCode uint32, err error) {
	const (
		PING = 1
		PONG = 2
	)

	if p.rt <= 0 {
		p.rt = time.Second * 3
	}

	rt := p.rt
	err = p.ReadConn(conn)
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			p.BeginWrite()
			p.WriteU32(PING)
			p.EndWrite()
			if _, err = p.FlushToConn(conn); err != nil {
				return
			}
			p.rt = time.Second
			err = p.ReadConn(conn)
			p.rt = rt
			if err != nil {
				return
			}
		} else {
			return
		}
	}

	msgCode = p.ReadU32()
	switch msgCode {
	case PING:
		p.BeginWrite()
		p.WriteU32(PONG)
		p.EndWrite()
		if _, err = p.FlushToConn(conn); err != nil {
			return
		}
		return p.ReadConnWithKeepAlive(conn)
	case PONG:
		return p.ReadConnWithKeepAlive(conn)
	}

	return
}

var (
	httpHeadEndAt         = []byte{'\r', '\n', '\r', '\n'}
	httpHeadRowAt         = []byte{'\r', '\n'}
	httpHeadContentLength = []byte("Content-Length: ")
)

// ReadHTTPHeader 从网络中读取http格式的数据(头域)
func (p *Packet) ReadHTTPHeader(conn net.Conn) (err error) {
	if p.rt > 0 {
		err = conn.SetReadDeadline(time.Now().Add(p.rt))
		if err != nil {
			return
		}
	}
	p.Reset()
	var n, s int
	for {
		s = p.w
		buf := p.Allocate(1024)
		n, err = conn.Read(buf)
		if err != nil {
			return
		}
		p.w = s + n
		if bytes.Index(p.buf[0:p.w], httpHeadEndAt) >= 0 {
			return
		}
	}
}

func (p *Packet) sax(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// Unescape 解码URL参数
func (p *Packet) Unescape(s []byte) int {
	var (
		cx = 0
		px = 0
		sx = len(s)
	)

	for px < sx {
		switch s[px] {
		case '%':
			s[cx] = p.sax(s[px+1])<<4 | p.sax(s[px+2])
			px += 3
			cx++
		default:
			s[cx] = s[px]
			cx++
			px++
		}
	}
	return cx
}

// ReadHTTPStream 从网络中读取http格式的数据(body)
func (p *Packet) ReadHTTPStream(conn net.Conn) (err error) {
	var bodySize int
	if i, er := strconv.Atoi(p.HTTPHeaderValue(httpHeadContentLength)); er == nil {
		bodySize = i
	} else {
		bodySize = 0
	}

	if i := bytes.Index(p.buf[0:p.w], httpHeadEndAt); i < 0 {
		p.Reset()
	} else {
		i += len(httpHeadEndAt)
		copy(p.buf, p.buf[i:p.w])
		p.r = 0
		p.w = p.w - i
	}

	if bodySize <= 0 {
		return
	}

	if p.w >= bodySize {
		return
	}

	if p.rt > 0 {
		err = conn.SetReadDeadline(time.Now().Add(p.rt))
		if err != nil {
			return
		}
	}
	_, err = io.ReadFull(conn, p.Allocate(bodySize-p.w))
	return err
}

// ReadHTTPBody 从网络中读取http格式的数据(body)
func (p *Packet) ReadHTTPBody(conn net.Conn) (err error) {
	err = p.ReadHTTPStream(conn)
	if err == nil {
		p.w = p.Unescape(p.buf[:p.w])
	}
	return err
}

// HTTPHeaderValue http头域中的值
func (p *Packet) HTTPHeaderValue(key []byte) string {
	return string(bytes.TrimSpace(p.DataBetween(key, httpHeadRowAt)))
}
package packet

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var packetPool = sync.Pool{
	New: func() interface{} {
		return &Packet{}
	},
}

// New 创建数据包
func New(capacity int) *Packet {
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	pack.r = 0
	pack.w = 0
	pack.rt = 0
	pack.wt = 0
	data := getBytes(capacity)
	pack.buf = data[:cap(data)]
	return pack
}

// NewWithData 通过已有数据创建数据包
func NewWithData(data []byte) *Packet {
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	pack.r = 0
	pack.w = len(data)
	pack.rt = 0
	pack.wt = 0
	pack.buf = data[:cap(data)]
	return pack
}

// Free 释放数据包
func Free(pack *Packet) {
	if pack == nil {
		return
	}
	if !atomic.CompareAndSwapUint32(&pack.freed, 0, 1) {
		return
	}
	if pack.buf != nil {
		putBytes(pack.buf)
		pack.buf = nil
	}
	packetPool.Put(pack)
}

// Packet 数据包
type Packet struct {
	freed uint32
	buf   []byte
	r     int
	w     int
	rt    time.Duration
	wt    time.Duration
}

// SetTimeout 设置读/写超时时长
func (p *Packet) SetTimeout(readTimeout, writeTimeout time.Duration) {
	p.rt, p.wt = readTimeout, writeTimeout
}

// GetTimeout 获取超时设置
func (p *Packet) GetTimeout() (rt, wt time.Duration) {
	rt, wt = p.rt, p.wt
	return
}

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

// ReadHTTPBody 从网络中读取http格式的数据(body)
func (p *Packet) ReadHTTPBody(conn net.Conn) (err error) {
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
	if err == nil {
		p.w = p.Unescape(p.buf[:p.w])
	}
	return err
}

// HTTPHeaderValue http头域中的值
func (p *Packet) HTTPHeaderValue(key []byte) string {
	return string(bytes.TrimSpace(p.DataBetween(key, httpHeadRowAt)))
}

// Allocate 分配指定大小的缓冲区
func (p *Packet) Allocate(n int) []byte {
	p.growCap(n)
	w := p.w
	p.w += n
	return p.buf[w:p.w]
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

// Reset 重置数据包状态
func (p *Packet) Reset() {
	p.r, p.w = 0, 0
}

// Size 数据长度
func (p *Packet) Size() int {
	return p.w - p.r
}

// Seek 定位位置
func (p *Packet) Seek(readIndex, writeIndex int) {
	if readIndex >= 0 {
		p.r = readIndex
	}
	if writeIndex >= 0 {
		p.w = writeIndex
	}
}

// ReadWhen 读取数据，直到读到dst为止
func (p *Packet) ReadWhen(dst byte) []byte {
	var st = p.r
	for p.r < p.w {
		if p.buf[p.r] == dst {
			break
		}
		p.r++
	}
	return p.buf[st:p.r]
}

// At 指定位置的字符值
func (p *Packet) At(i int) byte {
	if i < p.r || i >= p.w {
		return 0
	}
	return p.buf[i]
}

// Slice 指定位置的数据
func (p *Packet) Slice(s, e int) []byte {
	if e == -1 || e > p.w {
		e = p.w
	}
	if s > e {
		return nil
	}
	return p.buf[s:e]
}

// Data 当前数据
func (p *Packet) Data() []byte {
	return p.buf[p.r:p.w]
}

// Copy 复制当前数据
func (p *Packet) Copy() *Packet {
	ins := packetPool.Get().(*Packet)
	ins.freed = 0
	ins.rt, ins.wt = p.rt, p.wt

	data := p.Data()
	ins.r, ins.w = 0, len(data)
	buf := getBytes(ins.w)
	buf = buf[:cap(buf)]
	copy(buf, data)
	ins.buf = buf
	return ins
}

// CopyReaderToConn 从输入流中读取数据并放到输出流中去
func (p *Packet) CopyReaderToConn(conn net.Conn, src io.Reader, maxN int64) (written int64, err error) {
	return p.CopyReaderToConnWithProgress(conn, src, maxN, nil)
}

// CopyReaderToConnWithProgress 从输入流中读取数据并放到输出流中去
func (p *Packet) CopyReaderToConnWithProgress(conn net.Conn, src io.Reader, maxN int64, onProgress func(int64)) (written int64, err error) {
	const (
		BUFSIZE = 1024
		WT      = time.Second * 10
	)
	p.Reset()
	buf := p.Allocate(BUFSIZE)
	if maxN > 0 {
		src = io.LimitReader(src, maxN)
	}

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			if et := conn.SetWriteDeadline(time.Now().Add(WT)); et != nil {
				err = et
				return
			}
			nw, ew := conn.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				if onProgress != nil {
					onProgress(written)
				}
			}
			if ew != nil {
				err = ew
				return
			}
			if nr != nw {
				err = io.ErrShortWrite
				return
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			return
		}
	}
}

// MoveToEnd 将指定的区间数据移至未尾
func (p *Packet) MoveToEnd(s, e int) {
	rs := e - s
	buf := getBytes(rs)[:rs]
	copy(buf, p.buf[s:e])
	copy(p.buf[s:], p.buf[e:])
	copy(p.buf[p.w-rs:], buf)
	putBytes(buf)
}

// DataBetween 指定区间值的数据
func (p *Packet) DataBetween(from, dest []byte) []byte {
	if cap(p.buf) == 0 {
		return nil
	}

	bs := p.buf[p.r:p.w]
	start := bytes.Index(bs, from)
	if start < 0 {
		return nil
	}
	start += len(from)
	bs = bs[start:]
	end := bytes.Index(bs, dest)
	if end < 0 {
		end = len(bs)
	}
	return bs[:end]
}

// Index 查找指定的文本位置
func (p *Packet) Index(v []byte) int {
	return bytes.Index(p.buf[p.r:p.w], v)
}

// HasPrefix 是否以指定文本开头
func (p *Packet) HasPrefix(prefix []byte) bool {
	if cap(p.buf) == 0 {
		return false
	}
	return bytes.HasPrefix(p.buf[p.r:p.w], prefix)
}

// HasSuffix 是否以指定文本结尾
func (p *Packet) HasSuffix(suffix []byte) bool {
	if cap(p.buf) == 0 {
		return false
	}
	return bytes.HasSuffix(p.buf[p.r:p.w], suffix)
}

// Write 实现io.Writer
func (p *Packet) Write(data []byte) (n int, err error) {
	l := len(data)
	if l <= 0 {
		n = 0
		return
	}

	p.growCap(l)
	n = copy(p.buf[p.w:], data)
	p.w += n
	return
}

// Read 实现io.Reader
func (p *Packet) Read(data []byte) (n int, err error) {
	if p.r >= p.w {
		err = io.EOF
		return
	}

	n = copy(data, p.buf[p.r:p.w])
	p.r += n
	return
}

// WriteByte 实现io.ByteWriter
func (p *Packet) WriteByte(b byte) error {
	p.growCap(1)
	p.buf[p.w] = b
	p.w++
	return nil
}

// ReadByte 实现io.ByteReader
func (p *Packet) ReadByte() (b byte, err error) {
	if p.r >= p.w {
		err = io.EOF
	} else {
		b = p.buf[p.r]
		p.r++
	}
	return
}

// Mask 对[s, e)之间的数据进行掩码
func (p *Packet) Mask(mask []byte, s, e int) {
	m := len(mask)
	if s >= e || e > p.w || m == 0 {
		return
	}

	buf := p.buf[s:e]
	l := len(buf)
	for i := 0; i < l; i++ {
		buf[i] ^= mask[i%m]
	}
}

// growCap 增加缓冲区容量
func (p *Packet) growCap(n int) {
	c := cap(p.buf)
	neds := n + p.w
	if neds < c {
		return
	}
	buf := getBytes(neds * 2)
	buf = buf[:cap(buf)]
	if c > 0 {
		copy(buf, p.buf[:p.w])
		putBytes(p.buf)
	}
	p.buf = buf
}

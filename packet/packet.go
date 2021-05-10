package packet

import (
	"bytes"
	"io"
	"net"
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

// Allocate 分配指定大小的缓冲区
func (p *Packet) Allocate(n int) []byte {
	p.grow(n)
	w := p.w
	p.w += n
	return p.buf[w:p.w]
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

// Skip 跳过一定的字节数
func (p *Packet) Skip(count int) {
	p.r += count
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

// SliceNum 获取指定数量的数据
func (p *Packet) SliceNum(num int) []byte {
	r := p.r
	e := r + num
	if e > p.w {
		e = p.w
	}
	p.r = e
	return p.buf[r:e]
}

// Data 当前数据
func (p *Packet) Data() []byte {
	return p.buf[p.r:p.w]
}

// CopyData
func (p *Packet) CopyData() []byte {
	s := p.Size()
	cp := getBytes(s)[:s]
	copy(cp, p.Data())
	return cp
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
	copy(buf[:ins.w], data)
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

// grow 增加缓冲区容量
func (p *Packet) grow(n int) {
	m, c := p.w+n, cap(p.buf)
	if m <= c {
		return
	}
	buf := getBytes(m * 2)
	buf = buf[:cap(buf)]
	if c > 0 {
		copy(buf[:p.w], p.buf[:p.w])
		putBytes(p.buf)
	}
	p.buf = buf
}

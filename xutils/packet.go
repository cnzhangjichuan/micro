package xutils

import (
	"encoding/binary"
	"io"
	"math"
)

func NewDataPacket(data []byte) *Packet {
	return &Packet{
		buf: data,
		ri:  0,
		wi:  len(data),
	}
}

func NewPacket(n int) *Packet {
	const defBuffCap = 1024

	if n <= 0 {
		n = defBuffCap
	}

	return &Packet{
		buf: make([]byte, n),
		ri:  0,
		wi:  0,
	}
}

// 处理数据包
type Packet struct {
	buf []byte
	ri  int
	wi  int
}

// 实现io.Writer
func (p *Packet) Write(data []byte) (n int, err error) {
	err = nil
	// grow buf cap.
	p.growCap(len(data))

	// copy data
	n = copy(p.buf[p.wi:], data)
	p.wi += n
	return
}

// 实现io.Reader
func (p *Packet) Read(data []byte) (n int, err error) {
	err = nil
	n = copy(data, p.buf[p.ri:])
	p.ri += n
	return
}

// 实现io.ByteWriter
func (p *Packet) WriteByte(b byte) error {
	p.growCap(1)
	p.buf[p.wi] = b
	p.wi += 1
	return nil
}

// 实现io.ByteReader
func (p *Packet) ReadByte() (b byte, err error) {
	if p.ri >= p.wi {
		err = io.EOF
	} else {
		b = p.buf[p.ri]
		p.ri += 1
	}
	return
}

func (p *Packet) Reset() {
	p.ri = 0
	p.wi = 0
}

func (p *Packet) WriteString(s string) {
	ss := UnsafeStringToBytes(s)

	l := len(ss)

	// header size
	p.growCap(binary.MaxVarintLen64)
	n := binary.PutVarint(p.buf[p.wi:], int64(l))
	p.wi += n

	// data
	p.growCap(l)
	n = copy(p.buf[p.wi:], ss)
	p.wi += n
}

func (p *Packet) ReadString() string {
	// read header size
	n, _ := binary.ReadVarint(p)

	// read data
	ri := p.ri
	p.ri += int(n)
	if p.ri > p.wi {
		p.ri = p.wi
	}

	return string(p.buf[ri:p.ri])
}

func (p *Packet) WriteU64(x uint64) {
	p.growCap(binary.MaxVarintLen64)
	n := binary.PutUvarint(p.buf[p.wi:], x)
	p.wi += n
}

func (p *Packet) ReadU64() uint64 {
	n, _ := binary.ReadUvarint(p)
	return n
}

func (p *Packet) WriteU32(x uint32) {
	p.WriteU64(uint64(x))
}

func (p *Packet) ReadU32() uint32 {
	return uint32(p.ReadU64())
}

func (p *Packet) WriteI64(x int64) {
	p.growCap(binary.MaxVarintLen64)
	n := binary.PutVarint(p.buf[p.wi:], x)
	p.wi += n
}

func (p *Packet) ReadI64() int64 {
	n, _ := binary.ReadVarint(p)
	return n
}

func (p *Packet) WriteI32(x int32) {
	p.WriteI64(int64(x))
}

func (p *Packet) ReadI32() int32 {
	return int32(p.ReadI64())
}

func (p *Packet) WriteF32(f32 float32) {
	p.WriteU64(uint64(math.Float32bits(f32)))
}

func (p *Packet) ReadF32() float32 {
	i := p.ReadU64()
	return math.Float32frombits(uint32(i))
}

func (p *Packet) WriteF64(f64 float64) {
	p.WriteU64(math.Float64bits(f64))
}

func (p *Packet) ReadF64() float64 {
	i := p.ReadU64()
	return math.Float64frombits(i)
}

// 数据长度
func (p *Packet) Size() int {
	s := p.wi - p.ri
	if s < 0 {
		s = 0
	}
	return s
}

// 定位位置
func (p *Packet) Seek(readIndex, writeIndex int) {
	if readIndex >= 0 {
		p.ri = readIndex
	}
	if writeIndex >= 0 {
		p.wi = writeIndex
	}
}

// 当前数据
func (p *Packet) Data() []byte {
	return p.buf[p.ri:p.wi]
}

// 将数据发送到网络连接上
func (p *Packet) FlushToConn(conn io.Writer) (n int, err error) {
	n, err = conn.Write(p.buf[p.ri:p.wi])
	return
}

// 从网络连接中读取数据
func (p *Packet) ReadConn(conn io.Reader) (err error) {
	p.Reset()
	hd := p.Allocate(4)
	_, err = io.ReadFull(conn, hd)
	if err != nil {
		return
	}
	n := binary.LittleEndian.Uint32(hd)
	p.Reset()
	_, err = io.ReadFull(conn, p.Allocate(int(n)))
	return
}

// 开始写入数据
func (p *Packet) BeginWrite() {
	p.ri = 0
	p.wi = 0
	p.Allocate(4)
}

// 结束写入数据
func (p *Packet) EndWrite() {
	binary.LittleEndian.PutUint32(p.buf[:4], uint32(p.wi-4))
}

// 分配指定大小的缓冲区
func (p *Packet) Allocate(n int) []byte {
	p.growCap(n)
	wi := p.wi
	p.wi += n
	return p.buf[wi:p.wi]
}

// 扩大缓冲区容量
func (p *Packet) growCap(n int) {
	cp := len(p.buf)
	if n < cp-p.wi {
		return
	}
	var buf = p.buf[0:p.wi]
	cp += cp
	min := n + p.wi
	if cp < min {
		cp = min
	}
	p.buf = make([]byte, cp)
	copy(p.buf, buf)
}

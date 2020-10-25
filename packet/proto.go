package packet

import (
	"encoding/binary"
	"math"
	"time"
	"unsafe"
)

// WriteString 写入字符串
func (p *Packet) WriteString(s string) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	ss := *(*[]byte)(unsafe.Pointer(&h))
	l := len(ss)

	// header
	p.growCap(binary.MaxVarintLen64)
	n := binary.PutVarint(p.buf[p.w:], int64(l))
	p.w += n

	// data
	p.growCap(l)
	n = copy(p.buf[p.w:], ss)
	p.w += n
}

// WriteStrings 写入字符串数组
func (p *Packet) WriteStrings(ss []string) {
	p.WriteU64(uint64(len(ss)))
	for _, s := range ss {
		p.WriteString(s)
	}
}

// ReadString 读出字符串数组
func (p *Packet) ReadString() string {
	// read header size
	n, _ := binary.ReadVarint(p)
	if n <= 0 {
		return ""
	}

	// read data
	r := p.r
	p.r += int(n)
	if p.r > p.w {
		p.r = p.w
	}

	return string(p.buf[r:p.r])
}

// ReadStrings 读出字符串数组
func (p *Packet) ReadStrings() []string {
	c := p.ReadU64()
	ss := make([]string, 0, c)
	for i := uint64(0); i < c; i++ {
		ss = append(ss, p.ReadString())
	}
	return ss
}

// WriteU64 写入uint64
func (p *Packet) WriteU64(x uint64) {
	p.growCap(binary.MaxVarintLen64)
	n := binary.PutUvarint(p.buf[p.w:], x)
	p.w += n
}

// WriteU64S 写入[]uint64
func (p *Packet) WriteU64S(us []uint64) {
	p.WriteU64(uint64(len(us)))
	for _, s := range us {
		p.WriteU64(s)
	}
}

// ReadU64 读出uint64
func (p *Packet) ReadU64() uint64 {
	n, _ := binary.ReadUvarint(p)
	return n
}

// ReadU64S 读出[]uint64
func (p *Packet) ReadU64S() []uint64 {
	c := p.ReadU64()
	us := make([]uint64, 0, c)
	for i := uint64(0); i < c; i++ {
		us = append(us, p.ReadU64())
	}
	return us
}

// WriteU32 写入uint32
func (p *Packet) WriteU32(x uint32) {
	p.WriteU64(uint64(x))
}

// WriteU32S 写入[]uint32
func (p *Packet) WriteU32S(us []uint32) {
	p.WriteU64(uint64(len(us)))
	for _, s := range us {
		p.WriteU32(s)
	}
}

// ReadU32 读出uint32
func (p *Packet) ReadU32() uint32 {
	return uint32(p.ReadU64())
}

// ReadU32S 读出[]uint32
func (p *Packet) ReadU32S() []uint32 {
	c := p.ReadU64()
	us := make([]uint32, 0, c)
	for i := uint64(0); i < c; i++ {
		us = append(us, p.ReadU32())
	}
	return us
}

// WriteI64 写入int64
func (p *Packet) WriteI64(x int64) {
	p.growCap(binary.MaxVarintLen64)
	n := binary.PutVarint(p.buf[p.w:], x)
	p.w += n
}

// WriteI64S 写入[]int64
func (p *Packet) WriteI64S(is []int64) {
	p.WriteU64(uint64(len(is)))
	for _, s := range is {
		p.WriteI64(s)
	}
}

// ReadI64 写出int64
func (p *Packet) ReadI64() int64 {
	n, _ := binary.ReadVarint(p)
	return n
}

// ReadI64S 写出[]int64
func (p *Packet) ReadI64S() []int64 {
	c := p.ReadU64()
	is := make([]int64, 0, c)
	for i := uint64(0); i < c; i++ {
		is = append(is, p.ReadI64())
	}
	return is
}

// WriteI32 写入int32
func (p *Packet) WriteI32(x int32) {
	p.WriteI64(int64(x))
}

// ReadI32 读出int32
func (p *Packet) ReadI32() int32 {
	return int32(p.ReadI64())
}

// ReadI32S 读出[]int32
func (p *Packet) ReadI32S() []int32 {
	c := p.ReadU64()
	is := make([]int32, 0, c)
	for i := uint64(0); i < c; i++ {
		is = append(is, p.ReadI32())
	}
	return is
}

// WriteI32S 写入[]int32
func (p *Packet) WriteI32S(is []int32) {
	p.WriteU64(uint64(len(is)))
	for _, s := range is {
		p.WriteI32(s)
	}
}

// WriteF32 写入float32
func (p *Packet) WriteF32(f32 float32) {
	p.WriteU64(uint64(math.Float32bits(f32)))
}

// WriteF32S 写入[]float32
func (p *Packet) WriteF32S(fs []float32) {
	p.WriteU64(uint64(len(fs)))
	for _, s := range fs {
		p.WriteF32(s)
	}
}

// ReadF32 读出float32
func (p *Packet) ReadF32() float32 {
	i := p.ReadU64()
	return math.Float32frombits(uint32(i))
}

// ReadF32S 读出[]float32
func (p *Packet) ReadF32S() []float32 {
	c := p.ReadU64()
	fs := make([]float32, 0, c)
	for i := uint64(0); i < c; i++ {
		fs = append(fs, p.ReadF32())
	}
	return fs
}

// WriteF64 写入float64
func (p *Packet) WriteF64(f64 float64) {
	p.WriteU64(math.Float64bits(f64))
}

// WriteF64S 写入[]float64
func (p *Packet) WriteF64S(fs []float64) {
	p.WriteU64(uint64(len(fs)))
	for _, s := range fs {
		p.WriteF64(s)
	}
}

// ReadF64 读出float64
func (p *Packet) ReadF64() float64 {
	i := p.ReadU64()
	return math.Float64frombits(i)
}

// ReadF64S 读出[]float64
func (p *Packet) ReadF64S() []float64 {
	c := p.ReadU64()
	fs := make([]float64, 0, c)
	for i := uint64(0); i < c; i++ {
		fs = append(fs, p.ReadF64())
	}
	return fs
}

// WriteTime 写入时间
func (p *Packet) WriteTime(t time.Time) {
	p.WriteI64(t.Unix())
}

// ReadTime 读出时间
func (p *Packet) ReadTime() time.Time {
	return time.Unix(p.ReadI64(), 0)
}

// WriteBool 写入布尔值
func (p *Packet) WriteBool(b bool) {
	if b {
		p.WriteByte(1)
	} else {
		p.WriteByte(0)
	}
}

// WriteBools 写入[]bool
func (p *Packet) WriteBools(bs []bool) {
	p.WriteU64(uint64(len(bs)))
	for _, b := range bs {
		p.WriteBool(b)
	}
}

// ReadBool 读出布尔值
func (p *Packet) ReadBool() bool {
	b, err := p.ReadByte()
	if err != nil {
		return false
	}
	return b == 1
}

// ReadBools 读出[]bool
func (p *Packet) ReadBools() []bool {
	c := p.ReadU64()
	bs := make([]bool, 0, c)
	for i := uint64(0); i < c; i++ {
		bs = append(bs, p.ReadBool())
	}
	return bs
}
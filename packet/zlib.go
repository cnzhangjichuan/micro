package packet

import (
	"compress/gzip"
	"compress/zlib"
	"encoding/hex"
	"io"
	"unsafe"
)

// UnCompress 解压数据
func UnCompress(data []byte, call func(*Packet)) bool {
	const BS = 1024

	if len(data) == 0 {
		return false
	}

	ls := hex.DecodedLen(len(data))
	p := New(ls)

	var err error
	p.w, err = hex.Decode(p.Allocate(ls), data)
	if err != nil {
		Free(p)
		return false
	}
	if b, _ := p.ReadByte(); b == 0 {
		call(p)
		Free(p)
		return true
	}

	gzr, err := zlib.NewReader(p)
	if err != nil {
		call(p)
		Free(p)
		return true
	}
	ret := New(10 * ls)
	buf := getBytes(BS)[:BS]
	io.CopyBuffer(ret, gzr, buf)
	putBytes(buf)
	gzr.Close()
	Free(p)
	call(ret)
	Free(ret)
	return true
}

// UnCompress 从指定的位置解压数据
func (p *Packet) UnCompress(s int) {
	const BS = 1024

	if s < 0 {
		s = p.r
	}
	r := p.r
	p.Seek(s, -1)
	gzr, err := zlib.NewReader(p)
	if err != nil {
		p.r = r
		return
	}
	rpf := New(10 * (p.w - s))
	buf := getBytes(BS)[:BS]
	io.CopyBuffer(rpf, gzr, buf)
	putBytes(buf)
	gzr.Close()
	p.r, p.w = r, s
	p.Write(rpf.Data())
	Free(rpf)
}

// GzipUnCompress 从指定的位置解压数据(Gzip)
func (p *Packet) GzipUnCompress(s int) {
	const BS = 1024

	if s < 0 {
		s = p.r
	}
	r := p.r
	p.Seek(s, -1)
	gzr, err := gzip.NewReader(p)
	if err != nil {
		p.r = r
		return
	}
	rpf := New(10 * (p.w - s))
	buf := getBytes(BS)[:BS]
	io.CopyBuffer(rpf, gzr, buf)
	putBytes(buf)
	gzr.Close()
	p.r, p.w = r, s
	p.Write(rpf.Data())
	Free(rpf)
}

// Compress 压缩数据
func Compress(data []byte, call func(string)) bool {
	if len(data) == 0 {
		call(`00`)
		return false
	}

	ok := true
	wpf := New(len(data) * 2)
	wpf.WriteByte(1)
	gzw, _ := zlib.NewWriterLevel(wpf, zlib.BestSpeed)
	gzw.Write(data)
	gzw.Flush()
	gzw.Close()

	if wpf.Size() >= len(data) {
		ok = false
		wpf.Reset()
		wpf.WriteByte(0)
		wpf.Write(data)
	}

	s := wpf.Size()
	dst := wpf.Allocate(s * 2)
	hex.Encode(dst, wpf.Slice(0, s))
	call(*(*string)(unsafe.Pointer(&dst)))
	Free(wpf)
	return ok
}

// Compress 从指定的位置压缩数据
func (p *Packet) Compress(s int) bool {
	if s < 0 {
		s = 0
	}
	wpf := New(p.w - s)
	gzw, _ := zlib.NewWriterLevel(wpf, zlib.BestSpeed)
	gzw.Write(p.Slice(s, -1))
	gzw.Flush()
	gzw.Close()

	if wpf.Size() >= p.w-s {
		Free(wpf)
		return false
	}
	p.w = s
	p.Write(wpf.Data())
	Free(wpf)
	return true
}

// GzipCompress 从指定的位置压缩数据(Gzip)
func (p *Packet) GzipCompress(s int) bool {
	if s < 0 {
		s = 0
	}
	wpf := New(p.w - s)
	gzw, _ := gzip.NewWriterLevel(wpf, gzip.BestSpeed)
	gzw.Write(p.Slice(s, -1))
	gzw.Flush()
	gzw.Close()

	if wpf.Size() >= p.w-s {
		Free(wpf)
		return false
	}
	p.w = s
	p.Write(wpf.Data())
	Free(wpf)
	return true
}

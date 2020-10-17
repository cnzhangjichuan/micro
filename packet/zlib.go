package packet

import (
	"compress/zlib"
	"io"
)

// UnCompress 从指定的位置解压数据
func (p *Packet) UnCompress(s int) {
	const BS = 1024

	if s < 0 {
		s = p.r
	}
	r := p.r
	p.Seek(s, -1)
	gzr, _ := zlib.NewReader(p)
	rpf := New(10 * (p.w - s))
	buf := getBytes(BS)[:BS]
	io.CopyBuffer(rpf, gzr, buf)
	putBytes(buf)
	gzr.Close()
	p.r, p.w = r, s
	p.Write(rpf.Data())
	Free(rpf)
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

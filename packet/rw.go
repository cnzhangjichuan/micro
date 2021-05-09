package packet

import "io"

// Write 实现io.Writer
func (p *Packet) Write(data []byte) (n int, err error) {
	l := len(data)
	if l <= 0 {
		n = 0
		return
	}

	p.grow(l)
	n = copy(p.buf[p.w:p.w+l], data)
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
	p.grow(1)
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
package packet

import (
	"io"
	"os"
	"path/filepath"
)

// NewWithReader 通过Reader创建
func NewWithReader(r io.Reader) (pack *Packet, err error) {
	pack = New(2048)
	buf := getBytes(1024)[:1024]
	_, err = io.CopyBuffer(pack, r, buf)
	putBytes(buf)
	return
}

// NewWithFile 通过文件内容创建
func NewWithFile(name string) (pack *Packet, err error) {
	var fd *os.File
	fd, err = os.Open(name)
	if err != nil {
		return
	}
	pack = New(2048)
	buf := getBytes(1024)[:1024]
	_, err = io.CopyBuffer(pack, fd, buf)
	putBytes(buf)
	fd.Close()
	if err != nil {
		Free(pack)
		pack = nil
	}
	return
}

// LoadFromReader 从流中加载数据
func (p *Packet) LoadFromReader(r io.Reader) error {
	p.Reset()
	buf := getBytes(1024)[:1024]
	_, err := io.CopyBuffer(p, r, buf)
	putBytes(buf)
	return err
}

// LoadFile 加载文件内容
func (p *Packet) LoadFile(name string) error {
	p.Reset()
	fd, err := os.Open(name)
	if err != nil {
		return err
	}
	buf := getBytes(1024)[:1024]
	_, err = io.CopyBuffer(p, fd, buf)
	putBytes(buf)
	fd.Close()
	return err
}

// ReadAt 实现io.ReadAt
func (p *Packet) ReadAt(buf []byte, off int64) (n int, err error) {
	of := int(off)
	if of >= p.w {
		return 0, io.EOF
	}
	dst := of + len(buf)
	if dst >= p.w {
		dst = p.w
		err = io.EOF
	}
	n = dst - of
	copy(buf, p.buf[of:dst])
	return
}

// SaveToFile 保存到文件
func (p *Packet) SaveToFile(name string) error {
	os.MkdirAll(filepath.Dir(name), os.ModePerm)
	fd, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = fd.Write(p.buf[:p.w])
	fd.Close()

	return err
}

package packet

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

// AutoCode 生成packet包协议代码
func AutoCode(dirName string) error {
	c := codec{
		payload: make(map[string][]structure, 64),
	}
	return c.AutoCode(dirName)
}

// codec 代码生成器
type codec struct {
	payload map[string][]structure
}

// AutoCode 生成packet.Encoder和packet.Decoder
func (c *codec) AutoCode(dirName string) error {
	err := c.parseDir(dirName)
	if err != nil {
		return err
	}

	for fdr, ss := range c.payload {
		fd, err := os.Create(filepath.Join(fdr, "codec.go"))
		if err != nil {
			return err
		}
		buf := bufio.NewWriter(fd)
		buf.WriteString(strings.Join([]string{
			`package `, filepath.Base(fdr), "\n\n",
			`import "github.com/micro/packet"`, "\n\n",
		}, ""))

		// structures
		for i := 0; i < len(ss); i++ {
			// Decode
			buf.WriteString(strings.Join([]string{
				`// Decode `, ss[i].Name, ` generate by codec.`, "\n",
				`func (o *`, ss[i].Name, `) Decode(p *packet.Packet) {`, "\n",
			}, ""))
			for _, ft := range ss[i].Fields {
				c.decodeFields(buf, ft)
			}
			buf.WriteString(strings.Join([]string{
				`}`, "\n\n",
			}, ""))

			// Encode
			buf.WriteString(strings.Join([]string{
				`// Encode `, ss[i].Name, ` generate by codec.`, "\n",
				`func (o *`, ss[i].Name, `) Encode(p *packet.Packet) {`, "\n",
			}, ""))
			for _, ft := range ss[i].Fields {
				c.encodeFields(buf, ft)
			}
			buf.WriteString(strings.Join([]string{
				`}`, "\n\n\n",
			}, ""))
		}

		err = buf.Flush()
		fd.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// structure 结构体描述
type structure struct {
	Name   string
	Fields [][]string
}

func (c *codec) parseDir(dirName string) error {
	return filepath.Walk(dirName, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		fd, err := os.Open(p)
		if err != nil {
			return err
		}
		buf := bufio.NewReader(fd)
		packageName := filepath.Dir(p)
		ss := c.payload[packageName]
		ss = c.parseFile(ss, buf)
		fd.Close()
		c.payload[packageName] = ss

		return nil
	})
}

// parseFile 从文件中解析结构体数据
func (c *codec) parseFile(ss []structure, buf *bufio.Reader) []structure {
	prefix := []byte(`type `)
	sfg := []byte(` struct `)
	sed := []byte(`}`)
	spl := []byte(` `)
	spn := []byte(`,`)
	doc := []byte(`//`)
	fType := ""
	for {
		line, _, err := buf.ReadLine()
		if err != nil {
			return ss
		}
		if bytes.HasPrefix(line, doc) {
			continue
		}
		if !bytes.HasPrefix(line, prefix) || bytes.Index(line, sfg) <= 0 {
			continue
		}
		var s structure
		s.Name = string(bytes.Split(line, spl)[1])

		// read fields
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				break
			}
			line = bytes.TrimSpace(line)
			if bytes.HasPrefix(line, sed) {
				break
			}
			if len(line) == 0 {
				continue
			}
			if bytes.HasPrefix(line, doc) {
				continue
			}

			row := bytes.Split(line, spl)
			fType = ""
			for i := 1; i < len(row); i++ {
				t := bytes.TrimSpace(row[i])
				if len(t) > 0 {
					fType = string(t)
					break
				}
			}
			ns := bytes.Split(row[0], spn)
			for _, n := range ns {
				s.Fields = append(s.Fields, []string{string(bytes.TrimSpace(n)), fType})
			}
		}
		ss = append(ss, s)
	}
}

// encodeFields 对结构体字段进行编码
func (c *codec) encodeFields(fd *bufio.Writer, ft []string) {
	switch ft[1] {
	default:
		if strings.HasPrefix(ft[1], "map[") {
			// not support map.
		} else if strings.HasPrefix(ft[1], "[]") {
			fd.WriteString(strings.Join([]string{
				`	c`, ft[0], ` := uint64(len(o.`, ft[0], `))`, "\n",
				`	p.WriteU64(c`, ft[0], `)`, "\n",
				`	for i := uint64(0); i < c`, ft[0], `; i++ {`, "\n",
				`		o.`, ft[0], `[i].Encode(p)`, "\n",
				`	}`, "\n",
			}, ""))
		} else {
			fd.WriteString(strings.Join([]string{
				`	o.`, ft[0], `.Encode(p)`, "\n",
			}, ""))
		}
	case "string":
		fd.WriteString(strings.Join([]string{
			`	p.WriteString(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]string":
		fd.WriteString(strings.Join([]string{
			`	p.WriteStrings(o.`, ft[0], `)`, "\n",
		}, ""))
	case "uint64":
		fd.WriteString(strings.Join([]string{
			`	p.WriteU64(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]uint64":
		fd.WriteString(strings.Join([]string{
			`	p.WriteU64S(o.`, ft[0], `)`, "\n",
		}, ""))
	case "uint32":
		fd.WriteString(strings.Join([]string{
			`	p.WriteU32(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]uint32":
		fd.WriteString(strings.Join([]string{
			`	p.WriteU32S(o.`, ft[0], `)`, "\n",
		}, ""))
	case "int64":
		fd.WriteString(strings.Join([]string{
			`	p.WriteI64(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]int64":
		fd.WriteString(strings.Join([]string{
			`	p.WriteI64S(o.`, ft[0], `)`, "\n",
		}, ""))
	case "int32":
		fd.WriteString(strings.Join([]string{
			`	p.WriteI32(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]int32":
		fd.WriteString(strings.Join([]string{
			`	p.WriteI32S(o.`, ft[0], `)`, "\n",
		}, ""))
	case "float32":
		fd.WriteString(strings.Join([]string{
			`	p.WriteF32(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]float32":
		fd.WriteString(strings.Join([]string{
			`	p.WriteF32S(o.`, ft[0], `)`, "\n",
		}, ""))
	case "float64":
		fd.WriteString(strings.Join([]string{
			`	p.WriteF64(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]float64":
		fd.WriteString(strings.Join([]string{
			`	p.WriteF64S(o.`, ft[0], `)`, "\n",
		}, ""))
	case "bool":
		fd.WriteString(strings.Join([]string{
			`	p.WriteBool(o.`, ft[0], `)`, "\n",
		}, ""))
	case "[]bool":
		fd.WriteString(strings.Join([]string{
			`	p.WriteBools(o.`, ft[0], `)`, "\n",
		}, ""))
	case "time.Time":
		fd.WriteString(strings.Join([]string{
			`	p.WriteTime(o.`, ft[0], `)`, "\n",
		}, ""))
	}
}

// decodeFields 对结构体字段进行解码
func (c *codec) decodeFields(fd *bufio.Writer, ft []string) {
	switch ft[1] {
	default:
		if strings.HasPrefix(ft[1], "map[") {
			// not support map.
		} else if strings.HasPrefix(ft[1], "[]") {
			fd.WriteString(strings.Join([]string{
				`	c`, ft[0], ` := p.ReadU64()`, "\n",
				`	o.`, ft[0], ` = make(`, ft[1], `, c`, ft[0], `)`, "\n",
				`	for i := uint64(0); i < c`, ft[0], `; i++ {`, "\n",
				`		o.`, ft[0], `[i].Decode(p)`, "\n",
				`	}`, "\n",
			}, ""))
		} else {
			fd.WriteString(strings.Join([]string{
				`	o.`, ft[0], `.Decode(p)`, "\n",
			}, ""))
		}
	case "string":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadString()`, "\n",
		}, ""))
	case "[]string":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadStrings()`, "\n",
		}, ""))
	case "uint64":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadU64()`, "\n",
		}, ""))
	case "[]uint64":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadU64S()`, "\n",
		}, ""))
	case "uint32":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadU32()`, "\n",
		}, ""))
	case "[]uint32":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadU32S()`, "\n",
		}, ""))
	case "int64":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadI64()`, "\n",
		}, ""))
	case "[]int64":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadI64S()`, "\n",
		}, ""))
	case "int32":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadI32()`, "\n",
		}, ""))
	case "[]int32":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadI32S()`, "\n",
		}, ""))
	case "float32":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadF32()`, "\n",
		}, ""))
	case "[]float32":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadF32S()`, "\n",
		}, ""))
	case "float64":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadF64()`, "\n",
		}, ""))
	case "[]float64":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadF64S()`, "\n",
		}, ""))
	case "bool":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadBool()`, "\n",
		}, ""))
	case "[]bool":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadBools()`, "\n",
		}, ""))
	case "time.Time":
		fd.WriteString(strings.Join([]string{
			`	o.`, ft[0], ` = p.ReadTime()`, "\n",
		}, ""))
	}
}

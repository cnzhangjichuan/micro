package config

import (
	"errors"
	"io"

	"github.com/micro/packet"
	"github.com/micro/xutils"
	"github.com/xlsx"
)

// Save 更新数据
func (e *excel) Save(name string, r io.Reader) error {
	const dataStartIndex = 3

	e.Lock()
	defer e.Unlock()

	pack, err := packet.NewWithReader(r)
	if err != nil {
		return err
	}
	xf, err := xlsx.OpenReaderAt(pack, int64(pack.Size()))
	if err != nil {
		packet.Free(pack)
		return err
	}
	if len(xf.Sheets) == 0 {
		packet.Free(pack)
		return errors.New(`excel: empty data`)
	}
	rows := xf.Sheets[0].Rows
	if len(rows) < dataStartIndex {
		packet.Free(pack)
		return errors.New(`excel: too few rows`)
	}
	pack.Reset()

	// 设置列名信息
	names, types := e.title(rows)
	pack.WriteStrings(names)

	// 设置数据大小
	realRows := int64(0)
	for r := dataStartIndex; r < len(rows); r++ {
		cs := rows[r].Cells
		if len(cs) == 0 || cs[0].String() == "" {
			break
		}
		realRows++
	}
	pack.WriteI64(realRows)

	// 写入行数据
	for i := int64(0); i < realRows; i++ {
		e.write(pack, rows[dataStartIndex+i], types)
	}
	data := pack.CopyData()

	// 替换当前数据
	out := packet.New(4096)
	replaced := false
	if err = pack.LoadFile(e.path); err == nil {
		for {
			n := pack.ReadString()
			if n == "" {
				break
			}
			out.WriteString(n)
			s := pack.ReadI64()
			if n == name {
				// 新的数据
				replaced = true
				pack.Skip(int(s))
				out.WriteI64(int64(len(data)))
				out.Write(data)
			} else {
				// 原数据
				out.WriteI64(s)
				out.Write(pack.SliceNum(int(s)))
			}
		}
	}
	packet.Free(pack)

	// 如果没有替换成功, 直接追加到文件尾部
	if !replaced {
		out.WriteString(name)
		out.WriteI64(int64(len(data)))
		out.Write(data)
	}

	err = out.SaveToFile(e.path)
	packet.Free(out)
	return err
}

// title 获取表头信息
func (e *excel) title(rows []*xlsx.Row) (names, types []string) {
	const (
		nameRowIndex = 1
		typeRowIndex = 2
	)
	nCells, tCells := rows[nameRowIndex].Cells, rows[typeRowIndex].Cells
	nls := len(nCells)
	names = make([]string, 0, nls)
	tls := len(tCells)
	types = make([]string, nls)
	for i, c := range nCells {
		v := c.String()

		// 如果字段没有命名,跳过
		if v == "" {
			types[i] = ""
			continue
		}

		if i < tls {
			types[i] = tCells[i].String()
		} else {
			types[i] = ""
		}

		switch types[i] {
		case typeString, typeSString, typeFloat, typeSFloat, typeInt, typeSInt, typeDate, typeSDate:
			names = append(names, v)
		}
	}
	return
}

// write 将一行内容写入到packet中
func (e *excel) write(pack *packet.Packet, row *xlsx.Row, types []string) {
	var (
		cells = row.Cells
		cls   = len(cells)
	)
	for i, typ := range types {
		v := ""
		if i < cls {
			v = cells[i].String()
		}
		switch typ {
		case typeString, typeSString:
			pack.WriteString(v)
		case typeFloat, typeSFloat:
			if v != "" {
				pack.WriteF32(xutils.ParseF32(v, 0))
			} else {
				pack.WriteF32(0)
			}
		case typeInt, typeSInt:
			if v != "" {
				pack.WriteI32(xutils.ParseI32(v, 0))
			} else {
				pack.WriteI32(0)
			}
		case typeDate, typeSDate:
			if x, err := xutils.ParseTime(v); err == nil {
				pack.WriteI64(x)
			} else {
				pack.WriteI64(0)
			}
		}
	}
}

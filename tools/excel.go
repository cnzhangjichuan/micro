package tools

import (
	"errors"
	"io"

	"github.com/micro"
	"github.com/micro/packet"
	"github.com/micro/xutils"
	"github.com/xlsx"
)

var errExcelEmptyData = errors.New("excel: empty data")

// ExcelReadFrom 从文件中读取数据
func ExcelReadFrom(name string, onSetup func(uint64), onReadRow func(*packet.Packet, []string, uint64)) {
	pack, err := packet.NewWithFile(name)
	if err != nil {
		micro.Errorf("load %s error: %v", name, err)
		return
	}

	names := pack.ReadStrings()
	size := pack.ReadU64()
	onSetup(size)
	for i := uint64(0); i < size; i++ {
		onReadRow(pack, names, i)
	}
	packet.Free(pack)
}

// ExcelSaveTo 将Excel数据保存到指定的文件中
func ExcelSaveTo(r io.Reader, dst string) error {
	return ExcelSaveToWithTypeIndex(r, dst, 3)
}

// ExcelSaveToWithTypeIndex 将Excel数据保存到指定的文件中
func ExcelSaveToWithTypeIndex(r io.Reader, dst string, typeIndex int) error {
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
		return errExcelEmptyData
	}
	rows := xf.Sheets[0].Rows
	if len(rows) < typeIndex+1 {
		packet.Free(pack)
		return errExcelEmptyData
	}
	pack.Reset()

	// 设置列名信息
	names, types := excelGetRowTypes(rows, typeIndex)
	pack.WriteStrings(names)

	// 设置数据
	rowStart := typeIndex + 1
	// 获取真实的行号
	rowLast := rowStart
	for r := rowStart; r < len(rows); r++ {
		cs := rows[r].Cells
		if len(cs) == 0 || cs[0].String() == "" {
			break
		}
		rowLast++
	}
	pack.WriteU64(uint64(rowLast - rowStart))
	for r := rowStart; r < rowLast; r++ {
		excelWriteRowToPacket(pack, rows[r], types)
	}
	err = pack.SaveToFile(dst)
	packet.Free(pack)

	return err
}

// excelGetRowTypes 获取列类型
func excelGetRowTypes(rows []*xlsx.Row, typeIndex int) (names, types []string) {
	nCells, tCells := rows[1].Cells, rows[typeIndex].Cells
	names = make([]string, 0, len(nCells))
	types = make([]string, len(tCells))
	for i, c := range nCells {
		v := c.String()
		if v != "" {
			names = append(names, v)
			types[i] = tCells[i].String()
			if types[i] == "" {
				types[i] = "string"
			}
		} else {
			types[i] = ""
		}
	}
	return
}

// excelWriteRowToPacket 将Excel内容写入到packet中
func excelWriteRowToPacket(pack *packet.Packet, row *xlsx.Row, types []string) {
	var (
		cells = row.Cells
		v     string
	)
	for i, typ := range types {
		if len(cells) > i {
			v = cells[i].String()
		} else {
			v = ""
		}
		switch typ {
		default:
			pack.WriteI32(xutils.ParseI32(v, 0))
		case "":
			// empty type, do nothing
		case "float32", "float":
			pack.WriteF32(xutils.ParseF32(v, 0))
		case "float64":
			pack.WriteF64(xutils.ParseF64(v, 1))
		case "string", "strings":
			pack.WriteString(v)
		}
	}
}

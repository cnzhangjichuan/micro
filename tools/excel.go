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
	if len(rows) < 4 {
		packet.Free(pack)
		return errExcelEmptyData
	}
	pack.Reset()

	// 设置列名信息
	names, types := excelGetRowTypes(rows)
	pack.WriteStrings(names)

	// 设置数据
	pack.WriteU64(uint64(len(rows) - 4))
	for r := 4; r < len(rows); r++ {
		excelWriteRowToPacket(pack, rows[r], types)
	}
	err = pack.SaveToFile(dst)
	packet.Free(pack)

	return err
}

// excelGetRowTypes 获取列类型
func excelGetRowTypes(rows []*xlsx.Row) (names, types []string) {
	nCells, tCells := rows[1].Cells, rows[3].Cells
	names = make([]string, len(nCells))
	types = make([]string, len(tCells))
	for i, c := range nCells {
		names[i] = c.String()
		types[i] = tCells[i].String()
	}
	return
}

// excelWriteRowToPacket 将Excel内容写入到packet中
func excelWriteRowToPacket(pack *packet.Packet, row *xlsx.Row, types []string) {
	cells := row.Cells
	for i, typ := range types {
		switch typ {
		default:
			if len(cells) > i {
				pack.WriteI32(xutils.ParseI32(cells[i].String(), 1))
			} else {
				pack.WriteI32(0)
			}
		case "float32", "float":
			if len(cells) > i {
				pack.WriteF32(xutils.ParseF32(cells[i].String(), 1))
			} else {
				pack.WriteF32(0)
			}
		case "float64":
			if len(cells) > i {
				pack.WriteF64(xutils.ParseF64(cells[i].String(), 1))
			} else {
				pack.WriteF64(0)
			}
		case "string":
			if len(cells) > i {
				pack.WriteString(cells[i].String())
			} else {
				pack.WriteString("")
			}
		}
	}
}

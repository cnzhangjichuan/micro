package tools

import (
	"errors"
	"io"
	"reflect"

	"github.com/micro"
	"github.com/micro/packet"
	"github.com/micro/xutils"
	"github.com/xlsx"
)

var errExcelEmptyData = errors.New("excel: empty data")

type OnSetupFunc func(uint64) interface{}

// ExcelUnmarshal 将excel流解析成结构体
func ExcelUnmarshal(r io.Reader, typeIndex int, onSetup OnSetupFunc) error {
	// 读取数据
	pack, err := ExcelSax(r, typeIndex)
	if err != nil {
		return err
	}

	// 放置数据
	ExcelAutoReadFromCache(pack, onSetup)
	packet.Free(pack)
	return nil
}

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

// ExcelReadFrom 从文件中读取数据
func ExcelAutoReadFrom(name string, onSetup OnSetupFunc) {
	pack, err := packet.NewWithFile(name)
	if err != nil {
		micro.Errorf("load %s error: %v", name, err)
		return
	}
	ExcelAutoReadFromCache(pack, onSetup)
	packet.Free(pack)
}

// ExcelReadFrom 从缓存中读取数据
func ExcelAutoReadFromCache(pack *packet.Packet, onSetup OnSetupFunc) {
	names := pack.ReadStrings()
	size := pack.ReadU64()
	arrayValue := reflect.ValueOf(onSetup(size))
	for i := uint64(0); i < size; i++ {
		value := arrayValue.Index(int(i))
		for _, n := range names {
			f := value.FieldByName(n)
			switch f.Kind() {
			case reflect.String:
				f.SetString(pack.ReadString())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				f.SetInt(pack.ReadI64())
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				f.SetUint(pack.ReadU64())
			case reflect.Float32:
				f.SetFloat(float64(pack.ReadF32()))
			case reflect.Float64:
				f.SetFloat(pack.ReadF64())
			case reflect.Struct:
				m := f.Addr().MethodByName("Parse")
				if m.IsValid() {
					m.Call([]reflect.Value{reflect.ValueOf(pack.ReadString())})
				} else {
					pack.ReadString()
				}
			case reflect.Ptr:
				f.Set(reflect.New(f.Type().Elem()))
				m := f.MethodByName("Parse")
				if m.IsValid() {
					m.Call([]reflect.Value{reflect.ValueOf(pack.ReadString())})
				} else {
					pack.ReadString()
				}
			}
		}
	}
}

// ExcelSaveTo 将Excel数据保存到指定的文件中
func ExcelSaveTo(r io.Reader, dst string) error {
	return ExcelSaveToWithTypeIndex(r, dst, 3)
}

// ExcelSaveToWithTypeIndex 将Excel数据保存到指定的文件中
func ExcelSaveToWithTypeIndex(r io.Reader, dst string, typeIndex int) error {
	pack, err := ExcelSax(r, typeIndex)
	if err != nil {
		return err
	}

	// 保存到文件
	err = pack.SaveToFile(dst)
	packet.Free(pack)

	return err
}

// ExcelSax 将Excel数据解析到缓存中
func ExcelSax(r io.Reader, typeIndex int) (*packet.Packet, error) {
	pack, err := packet.NewWithReader(r)
	if err != nil {
		return nil, err
	}

	xf, err := xlsx.OpenReaderAt(pack, int64(pack.Size()))
	if err != nil {
		packet.Free(pack)
		return nil, err
	}
	if len(xf.Sheets) == 0 {
		packet.Free(pack)
		return nil, errExcelEmptyData
	}
	rows := xf.Sheets[0].Rows
	if len(rows) < typeIndex+1 {
		packet.Free(pack)
		return nil, errExcelEmptyData
	}
	pack.Reset()

	// 设置列名信息
	names, types := excelGetTitle(rows, typeIndex)
	pack.WriteStrings(names)

	// 获取真实的行号
	s := typeIndex + 1
	e := s
	for r := s; r < len(rows); r++ {
		cs := rows[r].Cells
		if len(cs) == 0 || cs[0].String() == "" {
			break
		}
		e++
	}
	// 写入数据大小
	pack.WriteU64(uint64(e - s))

	// 写入行数据
	for r := s; r < e; r++ {
		excelWriteRow(pack, rows[r], types)
	}

	return pack, nil
}

// excelGetRowTypes 获取表头信息
func excelGetTitle(rows []*xlsx.Row, typeIndex int) (names, types []string) {
	nCells, tCells := rows[1].Cells, rows[typeIndex].Cells
	names = make([]string, 0, len(nCells))
	types = make([]string, cap(names))
	for i, c := range nCells {
		v := c.String()
		if v != "" {
			if len(tCells) > i {
				types[i] = tCells[i].String()
			}
			if types[i] == "" {
				types[i] = "ignore"
			}
			if types[i] != "ignore" {
				names = append(names, v)
			}
		} else {
			types[i] = "ignore"
		}
	}
	return
}

// excelWriteRow 将一行内容写入到packet中
func excelWriteRow(pack *packet.Packet, row *xlsx.Row, types []string) {
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
			pack.WriteString(v)
		case "ignore":
			// 忽略该列数据
		case "float", "float32":
			pack.WriteF32(xutils.ParseF32(v, 0))
		case "float64":
			pack.WriteF64(xutils.ParseF64(v, 1))
		case "int", "uint":
			pack.WriteI64(xutils.ParseI64(v, 0))
		}
	}
}

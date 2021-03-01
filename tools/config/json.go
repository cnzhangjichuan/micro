package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/micro/xutils"
	"github.com/xlsx"
)

// ToJSON 转换成Json
func (e *excel) ToJSON(src, dst string) error {
	os.MkdirAll(dst, os.ModePerm)

	// 处理目录文件
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".xlsx") {
			return nil
		}
		name := xutils.ParseFileName(info.Name())
		return e.toJSON(path, filepath.Join(dst, name+".js"))
	})
}

func (e *excel) toJSON(src, dst string) error {
	const dataStartIndex = 3

	xf, err := xlsx.OpenFile(src)
	if err != nil {
		return err
	}
	if len(xf.Sheets) == 0 {
		return errors.New("excel: empty data")
	}

	fd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fd.Close()

	sheet := xf.Sheets[0]
	rows := sheet.Rows
	props := rows[dataStartIndex-2].Cells
	types := rows[dataStartIndex-1].Cells
	fd.WriteString("module.exports={\n")
	first := true
	for c, pn := range props {
		name := pn.String()
		if name == "" {
			continue
		}
		typ := types[c].String()
		if !e.isClientField(typ) {
			continue
		}
		if !first {
			fd.WriteString(",\n")
		} else {
			first = false
		}
		fd.WriteString(name)
		fd.WriteString(":[")
		for r := dataStartIndex; r < len(rows); r++ {
			v := ""
			if c < len(rows[r].Cells) {
				v = rows[r].Cells[c].String()
			}
			if c == 0 && v == "" {
				break
			}
			if r > dataStartIndex {
				fd.WriteString(",")
			}
			switch typ {
			default:
				if v == "" {
					v = "0"
				}
				fd.WriteString(v)
			case typeString, typeCString:
				fd.WriteString("\"")
				fd.WriteString(v)
				fd.WriteString("\"")
			case typeDate, typeCDate:
				if x, err := xutils.ParseTime(v); err == nil {
					fd.WriteString(strconv.Itoa(int(x)))
				} else {
					fd.WriteString("0")
				}
			}
		}
		fd.WriteString("]")
	}
	fd.WriteString("\n}")
	return nil
}

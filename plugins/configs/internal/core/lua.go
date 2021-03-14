package core

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/micro/xutils"
	"github.com/xlsx"
)

// ToLua 转换成lua文件
func (s *Service) ToLua(src, dst string) error {
	os.MkdirAll(dst, os.ModePerm)

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".xlsx") {
			return nil
		}
		name := xutils.ParseFileName(info.Name())
		return s.toLua(path, filepath.Join(dst, name+".lua"))
	})
}

func (s *Service) toLua(src, dst string) error {
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
	props := rows[dataStartIndex-2]
	types := rows[dataStartIndex-1]
	fd.WriteString("return {\n")
	for i, c := range props.Cells {
		name := c.String()
		if name == "" {
			continue
		}
		sty := types.Cells[i].String()
		if !s.isClientField(sty) {
			continue
		}
		fd.WriteString(name)
		fd.WriteString("={")
		for r := dataStartIndex; r < len(rows); r++ {
			v := ""
			if i < len(rows[r].Cells) {
				v = rows[r].Cells[i].String()
			}
			if r > dataStartIndex {
				fd.WriteString(",")
			}
			switch sty {
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
		fd.WriteString("},\n")
	}
	fd.WriteString(`}`)

	return nil
}

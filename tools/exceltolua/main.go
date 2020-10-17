package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/micro/xutils"
	"github.com/xlsx"
)

func main() {
	err := parseAllToLua(`../Tables`, `../LuaScript/config`)
	if err != nil {
		fmt.Println("数据转化失败: ", err)
	} else {
		fmt.Println("数据转化成功!")
	}
}

// ParseAllToLua 将srcDir目录下的excel文件转成lua文件
func parseAllToLua(srcDir, dstDir string) error {
	os.MkdirAll(dstDir, os.ModePerm)
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".xlsx") {
			return nil
		}
		name := xutils.ParseFileName(info.Name())
		err = parseToLua(path, filepath.Join(dstDir, name+".lua"))
		return err
	})
}

// ParseToLua 将excel转成lua
func parseToLua(src, dst string) error {
	xf, err := xlsx.OpenFile(src)
	if err != nil {
		return err
	}
	if len(xf.Sheets) == 0 {
		return errors.New("empty data")
	}
	fd, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fd.Close()

	sheet := xf.Sheets[0]
	rows := sheet.Rows
	head := rows[1]
	types := rows[2]
	fd.WriteString("return {\n")
	for i, c := range head.Cells {
		name := c.String()
		fd.WriteString(name)
		fd.WriteString("={")
		sty := types.Cells[i].String()
		for r := 4; r < len(rows); r++ {
			if r > 4 {
				fd.WriteString(",")
			}
			rw := rows[r]
			v := ""
			if len(rw.Cells) > i {
				v = rw.Cells[i].String()
			}
			if sty == "string" {
				fd.WriteString("\"")
				fd.WriteString(v)
				fd.WriteString("\"")
			} else if sty == "numbers_mul" {
				fd.WriteString("{")
				fd.WriteString(strings.Replace(v, ";", ",", -1))
				fd.WriteString("}")
			} else {
				if v == "" {
					fd.WriteString("0")
				} else {
					fd.WriteString(v)
				}
			}
		}
		fd.WriteString("},\n")
	}
	fd.WriteString(`}`)

	return nil
}

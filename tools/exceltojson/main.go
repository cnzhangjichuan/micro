package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/micro/xutils"
	"github.com/xlsx"
)

// 将excel数据转成json
func main() {
	const (
		src = `../tables`
		dst = `../assets/script/service/configs`
	)

	// 创建目录
	os.MkdirAll(dst, os.ModePerm)

	// 处理目录文件
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".xlsx") {
			return nil
		}
		name := xutils.ParseFileName(info.Name())
		if name == "" {
			return nil
		}
		xf, err := xlsx.OpenFile(path)
		if err != nil {
			fmt.Println("源文件解析失败：", err)
			return err
		}
		if len(xf.Sheets) == 0 {
			return nil
		}
		return parseToJSON(xf.Sheets[0].Rows, filepath.Join(dst, name+".js"))
	})
	if err == nil {
		fmt.Println("数据转化成功!")
	} else {
		fmt.Println("数据处理失败", err)
	}
}

// parseToJSON 生成json文件
func parseToJSON(rows []*xlsx.Row, fileName string) error {
	fd, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fd.Close()

	props := rows[1].Cells
	types := rows[2].Cells
	fd.WriteString("module.exports={\n")
	for c, pn := range props {
		name := pn.String()
		if name == "" {
			continue
		}
		if c > 0 {
			fd.WriteString(",\n")
		}
		fd.WriteString(name)
		fd.WriteString(":[")
		typ := types[c].String()
		for r := 3; r < len(rows); r++ {
			v := ""
			if len(rows[r].Cells) > c {
				v = rows[r].Cells[c].String()
			}
			if c == 0 && v == "" {
				break
			}
			if r > 3 {
				fd.WriteString(",")
			}
			switch typ {
			default:
				fd.WriteString("\"")
				fd.WriteString(v)
				fd.WriteString("\"")
			case "strings":
				fd.WriteString("[\"")
				if strings.Index(v, ";") >= 0 {
					fd.WriteString(strings.ReplaceAll(v, ";", "\",\""))
				} else {
					fd.WriteString(strings.ReplaceAll(v, ",", "\",\""))
				}
				fd.WriteString("\"]")
			case "int", "float":
				if v == "" {
					v = "0"
				}
				fd.WriteString(v)
			case "ints", "floats":
				fd.WriteString("[")
				fd.WriteString(v)
				fd.WriteString("]")
			}
		}
		fd.WriteString("]")
	}
	fd.WriteString("\n}")
	return nil
}

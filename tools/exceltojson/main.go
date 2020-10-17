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
		src = `./总表.xlsx`
		dst = `../assets/script/service/configs`
	)

	// 创建目录
	os.MkdirAll(dst, os.ModePerm)

	xf, err := xlsx.OpenFile(src)
	if err != nil {
		fmt.Println("源文件解析失败：", err)
		return
	}

	if len(xf.Sheets) == 0 {
		fmt.Println("没有可用的Sheet")
		return
	}

	for i := 0; i < len(xf.Sheets); i++ {
		sheet := xf.Sheets[i]
		fileName := xutils.ParseFileName(sheet.Name)
		if fileName == "" {
			// 没有配置文件名
			continue
		}
		err = parseToJSON(sheet.Rows, filepath.Join(dst, fileName+".js"))
		if err != nil {
			fmt.Printf("转换成JSON时出错[%s]: %v\n", fileName, err)
			return
		}
	}
	fmt.Println("数据转化成功!")
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
		if c > 0 {
			fd.WriteString(",\n")
		}
		name := pn.String()
		fd.WriteString(name)
		fd.WriteString(":[")
		typ := types[c].String()
		for r := 3; r < len(rows); r++ {
			if r > 3 {
				fd.WriteString(",")
			}
			v := ""
			if len(rows[r].Cells) > c {
				v = rows[r].Cells[c].String()
			}
			switch typ {
			default:
				fd.WriteString("\"")
				fd.WriteString(v)
				fd.WriteString("\"")
			case "strings":
				fd.WriteString("[\"")
				fd.WriteString(strings.ReplaceAll(v, ",", "\",\""))
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

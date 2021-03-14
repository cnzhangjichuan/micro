package main

import (
	"fmt"

	"github.com/micro/plugins/configs/internal/core"
)

// 将excel数据转成json
func main() {
	var c core.Service

	err := c.ToJSON(`../tables`, `../assets/script/configs`)
	if err != nil {
		fmt.Println("数据转化失败: ", err)
	} else {
		fmt.Println("数据转化成功!")
	}
}

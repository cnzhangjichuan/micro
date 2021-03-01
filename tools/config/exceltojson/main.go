package main

import (
	"fmt"
	"github.com/micro/tools/config"
)

// 将excel数据转成json
func main() {
	err := config.ToJSON(`../tables`, `../assets/script/configs`)
	if err != nil {
		fmt.Println("数据转化失败: ", err)
	} else {
		fmt.Println("数据转化成功!")
	}
}

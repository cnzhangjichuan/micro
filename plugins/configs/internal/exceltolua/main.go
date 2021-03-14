package main

import (
	"fmt"
	"github.com/micro/plugins/configs/internal/core"
)

func main() {
	var c core.Service

	err := c.ToLua(`../Tables`, `../LuaScript/config`)
	if err != nil {
		fmt.Println("数据转化失败: ", err)
	} else {
		fmt.Println("数据转化成功!")
	}
}

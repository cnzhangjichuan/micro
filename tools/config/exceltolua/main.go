package main

import (
	"fmt"
	
	"github.com/micro/tools/config"
)

func main() {
	err := config.ToLua(`../Tables`, `../LuaScript/config`)
	if err != nil {
		fmt.Println("数据转化失败: ", err)
	} else {
		fmt.Println("数据转化成功!")
	}
}

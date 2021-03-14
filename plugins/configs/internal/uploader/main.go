package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/micro"
	"github.com/micro/xutils"
)

// 上载配置文件到服务器上
func main() {
	var (
		devRemoteAddress string   // 开发地址
		disRemoteAddress string   // 发布地址
		serviceMask      string   // 服务器掩码
		files            []string // 文件列表
	)

	err := xutils.ReadLineFile(`./configs`, func(s string) error {
		if s == "" {
			return nil
		}

		if devRemoteAddress == "" {
			devRemoteAddress = s
			return nil
		}

		if disRemoteAddress == "" {
			disRemoteAddress = s
			return nil
		}

		if serviceMask == "" {
			serviceMask = s
			return nil
		}

		_, name := filepath.Split(s)
		files = append(files, xutils.ParseFileName(name), s)
		return nil
	})
	if err != nil {
		fmt.Println("load configs errors:", err)
		return
	}

	if xutils.HasString(os.Args, `release`) {
		// 上传到生产环境
		micro.RequestUploadService(disRemoteAddress, serviceMask, files)
	} else {
		// 上传到开发环境
		micro.RequestUploadService(devRemoteAddress, serviceMask, files)
	}
}

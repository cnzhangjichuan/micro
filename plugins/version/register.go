package version

import (
	"github.com/micro"
	"github.com/micro/plugins/version/internal/core"
)

// Register 注册版本管理模块
// rootUrl 版本文件的根url，文件的最终下载地址为rootUr+fileName.zip
func Register(rootUrl string) {
	const api = `version`

	var srv core.Service

	// 初始化服务
	srv.Init(rootUrl)

	// 版本校验
	micro.Register(api, srv.CheckVersion)

	// 重载版本配置
	micro.RegisterReloadFunc(srv.ReloadVersionFiles)

	// 文件上传
	micro.RegisterUploadFunc(api, srv.UploadFile)
}

// Difference 比对文件
func Difference(dst string, only []string, sVersion bool) error {
	var srv core.Service
	return srv.Difference(dst, only, sVersion)
}

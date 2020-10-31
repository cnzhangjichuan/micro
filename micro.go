package micro

import (
	"os"
	"strings"
)

// Service 开启服务
func Service(onStartup func()) error {
	var cmd, op, param string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i-1] == "-s" {
			cmd = strings.TrimSpace(os.Args[i])
			if len(os.Args) > i+1 {
				op = os.Args[i+1]
			}
			if len(os.Args) > i+2 {
				param = os.Args[i+2]
			}
			break
		}
	}

	switch cmd {
	default:
		// 启动服务
		return startupService(onStartup)

	case `reload`:
		// 重载服务配置
		return requestReloadService()

	case `stop`:
		// 关闭服务
		return requestCloseService()

	case `upload`:
		// 上传文件
		return requestUploadService(op, param, localeAddress())

	case `help`:
		Log("\n" +
			"-s start: startup service\n" +
			"-s stop: shutdown running service\n" +
			"-s reload: reload config file\n" +
			"-s upload: upload file\n")
		return nil
	}
}

// SendDataAll 给所有远端发送数据
func SendDataAll(data interface{}, api string) {
	SendDataWithUIDs(data, api, nil)
}

// SendDataWithUIDs 给指定的UIDs远端发送数据
func SendDataWithUIDs(data interface{}, api string, uids []string) {
	for _, m := range env.chains {
		m.SendData(data, api, uids)
	}
}

// SendDataWithGroup 按组分发数据
func SendDataWithGroup(data interface{}, api string, flag uint8, group string) {
	for _, m := range env.chains {
		m.SendGroup(data, api, flag, group)
	}
}

// RPC 远端调用
func RPC(out, in interface{}, srvName, api string) error {
	adr := env.registry.ServerAddress(srvName)
	if adr == "" {
		return errRPCNotFoundService
	}

	return env.rpc.Call(out, in, adr, api)
}

const (
	// StaBUSY 服务器状态：繁忙状态
	StaBUSY = 1
	// StaCLOSE 服务状态：维护中
	StaCLOSE = 2
)

// ServerState 服务状态
func ServerState(srvName string) byte {
	adr := env.registry.ServerAddress(srvName)
	if adr == "" {
		return StaCLOSE
	}
	return StaBUSY
}

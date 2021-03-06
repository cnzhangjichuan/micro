package micro

import (
	"os"
	"strings"
)

// Service 开启服务
func Service(onStartup func()) error {
	var cmd string
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i-1] == "-s" {
			cmd = strings.TrimSpace(os.Args[i])
			break
		}
	}

	switch cmd {
	default:
		// 启动服务
		return startupService(onStartup)

	case `stop`:
		// 关闭服务
		return requestCloseService()

	case `help`:
		Log("\n" +
			"-s start: startup service\n" +
			"-s stop: shutdown running service\n")
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

// RPCCenter 远端调用(中心服)
func RPCCenter(api string, in, out interface{}) error {
	adr := env.config.Registry
	if adr == "" {
		return errRPCNotFoundService
	}

	return env.rpc.Call(out, in, adr, api)
}

// RPC 远端调用(指定有服务器)
func RPC(srvName, api string, in, out interface{}) error {
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

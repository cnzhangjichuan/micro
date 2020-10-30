package micro

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/micro/packet"
)

var env struct {
	// 配置信息
	config struct {
		Name        string
		Address     string
		Registry    string
		AssetsCache bool
		Expired     int
		Mask        string
		DBResource  string
		DBTables    []string
		DBSQLs      []string
	}

	// 校验码
	authorize authorize

	// 登入/登出
	onLogin  lgOutterCaller
	onLogout lgOutterCaller

	// 会话
	cache packet.Cache

	// 连接监听
	lsr net.Listener

	// 处理函数
	chains []chain

	// 注册表
	registry registry

	// RPC
	rpc rpc

	// 业务接口
	bis map[string]bisDpo

	// 重载函数
	reloadFunc []func()

	// 文件上传
	uploadFunc map[string]uploadFunc

	// 日志
	log *log.Logger
}

func init() {
	// 处理日志
	SetLogger(os.Stderr)

	// 业务接口
	env.bis = make(map[string]bisDpo, 128)
	env.reloadFunc = make([]func(), 0, 16)
	env.uploadFunc = make(map[string]uploadFunc, 16)
}

// loadConfig 加载配配置信息
func loadConfig() error {
	pack := packet.New(1024)
	err := pack.LoadConfig("./config.json", &env.config)
	packet.Free(pack)

	// 处理监听地址
	if env.config.Address == "" {
		env.config.Address = ":9000"
	} else if strings.Index(env.config.Address, ":") < 0 {
		env.config.Address = ":" + env.config.Address
	}

	// 初始化校验码
	env.authorize.Init(env.config.Mask)

	return err
}

// localeAddress 获取本机配置的地址
func localeAddress() string {
	loadConfig()

	var address string
	idx := strings.Index(env.config.Address, ":")
	if idx > 0 {
		address = env.config.Address
	} else if idx == 0 {
		address = "127.0.0.1" + env.config.Address
	} else {
		address = "127.0.0.1:" + env.config.Address
	}
	return address
}

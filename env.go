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
		Name        string   // 服务名称
		Address     string   // 监听地址
		Registry    string   // 注册机地址
		AssetsCache bool     // web资源是否需要缓存
		Expired     int      // Session过期时间
		Mask        string   // 通信掩码
		OpenAt      string   // 开服时间
		DBResource  string   // 数据源
		UserTabName string   // 玩家基础数据存储名称
		DBSQLs      []string // 需要执行的SQL
		LogFlags    byte     // lDebug/lLog/lError
		Extra       []string // 扩展参数
	}

	// 校验码
	authorize authorize

	// 登入/登出
	onLogin  lgCaller
	onLogout lgCaller

	// 会话
	userCache packet.Cache

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
	rps map[string]bisDpo

	// 服务器关闭之前执行的函数
	closeFunc []func()

	// 文件上传
	uploadFunc map[string]uploadFunc

	// 日志
	log      *log.Logger
	logFlags byte
}

// Name 服器名称
func Name() string {
	return env.config.Name
}

// 获取扩展参数
func GetExtra(name string) string {
	for i, l := 1, len(env.config.Extra); i < l; i += 2 {
		if env.config.Extra[i-1] == name {
			return env.config.Extra[i]
		}
	}
	return ``
}

func init() {
	// 处理日志
	env.logFlags = lDebug | lLog | lError
	SetLogger(os.Stderr)

	// 业务接口
	env.bis = make(map[string]bisDpo, 64)
	env.rps = make(map[string]bisDpo, 64)
	env.uploadFunc = make(map[string]uploadFunc, 16)
}

// loadConfig 加载配配置信息
func loadConfig() error {
	env.config.LogFlags = 255
	pack := packet.New(1024)
	err := pack.LoadConfig("./config.json", &env.config)
	packet.Free(pack)

	// 设置日志输出标志
	if env.config.LogFlags != 255 {
		env.logFlags = env.config.LogFlags
	}

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

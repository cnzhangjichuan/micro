package micro

import (
	"errors"
	"net"
	"runtime"
	"time"

	"github.com/micro/packet"
	"github.com/micro/store"
	"github.com/micro/xutils"
)

// startupService 启动服务
func startupService(onStartup func()) error {
	const userTableName = `users`

	err := loadConfig()
	if err != nil {
		Debug("load config error: %v", err)
	}

	env.lsr, err = net.Listen("tcp", env.config.Address)
	if err != nil {
		return err
	}

	// Session会话
	if env.config.Expired < 0 {
		env.config.Expired = 0
	}

	// 数据存储
	expired := time.Duration(env.config.Expired) * time.Second
	if env.config.DBResource != "" {
		if !store.IsBackupOnErrorSetted() {
			store.SetBackupOnError(func(SQL string, err error) {
				Logf(">> SQL execute error:\n[%s]\n%v", SQL, err)
			})
		}
		if xutils.HasString(env.config.DBTables, userTableName) {
			env.cache = packet.NewCache(expired, store.NewSaver(userTableName))
		} else {
			env.cache = packet.NewCache(expired, nil)
		}
		err = store.Init(env.config.DBResource, env.config.DBTables, env.config.DBSQLs)
		if err != nil {
			Debug("init db error %v", err)
		}
	} else {
		env.cache = packet.NewCache(expired, nil)
	}

	// 调用启动前逻辑
	if onStartup != nil {
		onStartup()
	}

	// 初始化处理器
	env.chains = []chain{
		&http{},
		&env.rpc,
		&websocket{},
		&env.registry,
		&closer{},
		&reloader{},
		&uploader{},
	}
	for i := 0; i < len(env.chains); i++ {
		env.chains[i].Init()
	}

	// 处理请求
	var conn net.Conn
	for {
		conn, err = env.lsr.Accept()
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			break
		}
		go func(conn net.Conn) {
			defer func() {
				conn.Close()
				err := recover()
				if err == nil {
					return
				}
				pack := packet.New(1024)
				buf := pack.Allocate(1024)
				buf = buf[:runtime.Stack(buf, false)]
				Debug("\nprocess-conn error: %v\n%s\n\n", err, buf)
				packet.Free(pack)
			}()
			processConn(conn)
		}(conn)
	}

	// 清理资源
	env.lsr.Close()
	for i := 0; i < len(env.chains); i++ {
		env.chains[i].Close()
	}
	store.Close()

	return errors.New("service is down")
}

type bisDpo func(dpo Dpo) (resp interface{}, errCode string)

// Register 注册业务接口
func Register(api string, df bisDpo) {
	const errUNKNOWN = `Unknown`

	env.bis[api] = func(dpo Dpo) (resp interface{}, errCode string) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			pack := packet.New(1024)
			buf := pack.Allocate(1024)
			buf = buf[:runtime.Stack(buf, false)]
			Debug("\nhande [%s] error: %v\n%s\n\n", api, err, buf)
			packet.Free(pack)
			errCode = errUNKNOWN
		}()
		resp, errCode = df(dpo)
		return
	}
}

// findBis 查找业务
func findBis(api string) (bisDpo, bool) {
	df, ok := env.bis[api]
	return df, ok
}

// processConn 处理请求
func processConn(conn net.Conn) {
	const TIMEOUT = time.Second * 3

	pack := packet.New(2048)
	pack.SetTimeout(TIMEOUT, TIMEOUT)
	if err := pack.ReadHTTPHeader(conn); err != nil {
		packet.Free(pack)
		return
	}

	// 协议升级类型
	upgrade := pack.HTTPHeaderValue(httpUpgrade)

	// 处理请求
	for i := 0; i < len(env.chains); i++ {
		if env.chains[i].Handle(conn, upgrade, pack) {
			break
		}
	}

	// 释放数据包
	packet.Free(pack)
}

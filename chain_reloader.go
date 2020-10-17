package micro

import (
	"errors"
	"net"
	"runtime"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type reloader struct {
	baseChain
}

// Handle 处理Conn
func (c *reloader) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	if name != "reloader" {
		return false
	}

	auc := pack.HTTPHeaderValue(httpAuthorize)
	pack.BeginWrite()
	if _, ok := env.authorize.Check(auc); ok {
		for _, m := range env.chains {
			m.Reload()
		}
		for _, f := range env.reloadFunc {
			f()
		}
		pack.WriteString(`service has been reloaded.`)
	} else {
		pack.WriteString(`trespassing`)
	}
	pack.EndWrite()
	pack.FlushToConn(conn)

	return true
}

// RegisterReloadFunc 注册重载接口
func RegisterReloadFunc(f func()) {
	env.reloadFunc = append(env.reloadFunc, func() {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			pack := packet.New(1024)
			buf := pack.Allocate(1024)
			buf = buf[:runtime.Stack(buf, false)]
			Debug("\nreload error: %v\n%s\n\n", err, buf)
		}()
		f()
	})
}

// 请求重载服务配置
func requestReloadService() error {
	const TIMEOUT = time.Second * 3

	conn, err := net.DialTimeout("tcp", localeAddress(), time.Second)
	if err != nil {
		return errors.New("service not found, it may be closed")
	}

	pack := packet.New(512)
	pack.SetTimeout(TIMEOUT, TIMEOUT)

	// 发送请求
	pack.Write([]byte("Upgrade: reloader"))
	pack.Write(httpAuthorize)
	pack.Write(xutils.UnsafeStringToBytes(env.authorize.NewCode("")))
	pack.Write(httpRowAt)
	pack.Write(httpRowAt)
	if _, err = pack.FlushToConn(conn); err != nil {
		packet.Free(pack)
		conn.Close()
		return errors.New("signal couldn't be sent. service may be closed")
	}

	// 接收数据
	err = pack.ReadConn(conn)
	conn.Close()
	if err != nil {
		return errors.New("signal cannot be received. service may be closed")
	}
	err = errors.New(pack.ReadString())
	packet.Free(pack)
	return err
}

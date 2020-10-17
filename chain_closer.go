package micro

import (
	"errors"
	"net"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

type closer struct {
	baseChain
}

// Handle 处理Conn
func (c *closer) Handle(conn net.Conn, name string, pack *packet.Packet) bool {
	if name != "closer" {
		return false
	}

	auc := pack.HTTPHeaderValue(httpAuthorize)
	pack.BeginWrite()
	if _, ok := env.authorize.Check(auc); ok {
		if env.lsr != nil {
			env.lsr.Close()
		}
		pack.WriteString(`service has been closed.`)
	} else {
		pack.WriteString(`trespassing`)
	}
	pack.EndWrite()
	pack.FlushToConn(conn)

	return true
}

// 请求关闭服务
func requestCloseService() error {
	const TIMEOUT = time.Second * 3

	conn, err := net.DialTimeout("tcp", localeAddress(), time.Second)
	if err != nil {
		return errors.New("service not found, it may be closed")
	}

	pack := packet.New(512)
	pack.SetTimeout(TIMEOUT, TIMEOUT)

	// 发送请求
	pack.Write([]byte("Upgrade: closer"))
	pack.Write(httpRowAt)
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

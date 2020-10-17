package micro

import (
	"net"

	"github.com/micro/packet"
)

// chain
type chain interface {
	Init()
	Handle(net.Conn, string, *packet.Packet) bool
	SendData(interface{}, string, []string)
	SendGroup(interface{}, string, uint8, string)
	Reload()
	Close()
}

type baseChain struct{}

func (c *baseChain) Init() {}
func (c *baseChain) Handle(conn net.Conn, upgrade string, pack *packet.Packet) bool {
	return false
}
func (c *baseChain) SendData(data interface{}, api string, uids []string)             {}
func (c *baseChain) SendGroup(data interface{}, api string, flag uint8, group string) {}
func (c *baseChain) Reload()                                                          {}
func (c *baseChain) Close()                                                           {}

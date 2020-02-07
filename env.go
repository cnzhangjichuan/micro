package micro

import (
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"net"
)

// setter logger
func SetLogger(log types.Logger) {
	env.log = log
}

// log for error
func Error(fmt string, v ...interface{}) {
	if env.log != nil {
		env.log.Error(fmt, v...)
	}
}

// log for log
func Log(fmt string, v ...interface{}) {
	if env.log != nil {
		env.log.Log(fmt, v...)
	}
}

var env struct {
	id       string
	port     string
	address  string
	handlers its.Handlers
	log      types.Logger
}

func init() {
	env.handlers = make(its.Handlers)
}

func initEnv(config *types.EnvConfig) error {
	env.id = config.Id
	env.port = config.Port
	if env.log == nil {
		SetLogger(its.NewDefaultLogger())
	}

	// set address
	ifs, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, ifa := range ifs {
		if ifa.Flags&net.FlagUp == 0 {
			continue
		}
		if ifa.Flags&net.FlagLoopback != 0 {
			continue
		}
		ads, err := ifa.Addrs()
		if err != nil {
			return err
		}

		var ip net.IP
		for _, adr := range ads {
			switch v := adr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue
		}
		env.address = ip.String() + ":" + env.port
		break
	}
	return nil
}

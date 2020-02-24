package micro

import (
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/internal/xhttp"
	"github.com/cnzhangjichuan/micro/internal/xrpc"
	"github.com/cnzhangjichuan/micro/internal/xwsk"
	"github.com/cnzhangjichuan/micro/types"
	"github.com/cnzhangjichuan/micro/xutils"
	"net/http"
	"runtime"
	"time"
)

// server
func Service(config types.EnvConfig) error {
	const (
		DefaultReadTimeout  = time.Second * 10
		DefaultWriteTimeout = time.Second * 20
	)

	// init env
	if err := initEnv(&config); err != nil {
		return err
	}
	Log("Service[%s] started on %s", Id(), Address())

	// init http
	xhttp.InitEnv(&config, env.handlers, env.log)

	// init rpc
	xrpc.InitEnv(&config, env.handlers, env.log)

	// init web-socket
	xwsk.InitEnv(&config, env.handlers, env.log)

	if config.ReadTimeout == 0 {
		config.ReadTimeout = DefaultReadTimeout
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = DefaultWriteTimeout
	}

	srv := http.Server{
		Addr:         ":" + env.port,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		Handler:      &dispatcherHandler{},
	}

	return srv.ListenAndServe()
}

// register
func Register(api string, permit string, f func(types.Dpo) error) {
	env.handlers[api] = &its.Handler{
		Permit: permit,
		Func: func(dpo types.Dpo) (err error) {
			defer func() {
				if ems := recover(); ems != nil {
					if env.log != nil {
						buf := make([]byte, 1024)
						buf = buf[:runtime.Stack(buf, false)]
						Error("hande [%s] error: %v\n%s\n\n", api, ems, buf)
					}
					err = xutils.NewError(ems)
				}
			}()
			err = f(dpo)
			return
		},
	}
}

// service id
func Id() string {
	return env.id
}

// service address
func Address() string {
	return env.address
}

// ====================================================================================================
// api for rpc.

// load data
func Load(result interface{}, id, api string, request interface{}) error {
	return xrpc.Load(result, id, api, request)
}

// load data
func LoadByAddress(result interface{}, id, address, api string, request interface{}) error {
	return xrpc.LoadByAddress(result, id, address, api, request)
}

// ====================================================================================================
// api for web-socket.

// SendMessage
func SendMessage(message interface{}, userId ...string) {
	xwsk.SendMessage(message, userId...)
}

// SendRoomMessage
func SendRoomMessage(message interface{}, room string) {
	xwsk.SendRoomMessage(message, room)
}

// ====================================================================================================
// api for event-source

// push event-source
func PushEventSource(v interface{}, users ...string) bool {
	return xhttp.PushEventSource(v, users...)
}

// close pusher
func CloseEventSource(uid string) {
	xhttp.CloseEventSource(uid)
}

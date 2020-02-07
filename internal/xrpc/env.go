package xrpc

import (
	"bytes"
	"errors"
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"math/rand"
)

var env struct {
	id       string
	port     string
	handlers its.Handlers
	log      types.Logger
	address  string
	aut      authorization
	csg      clientsGroup
	sia      staIdAddress
	scp      staConnPool
}

const (
	xPREFIX_ERR  = 0
	xPREFIX_DATA = 1
)

// init env
func InitEnv(config *types.EnvConfig, handlers its.Handlers, logger types.Logger) {
	const DefaultMask = `MyMicroMask`

	env.id = config.Id
	env.port = config.Port
	env.handlers = handlers
	env.log = logger

	msk := config.Mask
	if msk == "" {
		msk = DefaultMask
	}
	env.aut.mask = []byte(msk)
	env.aut.errServiceNotFound = errors.New(`not found service.`)
	env.aut.errServiceClosed = errors.New(`service shutdown.`)
	env.aut.errServiceAuthFailed = errors.New(`authorization fail.`)

	env.csg.climap = make(map[string]*clients)
	env.sia.stamap = make(map[string]string)

	if config.Register != "" {
		// Register mine to the central server
		go syncState(config.Register)
	}
}

// log for error
func logError(fmt string, v ...interface{}) {
	if env.log != nil {
		env.log.Error(fmt, v...)
	}
}

// log for info
func logInfo(fmt string, v ...interface{}) {
	if env.log != nil {
		env.log.Log(fmt, v...)
	}
}

type authorization struct {
	mask []byte

	errServiceNotFound   error
	errServiceClosed     error
	errServiceAuthFailed error
}

// create authorization code.
func (a *authorization) Create() []byte {
	ac := make([]byte, 16)

	for i := 0; i < 16; i++ {
		ac[i] = 'A' + byte(rand.Intn(26))
	}
	return ac
}

// encoding authorization code.
func (a *authorization) Encoding(ac []byte) []byte {
	ret := make([]byte, len(ac))
	mkl := len(a.mask)

	for i := 0; i < len(ret); i++ {
		ret[i] = ac[i] + byte(i)
		ret[i] ^= a.mask[i%mkl]
	}
	return ret
}

// check authorization code.
func (a *authorization) Check(ac, acc []byte) bool {
	return bytes.Equal(a.Encoding(ac), acc)
}

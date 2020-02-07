package xrpc

import (
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"github.com/cnzhangjichuan/micro/xutils"
)

type rpcDpo struct {
	its.BaseDpo

	data []byte
	user types.User
	resp interface{}
}

func (r *rpcDpo) Request(v interface{}) error {
	return xutils.UnmarshalJson(r.data, v)
}

func (r *rpcDpo) Response(resp interface{}) {
	r.resp = resp
}

func (r *rpcDpo) GetUser() types.User {
	return r.user
}

func (r *rpcDpo) BindUser(u types.User) {
	r.user = u
}

func (r *rpcDpo) UnBindUser() {
	r.user = nil
}

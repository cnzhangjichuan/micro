package its

import "github.com/cnzhangjichuan/micro/types"

type Handler struct {
	Permit string
	Func   func(types.Dpo) error
}

type Handlers map[string]*Handler

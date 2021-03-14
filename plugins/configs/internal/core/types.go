package core

import "sync"

type Service struct {
	m sync.RWMutex

	path string
}

const (
	typeString  = `string`
	typeSString = `sstring`
	typeCString = `cstring`
	typeFloat   = `float`
	typeSFloat  = `sfloat`
	typeCFloat  = `cfloat`
	typeInt     = `int`
	typeSInt    = `sint`
	typeCInt    = `cint`
	typeDate    = `date`
	typeSDate   = `sdate`
	typeCDate   = `cdate`
)

func (s *Service) Init() {
	s.path = `./configs.jd`
}

// isClientField 是否是客户端字段
func (s *Service) isClientField(typ string) bool {
	switch typ {
	default:
		return false
	case typeString, typeCString, typeFloat, typeCFloat, typeInt, typeCInt, typeDate, typeCDate:
		return true
	}
}

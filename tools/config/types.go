package config

import (
	"io"
	"sync"
)

var exc = excel{
	path: `./configs.jd`,
}

// Load 加载配置数据
func Load(name string, onSetup OnSetupFunc) error {
	return exc.Load(name, onSetup)
}

// Save 保存本配置数据
func Save(name string, r io.Reader) error {
	return exc.Save(name, r)
}

// ToLua 转成Lua文件
func ToLua(src, dst string) error {
	return exc.ToLua(src, dst)
}

// ToJSON 转成JSON文件
func ToJSON(src, dst string) error {
	return exc.ToJSON(src, dst)
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

type excel struct {
	sync.RWMutex
	path string
}

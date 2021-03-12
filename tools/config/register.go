package config

import (
	"io"

	"github.com/micro"
)

// Register 注册数据处理接口
func Register(name string, inf func(int) interface{}, enf func()) {
	Load(name, inf)
	if enf != nil {
		enf()
	}
	micro.RegisterUploadFunc(name, func(r io.Reader) error {
		err := Save(name, r)
		if err == nil {
			Load(name, inf)
			if enf != nil {
				enf()
			}
		}
		return err
	})
}

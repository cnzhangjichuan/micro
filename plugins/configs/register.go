package configs

import (
	"io"

	"github.com/micro"
	"github.com/micro/plugins/configs/internal/core"
)

// Register 注册数据管理模块
func Register() *manager {
	var m manager
	m.srv.Init()

	return &m
}

// manager 数据管理
type manager struct {
	srv core.Service
}

// Register 实现micro.Configs接口
func (m *manager) Register(name string, instance func(int) interface{}, init func()) {
	// 加载数据
	m.srv.Load(name, instance)
	if init != nil {
		init()
	}

	// 注册更新接口
	micro.RegisterUploadFunc(name, func(r io.Reader) error {
		err := m.srv.Save(name, r)
		if err == nil {
			m.srv.Load(name, instance)
			if init != nil {
				init()
			}
		}
		return err
	})
}

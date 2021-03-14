package micro

// Configs 配置管理
type Configs interface {
	// Register 注册数据管理模块
	// name 数据模块名称
	// instance 模块实例方法
	// init 数据加载完成后, 对数据进行后续处理
	Register(name string, instance func(size int) interface{}, init func())
}

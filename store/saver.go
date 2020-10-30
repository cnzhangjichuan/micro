package store

// NewSaver 创建回载器
func NewSaver(name string) *saver {
	return &saver{name: name}
}

type saver struct {
	name string
}

// Save 将数据保存到指定的数据表中
func (s *saver) Save(data interface{}, id string) {
	if s == nil || s.name == "" {
		return
	}
	Save(data, s.name, id)
}

// Load 从数据表中加载数据
func (s *saver) Load(data interface{}, id string) bool {
	if s == nil || s.name == "" {
		return false
	}
	return Load(data, s.name, id)
}

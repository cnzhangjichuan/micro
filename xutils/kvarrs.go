package xutils

import "strings"

// SS 字符串数组
type SS struct {
	ss []string
}

// Parse 解析数据
func (s *SS) Parse(nv string) {
	if strings.Index(nv, ";") >= 0 {
		s.ss = strings.Split(nv, ";")
	} else {
		s.ss = strings.Split(nv, ",")
	}
}

// Get 获取数据
func (s *SS) Get(x int32) (string, bool) {
	i := int(x)
	if i < 0 || i >= len(s.ss) {
		return "", false
	}
	return s.ss[i], true
}

// All 返回所有数据
func (s *SS) All() []string {
	return s.ss
}

// II32 整型数组
type II32 struct {
	ii []int32
}

// Parse 解析数据
func (s *II32) Parse(nv string) {
	if strings.Index(nv, ";") >= 0 {
		for _, v := range strings.Split(nv, ";") {
			s.ii = append(s.ii, ParseI32(v, 0))
		}
	} else {
		for _, v := range strings.Split(nv, ",") {
			s.ii = append(s.ii, ParseI32(v, 0))
		}
	}
}

// Get 获取数据
func (s *II32) Get(x int32) (int32, bool) {
	i := int(x)
	if i < 0 || i >= len(s.ii) {
		return 0, false
	}
	return s.ii[i], true
}

// All 返回所有数据
func (s *II32) All() []int32 {
	return s.ii
}

// Awards 奖励
type Awards struct {
	vv []SI32
}

// Parse 解析数据
func (s *Awards) Parse(nv string) {
	if nv == "" {
		return
	}
	ss := strings.Split(nv, ";")
	s.vv = make([]SI32, len(ss))
	for x, v := range ss {
		s.vv[x].Parse(v)
	}
}

// Get 获取奖励
func (s *Awards) Get(x int32) *SI32 {
	i := int(x - 1)
	if i < 0 || i >= len(s.vv) {
		return nil
	}
	return &s.vv[i]
}

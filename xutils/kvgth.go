package xutils

import (
	"strings"
)

// Growth 成长值
type Growth struct {
	f32 []float32
}

// Parse 解析数据
func (s *Growth) Parse(nv string) {
	var ss []string
	if strings.Index(nv, ";") >= 0 {
		ss = strings.Split(nv, ";")
	} else {
		ss = strings.Split(nv, ",")
	}
	s.f32 = make([]float32, len(ss))
	for i := 0; i < len(ss); i++ {
		s.f32[i] = ParseF32(ss[i], 1)
	}
}

// Get 获取指定位置的数据
func (s *Growth) Get(x int32) float32 {
	l := len(s.f32)
	if l == 0 {
		return 1
	}
	i := int(x - 1)
	if i < 0 {
		return s.f32[0]
	}
	if i >= l {
		return s.f32[l-1]
	}
	return s.f32[i]
}

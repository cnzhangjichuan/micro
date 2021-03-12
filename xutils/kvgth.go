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
	if x < 0 || x >= int32(len(s.f32)) {
		return 1
	}
	return s.f32[x]
}

package xutils

import (
	"strings"
)

// SF32 string/float32值对
type SF32 struct {
	str []string
	f32 []float32
}

// Parse 解析数据
func (s *SF32) Parse(nv string) {
	if nv == "" {
		s.str = make([]string, 0)
		s.f32 = make([]float32, 0)
		return
	}

	if strings.Index(nv, ";") >= 0 {
		ss := strings.Split(nv, ";")
		s.str = make([]string, len(ss))
		s.f32 = make([]float32, len(ss))
		for x, v := range ss {
			in := strings.Split(v, ",")
			s.str[x] = in[0]
			if len(in) > 1 {
				s.f32[x] = ParseF32(in[1], 0)
			}
		}
	} else {
		ss := strings.Split(nv, ",")
		l := len(ss)
		sf := l / 2
		s.str = make([]string, 0, sf)
		s.f32 = make([]float32, 0, sf)
		for i := 1; i < l; i += 2 {
			s.str = append(s.str, ss[i-1])
			s.f32 = append(s.f32, ParseF32(ss[i], 0))
		}
	}
}

// Merge 添加数据
func (s *SF32) Merge(n string, f float32) {
	for x := 0; x < len(s.str); x++ {
		if s.str[x] == n {
			s.f32[x] += f
			return
		}
	}
	s.Add(n, f)
}

// MergeAll 添加数据
func (s *SF32) MergeAll(f *SF32) {
	for x := 0; x < f.Len(); x++ {
		n, v := f.Get(x)
		s.Merge(n, v)
	}
}

// Add 添加数据
func (s *SF32) Add(n string, f float32) {
	s.str = append(s.str, n)
	s.f32 = append(s.f32, f)
}

// AddAll 添加数据
func (s *SF32) AddAll(f *SF32) {
	s.str = append(s.str, f.str...)
	s.f32 = append(s.f32, f.f32...)
}

// Len 获取数据长度
func (s *SF32) Len() int {
	return len(s.str)
}

// Get 获取指定位置的数据
func (s *SF32) Get(x int) (string, float32) {
	if x < 0 || x >= len(s.str) {
		return "", 0
	}
	return s.str[x], s.f32[x]
}

// Set 设置指定位置的数据
func (s *SF32) Set(x int, v float32) {
	if 0 <= x && x < len(s.f32) {
		s.f32[x] = v
	}
}

// GetValue 获取指定名称的值
func (s *SF32) GetValue(name string) float32 {
	for i, n := range s.str {
		if n == name {
			return s.f32[i]
		}
	}
	return 0
}

// Reset 重置
func (s *SF32) Reset() {
	if len(s.str) > 0 {
		s.str = s.str[:0]
	}
	if len(s.f32) > 0 {
		s.f32 = s.f32[:0]
	}
}

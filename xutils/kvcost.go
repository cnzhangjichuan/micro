package xutils

import "strings"

// Cost 消耗
type Cost struct {
	str []string
	i32 []int32
}

// Parse 解析数据
func (s *Cost) Parse(nv string) {
	if nv == "" {
		s.str = make([]string, 0)
		s.i32 = make([]int32, 0)
		return
	}

	if strings.Index(nv, ";") >= 0 {
		ss := strings.Split(nv, ";")
		s.str = make([]string, len(ss))
		s.i32 = make([]int32, len(ss))
		for x, v := range ss {
			in := strings.Split(v, ",")
			s.str[x] = in[0]
			if len(in) > 1 {
				s.i32[x] = -ParseI32(in[1], 0)
			}
		}
	} else {
		ss := strings.Split(nv, ",")
		l := len(ss)
		si := l / 2
		s.str = make([]string, 0, si)
		s.i32 = make([]int32, 0, si)
		for i := 1; i < l; i += 2 {
			s.str = append(s.str, ss[i-1])
			s.i32 = append(s.i32, -ParseI32(ss[i], 0))
		}
	}
}

// Len 获取数据长度
func (s *Cost) Len() int {
	return len(s.str)
}

// Get 获取指定位置的数据
func (s *Cost) Get(x int) (string, int32) {
	if x < 0 || x >= len(s.str) {
		return "", 0
	}
	return s.str[x], s.i32[x]
}

// SI32 转化成SI32
func (s *Cost) SI32() SI32 {
	return SI32{
		str: s.str,
		i32: s.i32,
	}
}

// Costs 多段消耗
type Costs struct {
	cc []Cost
}

// Parse 解析数据
func (s *Costs) Parse(nv string) {
	if nv == "" {
		return
	}
	ss := strings.Split(nv, ";")
	s.cc = make([]Cost, len(ss))
	for x, v := range ss {
		s.cc[x].Parse(v)
	}
}

// Get 获取消耗
func (s *Costs) Get(x int32) *Cost {
	i := int(x - 1)
	if i < 0 || i >= len(s.cc) {
		return nil
	}
	return &s.cc[i]
}

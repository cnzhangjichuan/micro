package xutils

import (
	"strings"
)

// I32F32 次数/比例值对
type I32F32 struct {
	i32 []int32
	f32 []float32
}

// Parse 解析数据
func (s *I32F32) Parse(nv string) {
	if nv == "" {
		s.i32 = make([]int32, 0)
		s.f32 = make([]float32, 0)
		return
	}
	if strings.Index(nv, ";") >= 0 {
		ss := strings.Split(nv, ";")
		s.i32 = make([]int32, len(ss))
		s.f32 = make([]float32, len(ss))
		for x, v := range ss {
			in := strings.Split(v, ",")
			s.i32[x] = ParseI32(in[0], 0)
			if len(in) > 1 {
				s.f32[x] = ParseF32(in[1], 0)
			}
		}
	} else {
		ss := strings.Split(nv, ",")
		l := len(ss) / 2
		s.i32 = make([]int32, 0, l)
		s.f32 = make([]float32, 0, l)
		for i := 1; i < len(ss); i += 2 {
			s.i32 = append(s.i32, ParseI32(ss[i-1], 0))
			s.f32 = append(s.f32, ParseF32(ss[i], 0))
		}
	}
}

// Add 添加数据
func (s *I32F32) Add(i int32, f float32) {
	s.i32 = append(s.i32, i)
	s.f32 = append(s.f32, f)
}

// Len 获取数据长度
func (s *I32F32) Len() int {
	return len(s.i32)
}

// Get 获取指定位置的数据
func (s *I32F32) Get(x int) (int32, float32) {
	if x < 0 || x >= len(s.i32) {
		return 0, 0
	}
	return s.i32[x], s.f32[x]
}

// GetRate 获取指定次数的概率
// times 从1次开始
func (s *I32F32) GetTimesRate(times int32) float32 {
	// 没有找到匹配项
	// 取最接近的数据返回
	var (
		retTimes int32
		retRate  float32
	)
	for i, ts := range s.i32 {
		if times == ts {
			return s.f32[i]
		}
		if times > ts {
			rt, rr := ts, s.f32[i]
			if rt > retTimes {
				retTimes, retRate = rt, rr
			}
		}
	}
	return retRate
}

// SFF32
type SFF32 struct {
	s string
	v []float32
}

func (s *SFF32) Parse(nvv string) {
	if nvv == "" {
		return
	}
	var ss []string
	if strings.Index(nvv, ";") >= 0 {
		ss = strings.Split(nvv, ";")
	} else {
		ss = strings.Split(nvv, ",")
	}
	if len(ss) < 1 {
		return
	}
	s.v = make([]float32, 0, len(ss)-1)
	for x, v := range ss {
		if x == 0 {
			s.s = v
		} else {
			s.v = append(s.v, ParseF32(v, 0))
		}
	}
}

func (s *SFF32) Load(name string, vv ...*float32) bool {
	if s.s != name {
		return false
	}
	for i, v := range vv {
		if i >= len(s.v) {
			break
		}
		*v = *v + s.v[i]
	}
	return true
}

func (s *SFF32) Ratio(name string, ratio float32) bool {
	if s.s != name {
		return false
	}
	for i, v := range s.v {
		s.v[i] = v * ratio
	}
	return true
}

func (s *SFF32) GetName() string {
	return s.s
}

// I32S
type I32S struct {
	v []int32
}

func (s *I32S) Parse(nvv string) {
	if nvv == "" {
		return
	}
	var ss []string
	if strings.Index(nvv, ";") >= 0 {
		ss = strings.Split(nvv, ";")
	} else {
		ss = strings.Split(nvv, ",")
	}
	if len(ss) < 1 {
		return
	}
	s.v = make([]int32, 0, len(ss))
	for _, v := range ss {
		s.v = append(s.v, ParseI32(v, 0))
	}
}

func (s *I32S) Get(x int) (int32, bool) {
	l := len(s.v)
	if l == 0 {
		return 0, false
	}
	if x < 0 {
		return s.v[0], false
	}
	ok := true
	if x >= l {
		x = l - 1
		ok = false
	}
	return s.v[x], ok
}

func (s *I32S) Len() int {
	return len(s.v)
}

// F32S
type F32S struct {
	v []float32
}

func (s *F32S) Parse(nvv string) {
	if nvv == "" {
		return
	}
	var ss []string
	if strings.Index(nvv, ";") >= 0 {
		ss = strings.Split(nvv, ";")
	} else {
		ss = strings.Split(nvv, ",")
	}
	if len(ss) < 1 {
		return
	}
	s.v = make([]float32, 0, len(ss))
	for _, v := range ss {
		s.v = append(s.v, ParseF32(v, 0))
	}
}

func (s *F32S) Get(x int) (float32, bool) {
	l := len(s.v)
	if l == 0 {
		return 0, false
	}
	if x < 0 {
		return s.v[0], false
	}

	ok := true
	if x >= l {
		x = l - 1
		ok = false
	}
	return s.v[x], ok
}

func (s *F32S) Len() int {
	return len(s.v)
}

// SS
type SS struct {
	v []string
}

func (s *SS) Parse(nvv string) {
	if nvv == "" {
		return
	}
	if strings.Index(nvv, ";") >= 0 {
		s.v = strings.Split(nvv, ";")
	} else {
		s.v = strings.Split(nvv, ",")
	}
}

func (s *SS) Get(x int) (string, bool) {
	l := len(s.v)
	if l == 0 {
		return "", false
	}
	if x < 0 {
		return s.v[0], false
	}

	ok := true
	if x >= l {
		x = l - 1
		ok = false
	}
	return s.v[x], ok
}

func (s *SS) GetValues() []string {
	return s.v
}

func (s *SS) Len() int {
	return len(s.v)
}

package xutils

import "strings"

// SS 字符串数组
type SS struct {
	v []string
}

// Parse 解析数据
func (s *SS) Parse(nv string) {
	if strings.Index(nv, ";") >= 0 {
		s.v = strings.Split(nv, ";")
	} else {
		s.v = strings.Split(nv, ",")
	}
}

// Get 获取数据
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

// GetValues 返回所有数据
func (s *SS) GetValues() []string {
	return s.v
}

// HasAny 是否包含任一元素
func (s *SS) HasAny(ss []string) bool {
	for _, v := range ss {
		if HasString(s.v, v) {
			return true
		}
	}
	return false
}

func (s *SS) Len() int {
	return len(s.v)
}

// II32 整型数组
type II32 struct {
	v []int32
}

// Parse 解析数据
func (s *II32) Parse(nv string) {
	if strings.Index(nv, ";") >= 0 {
		for _, v := range strings.Split(nv, ";") {
			s.v = append(s.v, ParseI32(v, 0))
		}
	} else {
		for _, v := range strings.Split(nv, ",") {
			s.v = append(s.v, ParseI32(v, 0))
		}
	}
}

// Get 获取数据
func (s *II32) Get(x int) (int32, bool) {
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

// GetValues 返回所有数据
func (s *II32) GetValues() []int32 {
	return s.v
}

func (s *II32) Len() int {
	return len(s.v)
}

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

// I64S
type I64S struct {
	v []int64
}

func (s *I64S) Parse(nvv string) {
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
	s.v = make([]int64, 0, len(ss))
	for _, v := range ss {
		s.v = append(s.v, ParseI64(v, 0))
	}
}

func (s *I64S) Get(x int) (int64, bool) {
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

func (s *I64S) Len() int {
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

// FF32 整型数组
type FF32 struct {
	v []float32
}

// Parse 解析数据
func (s *FF32) Parse(nv string) {
	if strings.Index(nv, ";") >= 0 {
		for _, v := range strings.Split(nv, ";") {
			s.v = append(s.v, ParseF32(v, 0))
		}
	} else {
		for _, v := range strings.Split(nv, ",") {
			s.v = append(s.v, ParseF32(v, 0))
		}
	}
}

// Get 获取数据
func (s *FF32) Get(x int) (float32, bool) {
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

// GetValues 返回所有数据
func (s *FF32) GetValues() []float32 {
	return s.v
}

// Sum 反回总和
func (s *FF32) Sum() (ret float32) {
	for _, v := range s.v {
		ret += v
	}
	return
}

func (s *FF32) Len() int {
	return len(s.v)
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

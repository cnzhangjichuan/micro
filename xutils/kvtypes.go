package xutils

import (
	"strconv"
	"strings"
)

// SI32 string/int32值对
type SI32 struct {
	str []string
	i32 []int32
}

// 初始化
func (s *SI32) Init(n string, i int32) {
	if len(s.str) > 0 {
		s.str = s.str[:1]
		s.str[0] = n
		s.i32 = s.i32[:1]
		s.i32[0] = i
	} else {
		s.Add(n, i)
	}
}

// Ratio 将数量折上一定的系数
func (s *SI32) Ratio(r float32) (cpy SI32, ok bool) {
	for x := 0; x < s.Len(); x++ {
		n, v := s.Get(x)
		v = int32(float32(v) * r)
		if v >= 1 || v <= -1 {
			ok = true
			cpy.Add(n, v)
		}
	}
	return
}

// Ratio 将数量折上一定的系数
func (s *SI32) RatioOnly(r float32, nn string) (cpy SI32, ok bool) {
	for x := 0; x < s.Len(); x++ {
		n, v := s.Get(x)
		if n != nn {
			ok = true
			cpy.Add(n, v)
		} else {
			v = int32(float32(v) * r)
			if v >= 1 || v <= -1 {
				ok = true
				cpy.Add(n, v)
			}
		}
	}
	return
}

// Add 添加数据
func (s *SI32) Add(n string, i int32) {
	s.str = append(s.str, n)
	s.i32 = append(s.i32, i)
}

// AddAll 添加数据
func (s *SI32) AddAll(i *SI32) {
	for x := 0; x < i.Len(); x++ {
		n, v := i.Get(x)
		s.Add(n, v)
	}
}

// Merge 添加数据
func (s *SI32) Merge(n string, i int32) {
	for x := 0; x < len(s.str); x++ {
		if s.str[x] == n {
			s.i32[x] += i
			return
		}
	}
	s.Add(n, i)
}

// MergeAll 添加数据
func (s *SI32) MergeAll(i *SI32) {
	for x := 0; x < i.Len(); x++ {
		n, v := i.Get(x)
		s.Merge(n, v)
	}
}

// 设置为消耗
func (s *SI32) EnableCost() {
	for i, v := range s.i32 {
		if v > 0 {
			s.i32[i] = 0 - v
		}
	}
}

// Parse 解析数据
func (s *SI32) Parse(nv string) {
	const def = 1

	if nv == "" {
		s.str = make([]string, 0)
		s.i32 = make([]int32, 0)
		return
	}

	if strings.Index(nv, ";") >= 0 {
		items := strings.Split(nv, ";")
		s.str = make([]string, len(items))
		s.i32 = make([]int32, len(items))
		for x, v := range items {
			in := strings.Split(v, ",")
			s.str[x] = in[0]
			if len(in) < 2 {
				s.i32 = append(s.i32, def)
				continue
			}
			v, err := strconv.ParseInt(in[1], 10, 32)
			if err != nil {
				s.i32[x] = def
			} else {
				s.i32[x] = int32(v)
			}
		}
	} else {
		items := strings.Split(nv, ",")
		l := len(items)
		s.str = make([]string, 0, l)
		s.i32 = make([]int32, 0, l)
		i := 0
		for i < l {
			s.str = append(s.str, items[i])
			i += 1
			if i >= l {
				s.i32 = append(s.i32, def)
				break
			}
			v, err := strconv.ParseInt(items[i], 10, 32)
			if err != nil {
				s.i32 = append(s.i32, def)
				continue
			}
			i += 1
			s.i32 = append(s.i32, int32(v))
		}
	}
}

// Append 解析数据
func (s *SI32) Append(nv string) {
	const def = 1

	if nv == "" {
		return
	}

	// split by ';'
	if strings.Index(nv, ";") >= 0 {
		items := strings.Split(nv, ";")
		for _, v := range items {
			in := strings.Split(v, ",")
			s.str = append(s.str, in[0])
			if len(in) < 2 {
				s.i32 = append(s.i32, def)
				continue
			}
			v, err := strconv.ParseInt(in[1], 10, 32)
			if err != nil {
				s.i32 = append(s.i32, def)
			} else {
				s.i32 = append(s.i32, int32(v))
			}
		}
		return
	}

	// split by ','
	items := strings.Split(nv, ",")
	i := 0
	for i < len(items) {
		s.str = append(s.str, items[i])
		i += 1
		if i >= len(items) {
			s.i32 = append(s.i32, def)
			break
		}
		v, err := strconv.ParseInt(items[i], 10, 32)
		if err != nil {
			s.i32 = append(s.i32, def)
			continue
		}
		i += 1
		s.i32 = append(s.i32, int32(v))
	}
}

// Len 获取数据长度
func (s *SI32) Len() int {
	return len(s.str)
}

// Get 获取指定位置的数据
func (s *SI32) Get(x int) (string, int32) {
	if x < 0 || x >= len(s.str) {
		return "", 0
	}
	return s.str[x], s.i32[x]
}

// Set 设置指定位置的数据
func (s *SI32) Set(x int, v int32) {
	if 0 <= x && x < len(s.i32) {
		s.i32[x] = v
	}
}

// GetValue 获取指定名称的值
func (s *SI32) GetValue(name string) int32 {
	for i, n := range s.str {
		if n == name {
			return s.i32[i]
		}
	}
	return 0
}

// Reset 重置
func (s *SI32) Reset() {
	if len(s.str) > 0 {
		s.str = s.str[:0]
	}
	if len(s.i32) > 0 {
		s.i32 = s.i32[:0]
	}
}

// SF32 string/float32值对
type SF32 struct {
	str []string
	f32 []float32
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
	for x := 0; x < f.Len(); x++ {
		n, v := f.Get(x)
		s.Add(n, v)
	}
}

// Parse 解析数据
func (s *SF32) Parse(nv string) {
	if nv == "" {
		s.str = make([]string, 0)
		s.f32 = make([]float32, 0)
		return
	}

	if strings.Index(nv, ";") >= 0 {
		items := strings.Split(nv, ";")
		s.str = make([]string, len(items))
		s.f32 = make([]float32, len(items))
		for x, v := range items {
			in := strings.Split(v, ",")
			s.str[x] = in[0]
			if len(in) < 2 {
				s.f32[x] = 0
				continue
			}
			if v, err := strconv.ParseFloat(in[1], 32); err != nil {
				s.f32[x] = 0
			} else {
				s.f32[x] = float32(v)
			}
		}
	} else {
		items := strings.Split(nv, ",")
		l := len(items)
		s.str = make([]string, 0, l)
		s.f32 = make([]float32, 0, l)
		i := 0
		for i < l {
			s.str = append(s.str, items[i])
			i += 1
			if i >= l {
				s.f32 = append(s.f32, 0)
				break
			}
			v, err := strconv.ParseFloat(items[i], 32)
			if err != nil {
				s.f32 = append(s.f32, 0)
				continue
			}
			i += 1
			s.f32 = append(s.f32, float32(v))
		}
	}
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

// TimesRate 次数/比例值对
type I32F32 struct {
	i32 []int32
	f32 []float32
}

// Add 添加数据
func (s *I32F32) Add(i int32, f float32) {
	s.i32 = append(s.i32, i)
	s.f32 = append(s.f32, f)
}

// Parse 解析数据
func (s *I32F32) Parse(nv string) {
	if nv == "" {
		s.i32 = make([]int32, 0)
		s.f32 = make([]float32, 0)
		return
	}
	if strings.Index(nv, ";") >= 0 {
		items := strings.Split(nv, ";")
		s.i32 = make([]int32, len(items))
		s.f32 = make([]float32, len(items))
		for x, v := range items {
			in := strings.Split(v, ",")
			iv, err := strconv.ParseInt(in[0], 10, 32)
			if err != nil {
				s.i32[x] = 0
			} else {
				s.i32[x] = int32(iv)
			}
			if len(in) < 2 {
				s.f32[x] = 0
				continue
			}
			fv, err := strconv.ParseFloat(in[1], 32)
			if err != nil {
				s.f32[x] = 0
			} else {
				s.f32[x] = float32(fv)
			}
		}
	} else {
		items := strings.Split(nv, ",")
		l := len(items) / 2
		s.i32 = make([]int32, 0, l)
		s.f32 = make([]float32, 0, l)
		for i := 1; i < len(items); i += 2 {
			iv, err := strconv.ParseInt(items[i-1], 10, 32)
			if err != nil {
				s.i32 = append(s.i32, 0)
			} else {
				s.i32 = append(s.i32, int32(iv))
			}
			fv, err := strconv.ParseFloat(items[i], 32)
			if err != nil {
				s.f32 = append(s.f32, 0)
			} else {
				s.f32 = append(s.f32, float32(fv))
			}
		}
	}
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
		retTimes = int32(0)
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

// Growth 成长值
type Growth struct {
	f32 []float32
}

// Parse 解析数据
func (s *Growth) Parse(nv string) {
	var items []string
	if strings.Index(nv, ";") >= 0 {
		items = strings.Split(nv, ";")
	} else {
		items = strings.Split(nv, ",")
	}
	s.f32 = make([]float32, len(items))
	for i := 0; i < len(items); i++ {
		f64, err := strconv.ParseFloat(items[i], 32)
		if err != nil {
			f64 = 1
		}
		f32 := float32(f64)
		if i > 0 {
			f32 *= s.f32[i-1]
		}
		s.f32[i] = f32
	}
}

// Get 获取指定位置的数据
func (s *Growth) Get(x int32) float32 {
	if x < 0 || x >= int32(len(s.f32)) {
		return 1
	}
	return s.f32[x]
}

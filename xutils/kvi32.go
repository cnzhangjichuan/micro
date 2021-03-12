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

// Parse 解析数据
func (s *SI32) Parse(nv string) {
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
				s.i32[x] = ParseI32(in[1], 0)
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
			s.i32 = append(s.i32, ParseI32(ss[i], 0))
		}
	}
}

// String 返回字符串
func (s *SI32) String() string {
	l := s.Len()
	if l <= 0 {
		return ""
	}
	ss := make([]string, 0, 2*l)
	for x := 0; x < l; x++ {
		ss = append(ss, s.str[x])
		ss = append(ss, strconv.Itoa(int(s.i32[x])))
	}
	return strings.Join(ss, ",")
}

// Ratio 将数量折上一定的系数
func (s *SI32) Ratio(r float32) (cpy SI32, ok bool) {
	cpy.str = make([]string, s.Len())
	cpy.i32 = make([]int32, s.Len())
	for x := 0; x < s.Len(); x++ {
		cpy.str[x] = s.str[x]
		v := int32(float32(s.i32[x]) * r)
		if v >= 1 || v <= -1 {
			ok = true
			cpy.i32[x] = v
		} else {
			cpy.i32[x] = s.i32[x]
		}
	}
	return
}

// Ratio 将数量折上一定的系数
func (s *SI32) RatioOnly(r float32, nn string) (cpy SI32, ok bool) {
	cpy.str = make([]string, s.Len())
	cpy.i32 = make([]int32, s.Len())
	for x := 0; x < s.Len(); x++ {
		cpy.str[x] = s.str[x]
		if cpy.str[x] == nn {
			ok = true
			cpy.i32[x] = int32(float32(s.i32[x]) * r)
		} else {
			cpy.i32[x] = s.i32[x]
		}
	}
	return
}

// Ratio 将数量折上一定的系数
func (s *SI32) RatioThis(nn string, r float32) {
	for x := 0; x < s.Len(); x++ {
		if s.str[x] == nn {
			s.i32[x] = int32(float32(s.i32[x]) * r)
		}
	}
}

// Add 添加数据
func (s *SI32) Add(n string, i int32) {
	s.str = append(s.str, n)
	s.i32 = append(s.i32, i)
}

// AddAll 添加数据
func (s *SI32) AddAll(i *SI32) {
	s.str = append(s.str, i.str...)
	s.i32 = append(s.i32, i.i32...)
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

// Append 解析数据
func (s *SI32) Append(nv string) {
	if nv == "" {
		return
	}

	// split by ';'
	if strings.Index(nv, ";") >= 0 {
		ss := strings.Split(nv, ";")
		for _, v := range ss {
			in := strings.Split(v, ",")
			s.str = append(s.str, in[0])
			if len(in) > 1 {
				s.i32 = append(s.i32, ParseI32(in[1], 0))
			} else {
				s.i32 = append(s.i32, 0)
			}
		}
		return
	}

	// split by ','
	ss := strings.Split(nv, ",")
	for i := 1; i < len(ss); i++ {
		s.str = append(s.str, ss[i-1])
		s.i32 = append(s.i32, ParseI32(ss[i], 0))
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

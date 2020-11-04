package tools

import (
	"strings"

	"github.com/micro/xutils"
)

// SI32 string/int32值对
type SI32 struct {
	str []string
	i32 []int32
}

// Add 添加数据
func (mixed *SI32) Add(s string, i int32) {
	mixed.str = append(mixed.str, s)
	mixed.i32 = append(mixed.i32, i)
}

// Add 添加数据
func (mixed *SI32) Merge(s string, i int32) {
	for x := 0; x < len(mixed.str); x++ {
		if mixed.str[x] == s {
			mixed.i32[x] += i
			return
		}
	}
	mixed.Add(s, i)
}

// Parse 解析数据
func (mixed *SI32) Parse(s string) {
	if s == "" {
		mixed.str = make([]string, 0)
		mixed.i32 = make([]int32, 0)
		return
	}

	items := strings.Split(s, ";")
	mixed.str = make([]string, len(items))
	mixed.i32 = make([]int32, len(items))
	for x, v := range items {
		in := strings.Split(v, ",")
		mixed.str[x] = in[0]
		if len(in) > 1 {
			mixed.i32[x] = xutils.ParseI32(in[1], 0)
		} else {
			mixed.i32[x] = 0
		}
	}
}

// Len 获取数据长度
func (mixed *SI32) Len() int {
	return len(mixed.str)
}

// Get 获取指定位置的数据
func (mixed *SI32) Get(x int) (string, int32) {
	if x < 0 || x >= len(mixed.str) {
		return "", 0
	}
	return mixed.str[x], mixed.i32[x]
}

// Set 设置指定位置的数据
func (mixed *SI32) Set(x int, v int32) {
	if len(mixed.i32) > x {
		mixed.i32[x] = v
	}
}

// GetValue 获取指定名称的值
func (mixed *SI32) GetValue(name string) int32 {
	for i, n := range mixed.str {
		if n == name {
			return mixed.i32[i]
		}
	}
	return 0
}

// SF32 string/float32值对
type SF32 struct {
	str []string
	f32 []float32
}

// Add 添加数据
func (mixed *SF32) Add(s string, f float32) {
	mixed.str = append(mixed.str, s)
	mixed.f32 = append(mixed.f32, f)
}

// Parse 解析数据
func (mixed *SF32) Parse(s string) {
	if s == "" {
		mixed.str = make([]string, 0)
		mixed.f32 = make([]float32, 0)
		return
	}

	items := strings.Split(s, ";")
	mixed.str = make([]string, len(items))
	mixed.f32 = make([]float32, len(items))
	for x, v := range items {
		in := strings.Split(v, ",")
		mixed.str[x] = in[0]
		if len(in) > 1 {
			mixed.f32[x] = xutils.ParseF32(in[1], 0)
		} else {
			mixed.f32[x] = 0
		}
	}
}

// Len 获取数据长度
func (mixed *SF32) Len() int {
	return len(mixed.str)
}

// Get 获取指定位置的数据
func (mixed *SF32) Get(x int) (string, float32) {
	if x < 0 || x >= len(mixed.str) {
		return "", 0
	}
	return mixed.str[x], mixed.f32[x]
}

// Set 设置指定位置的数据
func (mixed *SF32) Set(x int, v float32) {
	if len(mixed.f32) > x {
		mixed.f32[x] = v
	}
}

// GetValue 获取指定名称的值
func (mixed *SF32) GetValue(name string) float32 {
	for i, n := range mixed.str {
		if n == name {
			return mixed.f32[i]
		}
	}
	return 0
}

// TimesRate 次数/比例值对
type I32F32 struct {
	i32 []int32
	f32 []float32
}

// Add 添加数据
func (mixed *I32F32) Add(i int32, f float32) {
	mixed.i32 = append(mixed.i32, i)
	mixed.f32 = append(mixed.f32, f)
}

// Parse 解析数据
func (mixed *I32F32) Parse(s string) {
	if s == "" {
		mixed.i32 = make([]int32, 0)
		mixed.f32 = make([]float32, 0)
		return
	}
	items := strings.Split(s, ";")
	mixed.i32 = make([]int32, len(items))
	mixed.f32 = make([]float32, len(items))
	for x, v := range items {
		in := strings.Split(v, ",")
		mixed.i32[x] = xutils.ParseI32(in[0], 1)
		if len(in) > 1 {
			mixed.f32[x] = xutils.ParseF32(in[1], 0)
		} else {
			mixed.f32[x] = 0
		}
	}
}

// Len 获取数据长度
func (mixed *I32F32) Len() int {
	return len(mixed.i32)
}

// Get 获取指定位置的数据
func (mixed *I32F32) Get(x int) (int32, float32) {
	if x < 0 || x >= len(mixed.i32) {
		return 0, 0
	}
	return mixed.i32[x], mixed.f32[x]
}

// GetRate 获取指定次数的概率
// times 从1次开始
func (t *I32F32) GetTimesRate(times int32) float32 {
	// 没有找到匹配项
	// 取最接近的数据返回
	var (
		retTimes = int32(0)
		retRate  float32
	)
	for i, ts := range t.i32 {
		if times == ts {
			return t.f32[i]
		}
		if times > ts {
			rt, rr := ts, t.f32[i]
			if rt > retTimes {
				retTimes, retRate = rt, rr
			}
		}
	}
	return retRate
}

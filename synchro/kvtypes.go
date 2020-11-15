package synchro

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
func (mixed *SI32) Init(s string, i int32) {
	if len(mixed.str) > 0 {
		mixed.str = mixed.str[:1]
		mixed.str[0] = s
		mixed.i32 = mixed.i32[:1]
		mixed.i32[0] = i
	} else {
		mixed.Add(s, i)
	}
}

// Add 添加数据
func (mixed *SI32) Add(s string, i int32) {
	mixed.str = append(mixed.str, s)
	mixed.i32 = append(mixed.i32, i)
}

// Merge 添加数据
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
	const def = 1

	if s == "" {
		mixed.str = make([]string, 0)
		mixed.i32 = make([]int32, 0)
		return
	}

	if strings.Index(s, ";") >= 0 {
		items := strings.Split(s, ";")
		mixed.str = make([]string, len(items))
		mixed.i32 = make([]int32, len(items))
		for x, v := range items {
			in := strings.Split(v, ",")
			mixed.str[x] = in[0]
			if len(in) < 2 {
				mixed.i32 = append(mixed.i32, def)
				continue
			}
			v, err := strconv.ParseInt(in[1], 10, 32)
			if err != nil {
				mixed.i32 = append(mixed.i32, def)
			} else {
				mixed.i32 = append(mixed.i32, int32(v))
			}
		}
	} else {
		items := strings.Split(s, ",")
		l := len(items)
		mixed.str = make([]string, 0, l)
		mixed.i32 = make([]int32, 0, l)
		i := 0
		for i < l {
			mixed.str = append(mixed.str, items[i])
			i += 1
			if i >= l {
				mixed.i32 = append(mixed.i32, def)
				break
			}
			v, err := strconv.ParseInt(items[i], 10, 32)
			if err != nil {
				mixed.i32 = append(mixed.i32, def)
				continue
			}
			i += 1
			mixed.i32 = append(mixed.i32, int32(v))
		}
	}
}

// Append 解析数据
func (mixed *SI32) Append(s string) {
	const def = 1

	if s == "" {
		return
	}
	if strings.Index(s, ";") >= 0 {
		items := strings.Split(s, ";")
		for _, v := range items {
			in := strings.Split(v, ",")
			mixed.str = append(mixed.str, in[0])
			if len(in) < 2 {
				mixed.i32 = append(mixed.i32, def)
				continue
			}
			v, err := strconv.ParseInt(in[1], 10, 32)
			if err != nil {
				mixed.i32 = append(mixed.i32, def)
			} else {
				mixed.i32 = append(mixed.i32, int32(v))
			}
		}
	} else {
		items := strings.Split(s, ",")
		i := 0
		for i < len(items) {
			mixed.str = append(mixed.str, items[i])
			i += 1
			if i >= len(items) {
				mixed.i32 = append(mixed.i32, def)
				break
			}
			v, err := strconv.ParseInt(items[i], 10, 32)
			if err != nil {
				mixed.i32 = append(mixed.i32, def)
				continue
			}
			i += 1
			mixed.i32 = append(mixed.i32, int32(v))
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

	if strings.Index(s, ";") >= 0 {
		items := strings.Split(s, ";")
		mixed.str = make([]string, len(items))
		mixed.f32 = make([]float32, len(items))
		for x, v := range items {
			in := strings.Split(v, ",")
			mixed.str[x] = in[0]
			if len(in) < 2 {
				mixed.f32[x] = 0
				continue
			}
			if v, err := strconv.ParseFloat(in[1], 32); err != nil {
				mixed.f32[x] = 0
			} else {
				mixed.f32[x] = float32(v)
			}
		}
	} else {
		items := strings.Split(s, ",")
		l := len(items)
		mixed.str = make([]string, 0, l)
		mixed.f32 = make([]float32, 0, l)
		i := 0
		for i < l {
			mixed.str = append(mixed.str, items[i])
			i += 1
			if i >= l {
				mixed.f32 = append(mixed.f32, 0)
				break
			}
			v, err := strconv.ParseFloat(items[i], 32)
			if err != nil {
				mixed.f32 = append(mixed.f32, 0)
				continue
			}
			i += 1
			mixed.f32 = append(mixed.f32, float32(v))
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
	if strings.Index(s, ";") >= 0 {
		items := strings.Split(s, ";")
		mixed.i32 = make([]int32, len(items))
		mixed.f32 = make([]float32, len(items))
		for x, v := range items {
			in := strings.Split(v, ",")
			iv, err := strconv.ParseInt(in[0], 10, 32)
			if err != nil {
				mixed.i32[x] = 0
			} else {
				mixed.i32[x] = int32(iv)
			}
			if len(in) < 2 {
				mixed.f32[x] = 0
				continue
			}
			fv, err := strconv.ParseFloat(in[1], 32)
			if err != nil {
				mixed.f32[x] = 0
			} else {
				mixed.f32[x] = float32(fv)
			}
		}
	} else {
		items := strings.Split(s, ",")
		l := len(items) / 2
		mixed.i32 = make([]int32, 0, l)
		mixed.f32 = make([]float32, 0, l)
		for i := 1; i < len(items); i += 2 {
			iv, err := strconv.ParseInt(items[i-1], 10, 32)
			if err != nil {
				mixed.i32 = append(mixed.i32, 0)
			} else {
				mixed.i32 = append(mixed.i32, int32(iv))
			}
			fv, err := strconv.ParseFloat(items[i], 32)
			if err != nil {
				mixed.f32 = append(mixed.f32, 0)
			} else {
				mixed.f32 = append(mixed.f32, float32(fv))
			}
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

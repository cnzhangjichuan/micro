package tools

import (
	"strings"

	"github.com/micro/xutils"
)

// IDNum Id/Num值对
type IDNum struct {
	Id  []string
	Num []int32
}

// Parse 解析数据
func (i *IDNum) Parse(s string) {
	if s == "" {
		i.Id = make([]string, 0)
		i.Num = make([]int32, 0)
		return
	}

	items := strings.Split(s, ";")
	i.Id = make([]string, len(items))
	i.Num = make([]int32, len(items))
	for x, v := range items {
		in := strings.Split(v, ",")
		i.Id[x] = in[0]
		if len(in) > 1 {
			i.Num[x] = xutils.ParseI32(in[1], 0)
		} else {
			i.Num[x] = 0
		}
	}
}

// Len 获取数据长度
func (i *IDNum) Len() int {
	return len(i.Id)
}

// Get 获取指定位置的数据
func (i *IDNum) Get(x int) (string, int32) {
	if x < 0 || x >= len(i.Id) {
		return "", 0
	}
	return i.Id[x], i.Num[x]
}

// IDFloat ID/Float值对
type IDFloat struct {
	Id  []string
	Num []float32
}

// Parse 解析数据
func (i *IDFloat) Parse(s string) {
	if s == "" {
		i.Id = make([]string, 0)
		i.Num = make([]float32, 0)
		return
	}

	items := strings.Split(s, ";")
	i.Id = make([]string, len(items))
	i.Num = make([]float32, len(items))
	for x, v := range items {
		in := strings.Split(v, ",")
		i.Id[x] = in[0]
		if len(in) > 1 {
			i.Num[x] = xutils.ParseF32(in[1], 0)
		} else {
			i.Num[x] = 0
		}
	}
}

// Len 获取数据长度
func (i *IDFloat) Len() int {
	return len(i.Id)
}

// Get 获取指定位置的数据
func (i *IDFloat) Get(x int) (string, float32) {
	if x < 0 || x >= len(i.Id) {
		return "", 0
	}
	return i.Id[x], i.Num[x]
}

// TimesRate 次数/比例值对
type TimesRate struct {
	Times []int32
	Rates []float32
}

// Parse 角析数据
func (t *TimesRate) Parse(s string) {
	if s == "" {
		t.Times = make([]int32, 0)
		t.Rates = make([]float32, 0)
		return
	}
	items := strings.Split(s, ";")
	t.Times = make([]int32, len(items))
	t.Rates = make([]float32, len(items))
	for x, v := range items {
		in := strings.Split(v, ",")
		t.Times[x] = xutils.ParseI32(in[0], 1)
		if len(in) > 1 {
			t.Rates[x] = xutils.ParseF32(in[1], 0)
		} else {
			t.Rates[x] = 1
		}
	}
}

// GetRate 获取指定次数的概率
func (t *TimesRate) GetRate(times int32) float32 {
	for i, ts := range t.Times {
		if ts == times {
			return t.Rates[i]
		}
		if ts < times {
			continue
		}
		if i > 0 {
			return t.Rates[i - 1]
		}
		return t.Rates[i]
	}
	return 1
}

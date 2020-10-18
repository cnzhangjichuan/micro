package tools

import (
	"strings"

	"github.com/micro/xutils"
)

// IdNum Id/Num值对
type IdNum struct {
	Id  []string
	Num []int32
}

// Parse 解析数据
func (i *IdNum) Parse(s string) {
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
func (i *IdNum) Len() int {
	return len(i.Id)
}

// Get 获取指定位置的数据
func (i *IdNum) Get(x int) (string, int32) {
	if x < 0 || x >= len(i.Id) {
		return "", 0
	}
	return i.Id[x], i.Num[x]
}

// IdFloat Id/Float值对
type IdFloat struct {
	Id  []string
	Num []float32
}

// Parse 解析数据
func (i *IdFloat) Parse(s string) {
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
func (i *IdFloat) Len() int {
	return len(i.Id)
}

// Get 获取指定位置的数据
func (i *IdFloat) Get(x int) (string, float32) {
	if x < 0 || x >= len(i.Id) {
		return "", 0
	}
	return i.Id[x], i.Num[x]
}

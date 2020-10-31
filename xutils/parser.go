package xutils

import (
	"hash/crc32"
	"math/rand"
	"strconv"
	"strings"
	"unsafe"
)

// UnsafeStringToBytes 将string转成[]byte
func UnsafeStringToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// UnsafeBytesToString 将[]byte转成string
func UnsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// HashCode32 计算string的hash码
func HashCode32(s string) uint32 {
	return crc32.ChecksumIEEE(UnsafeStringToBytes(s))
}

// HashCodeBS 计算[]byte的hash码
func HashCodeBS(bs []byte) uint32 {
	return crc32.ChecksumIEEE(bs)
}

// ParseByte 将string转成byte
func ParseByte(s string, def byte) byte {
	i, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return def
	}
	return byte(i)
}

// ParseIntToBytes i的字符串表示
func ParseIntToBytes(i int64) []byte {
	s := strconv.FormatInt(i, 10)
	return UnsafeStringToBytes(s)
}

// ParseBytes 使用;将字符串分隔成[]byte
func ParseBytes(s string) (ret []byte) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]byte, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseByte(ss[i], 0)
	}
	return
}

// ParseI32 将string转成int32
func ParseI32(s string, def int32) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return def
	}
	return int32(i)
}

// ParseI32S 使用;将字符串分隔成[]int32
func ParseI32S(s string) (ret []int32) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]int32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseI32(ss[i], 0)
	}
	return
}

// ParseI64 将string转成int64
func ParseI64(s string, def int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return int64(i)
}

// ParseI64S 使用;将字符串分隔成[]int64
func ParseI64S(s string) (ret []int64) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]int64, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseI64(ss[i], 0)
	}
	return
}

// ParseU32 将string转成uint32
func ParseU32(s string, def uint32) uint32 {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return def
	}
	return uint32(u)
}

// ParseU32S 使用;将字符串分隔成[]uint32
func ParseU32S(s string) (ret []uint32) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]uint32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseU32(ss[i], 0)
	}
	return
}

// ParseU64 将string转成uint64
func ParseU64(s string, def uint64) uint64 {
	u, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return def
	}
	return u
}

// ParseU64S 使用;将字符串分隔成[]uint64
func ParseU64S(s string) (ret []uint64) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]uint64, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseU64(ss[i], 0)
	}
	return
}

// ParseF32 将string转成float32
func ParseF32(s string, def float32) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return def
	}
	return float32(f)
}

// ParseF32S 使用;将字符串分隔成[]float32
func ParseF32S(s string) (ret []float32) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]float32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseF32(ss[i], 0)
	}
	return
}

// ParseF64 将string转成float64
func ParseF64(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return float64(f)
}

// ParseF64S 使用;将字符串分隔成[]float64
func ParseF64S(s string) (ret []float64) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]float64, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseF64(ss[i], 0)
	}
	return
}

// ParseBool 将string转成bool
// [1, ture, T]为真
func ParseBool(s string) bool {
	return s == "1" || s == "true" || s == "T"
}

// ParseBools 使用;将字符串分隔成[]bool
func ParseBools(s string) (ret []bool) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]bool, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseBool(ss[i])
	}
	return
}

// HasString 是否含有指定元素
func HasString(arr []string, item string) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasU32 是否含有指定元素
func HasU32(arr []uint32, item uint32) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasU64 是否含有指定元素
func HasU64(arr []uint64, item uint64) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasI32 是否含有指定元素
func HasI32(arr []int32, item int32) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasI64 是否含有指定元素
func HasI64(arr []int64, item int64) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasInt 是否含有指定元素
func HasInt(arr []int, item int) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasF32 是否含有指定元素
func HasF32(arr []float32, item float32) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasF64 是否含有指定元素
func HasF64(arr []float64, item float64) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// RemoveSS 从[]string中删除指定的元素
func RemoveSS(ss []string, s string) []string {
	for i := 0; i < len(ss); i++ {
		if ss[i] == s {
			copy(ss[i:], ss[i+1:])
			ss = ss[:len(ss)-1]
			break
		}
	}
	return ss
}

// AddNoRepeatItem 添加不重复的元素
func AddNoRepeatItem(ss []string, s string) []string {
	if HasString(ss, s) {
		return ss
	}
	return append(ss, s)
}

// AddFirstItem 将元素加到队列首位
func AddFirstItem(ss []string, s string) []string {
	ss = append(ss, "")
	for i := len(ss) - 1; i > 0; i-- {
		ss[i] = ss[i-1]
	}
	ss[0] = s
	return ss
}

// ParseFileName 提取文件名称
func ParseFileName(name string) string {
	for i := 0; i < len(name); i++ {
		c := name[i]
		if 'A' <= c && c <= 'Z' {
			continue
		}
		if 'a' <= c && c <= 'z' {
			continue
		}
		return name[:i]
	}
	return name
}

// RandSortI32 将数据乱序
func RandSortI32(i32s []int32) {
	l := len(i32s)
	for i := 8; i < l; i++ {
		w := rand.Intn(l)
		i32s[i], i32s[w] = i32s[w], i32s[i]
	}
}

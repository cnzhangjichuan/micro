package xutils

import (
	"bufio"
	"bytes"
	"hash/crc32"
	"io"
	"math/rand"
	"os"
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
	sl := len(ss)
	ret = make([]int32, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseI32(ss[i], 0)
	}
	return
}

// ParseIS 使用;将字符串分隔成[]int32
func ParseIS(s string) (ret []int) {
	if s == "" {
		return
	}
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
	sl := len(ss)
	ret = make([]int, sl)
	for i := 0; i < sl; i++ {
		v, err := strconv.Atoi(ss[i])
		if err != nil {
			ret[i] = 0
		} else {
			ret[i] = v
		}
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
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
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	} else if strings.Index(s, "/") >= 0 {
		sep = "/"
	}
	ss := strings.Split(s, sep)
	sl := len(ss)
	ret = make([]bool, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ParseBool(ss[i])
	}
	return
}

// HasString 是否含有指定元素
func HasString(arr []string, item string) bool {
	return IndexStrings(arr, item) >= 0
}

// IndexStrings 元素在列表中的位置
func IndexStrings(arr []string, item string) int {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return i
		}
	}
	return -1
}

// IndexI32 元素在列表中的位置
func IndexI32(arr []int32, item int32) int {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return i
		}
	}
	return -1
}

// EqualStrings 数组值是否相等
func EqualStrings(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if s2[i] != v {
			return false
		}
	}
	return true
}

// EqualF32 数组值是否相等
func EqualF32(s1, s2 []float32) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v := range s1 {
		if s2[i] != v {
			return false
		}
	}
	return true
}

// HasU32 是否含有指定元素
func HasU32(arr []uint32, item uint32) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasU64 是否含有指定元素
func HasU64(arr []uint64, item uint64) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasI32 是否含有指定元素
func HasI32(arr []int32, item int32) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasI64 是否含有指定元素
func HasI64(arr []int64, item int64) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasInt 是否含有指定元素
func HasInt(arr []int, item int) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasF32 是否含有指定元素
func HasF32(arr []float32, item float32) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// HasF64 是否含有指定元素
func HasF64(arr []float64, item float64) bool {
	for i, al := 0, len(arr); i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

// RemoveSS 从[]string中删除指定的元素
func RemoveSS(ss []string, s string) []string {
	for i, l := 0, len(ss); i < l; i++ {
		if ss[i] == s {
			copy(ss[i:], ss[i+1:])
			ss = ss[:len(ss)-1]
			break
		}
	}
	return ss
}

// RemoveAt 从[]string中删除指定的元素
func RemoveAt(ss []string, i int) []string {
	if i < 0 || i >= len(ss) {
		return ss
	}
	copy(ss[i:], ss[i+1:])
	return ss[:len(ss)-1]
}

// RemoveAtI32 从[]int32中删除指定的元素
func RemoveAtI32(ss []int32, i int) []int32 {
	if i < 0 || i >= len(ss) {
		return ss
	}
	copy(ss[i:], ss[i+1:])
	return ss[:len(ss)-1]
}

// AddNoRepeatItem 添加不重复的元素
func AddNoRepeatItem(ss []string, s string) []string {
	if HasString(ss, s) {
		return ss
	}
	return append(ss, s)
}

// AddFrontItem 将元素加到队列首位
func AddFrontItem(ss []string, s string) []string {
	ss = append(ss, "")
	for i := len(ss) - 1; i > 0; i-- {
		ss[i] = ss[i-1]
	}
	ss[0] = s
	return ss
}

// AddFrontString 添加不重复的队首元素
func AddFrontString(s string, item string, size int) (string, bool) {
	if s == "" {
		return item, item != ""
	}
	ss := strings.Split(s, ",")
	if len(ss) == 0 {
		return item, item != ""
	}
	if ss[0] == item {
		if len(ss) <= size {
			return s, false
		}
	} else {
		if i := IndexStrings(ss, item); i > 0 {
			ss = append(ss[:i], ss[i+1:]...)
		}
		ss = AddFrontItem(ss, item)
	}
	if len(ss) > size {
		ss = ss[:size]
	}
	return strings.Join(ss, ","), true
}

// AddBackString 添加不重复的队尾元素
func AddBackString(s string, item string, size int) (string, bool) {
	if s == "" {
		return item, item != ""
	}
	ss := strings.Split(s, ",")
	if len(ss) == 0 {
		return item, item != ""
	}
	if ss[len(ss)-1] == item {
		if len(ss) <= size {
			return s, false
		}
	} else {
		if i := IndexStrings(ss, item); i > 0 {
			ss = append(ss[:i], ss[i+1:]...)
		}
		ss = append(ss, item)
	}
	if len(ss) > size {
		ss = ss[len(ss)-size:]
	}
	return strings.Join(ss, ","), true
}

// 将string->[]string，并去除空元素
func SplitN(s string) []string {
	if s == "" {
		return nil
	}
	sep := ","
	if strings.Index(s, ";") >= 0 {
		sep = ";"
	}
	sa := strings.Split(s, sep)
	for i := len(sa) - 1; i >= 0; i-- {
		if sa[i] == "" {
			sa = append(sa[0:i], sa[i+1:]...)
		}
	}
	return sa
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
func RandSortI32(i32s []int32, start int) {
	l := len(i32s)
	if start < 0 {
		start = 0
	}
	for i := start; i < l; i++ {
		w := rand.Intn(l)
		i32s[i], i32s[w] = i32s[w], i32s[i]
	}
}

// RandString 将数组乱序
func RandString(ss []string, start int) {
	l := len(ss)
	if start < 0 {
		start = 0
	}
	for i := start; i < l; i++ {
		w := rand.Intn(l)
		ss[i], ss[w] = ss[w], ss[i]
	}
}

// RandMinMax 获取min到max之间的随机值
func RandMinMax(min, max int32) int32 {
	if min >= max {
		return min
	}
	return min + rand.Int31n(max-min+1)
}

// ReadLineFile 按行读取文件
func ReadLineFile(name string, lc func(string) error) error {
	fd, err := os.Open(name)
	if err != nil {
		return err
	}
	defer fd.Close()

	b := bufio.NewReader(fd)

	var line []byte
	for err == nil {
		line, _, err = b.ReadLine()
		if len(line) > 0 {
			err = lc(string(bytes.TrimSpace(line)))
		}
	}
	if err == io.EOF {
		err = nil
	}
	return err
}

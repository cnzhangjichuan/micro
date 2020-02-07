package xutils

import (
	"github.com/json-iterator/go"
	"io/ioutil"
	"strconv"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func MarshalJson(v interface{}) ([]byte, error) {
	switch d := v.(type) {
	default:
		return json.Marshal(v)
	case []byte:
		return d, nil
	}
}

func UnmarshalJson(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func LoadFromFile(conf interface{}, fileName string) error {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, conf)
}

func SplitStrings(s, sep string) []string {
	if s == "" {
		return nil
	}
	if sep == "" {
		sep = ","
	}
	return strings.Split(s, sep)
}

func ParseByte(s string, def byte) byte {
	i, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		return def
	}
	return byte(i)
}

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

func ParseI32(s string, def int32) int32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return def
	}
	return int32(i)
}

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

func ParseI64(s string, def int64) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return int64(i)
}

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

func ParseU32(s string, def uint32) uint32 {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return def
	}
	return uint32(i)
}

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

func ParseU64(s string, def uint64) uint64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return def
	}
	return uint64(i)
}

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

func ParseF32(s string, def float32) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return def
	}
	return float32(f)
}

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

func ParseF64(s string, def float64) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return float64(f)
}

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

func ParseBool(s string) bool {
	return s == "1"
}

func ParseBools(s string) (ret []bool) {
	if s == "" {
		return
	}
	ss := strings.Split(s, ";")
	sl := len(ss)
	ret = make([]bool, sl)
	for i := 0; i < sl; i++ {
		ret[i] = ss[i] == "1"
	}
	return
}

func HasString(arr []string, item string) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func HasU32(arr []uint32, item uint32) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func HasU64(arr []uint64, item uint64) bool {
	al := len(arr)
	for i := 0; i < al; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func ShitLeftStrings(arr []string) []string {
	al := len(arr)
	for i := 1; i < al; i++ {
		arr[i-1] = arr[i]
	}
	return arr[:al-1]
}

func ShitLeftU32S(arr []uint32) []uint32 {
	al := len(arr)
	for i := 1; i < al; i++ {
		arr[i-1] = arr[i]
	}
	return arr[:al-1]
}

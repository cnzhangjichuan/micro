package xutils

import (
	"crypto/md5"
	"encoding/hex"
	"hash/crc32"
	"unsafe"
)

func UnsafeStringToBytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func UnsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func HashCode32(s string) uint32 {
	return crc32.ChecksumIEEE(UnsafeStringToBytes(s))
}

func HasStringItem(ss []string, s string) bool {
	if s == "" || len(ss) == 0 {
		return false
	}
	for _, _s := range ss {
		if _s == s {
			return true
		}
	}
	return false
}

func MD5String(s string) string {
	ss := md5.Sum(UnsafeStringToBytes(s))
	return hex.EncodeToString(ss[:])
}
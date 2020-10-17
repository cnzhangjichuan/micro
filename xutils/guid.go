package xutils

import (
	"encoding/binary"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"
)

var uuidNext uint32
var uuidPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 20)
	},
}

// GUID 生成唯一标识符
func GUID(group uint16) string {
	uid := uuidPool.Get().([]byte)
	// group
	binary.BigEndian.PutUint16(uid[10:], group)
	// 时间
	binary.BigEndian.PutUint32(uid[12:], uint32(time.Now().Unix()))
	// 序号
	binary.BigEndian.PutUint32(uid[16:], atomic.AddUint32(&uuidNext, 1))

	hex.Encode(uid, uid[10:])
	s := string(uid)
	uuidPool.Put(uid)

	return s
}

package packet

import (
	"encoding/binary"
	"hash/crc32"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// New 创建缓存区
func NewCache(expired time.Duration, saver Saver) Cache {
	const SIZE = 512

	var c cache
	c.cze = uint32(runtime.NumCPU() * 2)
	c.cks = make([]chunk, c.cze)
	for i := uint32(0); i < c.cze; i++ {
		c.cks[i].Init(saver, expired, SIZE)
	}

	return &c
}

// Cache 数据缓存
type Cache interface {
	Has(string) bool
	Load(Serializable, string) bool
	Put(string, Serializable)
	Del(string)
}

// Saver 数据加载器
type Saver interface {
	// Save 将数据保存到指定的数据表中
	Save(data interface{}, id string)

	// Load 从数据表中加载数据
	Load(data interface{}, id string) bool
}

type cache struct {
	cks []chunk
	cze uint32
}

// Has 指定的数据是否存在
func (c *cache) Has(key string) bool {
	if key == "" {
		return false
	}
	return c.cks[hashCode32(key)%c.cze].Has(key)
}

// Load 从缓存中加载数据
func (c *cache) Load(data Serializable, key string) bool {
	if key == "" {
		return false
	}
	return c.cks[hashCode32(key)%c.cze].Load(data, key)
}

// Store 将数据放入缓存中
func (c *cache) Put(key string, data Serializable) {
	if key == "" || data == nil {
		return
	}
	c.cks[hashCode32(key)%c.cze].Put(key, data)
}

// Del 从缓存中删除数据
func (c *cache) Del(key string) {
	if key == "" {
		return
	}
	c.cks[hashCode32(key)%c.cze].Del(key)
}

// hashCode32 计算string的hash码
func hashCode32(s string) uint32 {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return crc32.ChecksumIEEE(*(*[]byte)(unsafe.Pointer(&h)))
}

// chunk 缓存块
type chunk struct {
	sync.RWMutex
	saver   Saver
	m       map[string][]byte
	expired time.Duration
	times   uint32
}

// Init 初始化
func (c *chunk) Init(saver Saver, expired time.Duration, size uint32) {
	c.saver = saver
	c.expired = expired
	c.m = make(map[string][]byte, size)
}

// Has 是否存在指定的数据
func (c *chunk) Has(k string) (ok bool) {
	c.clearExpired()

	c.RLock()
	_, ok = c.m[k]
	c.RUnlock()
	return
}

// Load 从chunk中加载数据
func (c *chunk) Load(v Serializable, k string) (ok bool) {
	c.clearExpired()

	c.RLock()
	data, ok := c.m[k]

	// 组装数据
	if ok {
		pack := NewWithData(data)
		if c.expired > 0 {
			pack.Seek(4, -1)
		}
		v.Decode(pack)
		pack.buf = nil
		Free(pack)
		c.RUnlock()
		return
	}
	c.RUnlock()

	// 没有缓存，从磁盘加载
	ok = c.loadFromDisk(v, k)
	if ok {
		c.Lock()
		c.put(k, v, false)
		c.Unlock()
	}
	return
}

// loadFromDisk 从磁盘加载数据
func (c *chunk) loadFromDisk(v Decoder, k string) bool {
	if c.saver == nil {
		return false
	}
	return c.saver.Load(v, k)
}

func (c *chunk) Put(k string, v Serializable) {
	c.Lock()
	c.put(k, v, true)
	c.Unlock()
}

// Put 设置数据到chunk中
func (c *chunk) put(k string, v Serializable, saved bool) {
	var (
		pack *Packet
		data []byte
		ok   bool
	)

	if data, ok = c.m[k]; ok {
		pack = NewWithData(data)
		pack.Reset()
	} else {
		pack = New(512)
	}

	// 设置过期时间
	if c.expired > 0 {
		binary.BigEndian.PutUint32(pack.Allocate(4), uint32(time.Now().Add(c.expired).Unix()))
	}

	// 编码
	v.Encode(pack)
	c.m[k] = pack.buf[:pack.w]
	pack.buf = nil
	Free(pack)

	// 更新数据到磁盘上
	if saved && c.saver != nil {
		c.saver.Save(v, k)
	}
}

// Del 从chunk中删除数据
func (c *chunk) Del(k string) {
	c.Lock()
	if pack, ok := c.m[k]; ok {
		delete(c.m, k)
		putBytes(pack)
	}
	c.Unlock()
}

// clearExpired 清除过期数据
func (c *chunk) clearExpired() {
	const TIMES = 1000

	if c.expired <= 0 {
		return
	}

	// 每千次调用，清除一次数据
	t := atomic.AddUint32(&c.times, 1)
	if t != TIMES {
		return
	}
	atomic.StoreUint32(&c.times, 0)

	c.Lock()
	now := uint32(time.Now().Unix())
	for k, pack := range c.m {
		if binary.BigEndian.Uint32(pack[:4]) <= now {
			delete(c.m, k)
			putBytes(pack)
		}
	}
	c.Unlock()
}

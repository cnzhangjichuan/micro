package cache

import (
	"encoding/binary"
	"sync"
	"sync/atomic"
	"time"

	"github.com/micro/cast/store"
	"github.com/micro/packet"
	"github.com/micro/xutils"
)

// chunk 缓存块
type chunk struct {
	sync.RWMutex
	name    string
	m       map[string]*packet.Packet
	expired time.Duration
	times   uint32
}

// Init 初始化
func (c *chunk) Init(name string, expired time.Duration, size uint32) {
	c.name = name
	c.expired = expired
	c.m = make(map[string]*packet.Packet, size)
}

// Has 是否存在指定的数据
func (c *chunk) Has(k string) (ok bool) {
	var pack *packet.Packet
	c.clearExpired()

	c.RLock()
	pack, ok = c.m[k]
	// 校验是否过期
	ok = ok && !c.isExpired(pack)
	c.RUnlock()
	return
}

// Load 从chunk中加载数据
func (c *chunk) Load(v packet.Packable, k string) (ok bool) {
	c.clearExpired()

	c.RLock()
	pack, ok := c.m[k]

	// 没有缓存，从磁盘加载
	if !ok {
		c.RUnlock()
		ok = c.loadFromDisk(v, k)
		if ok {
			c.Lock()
			c.put(k, v, false)
			c.Unlock()
		}
		return
	}

	// 回到起始位置
	pack.Seek(0, -1)

	// 数据已过期
	if c.isExpired(pack) {
		c.RUnlock()
		ok = c.loadFromDisk(v, k)
		if ok {
			c.Lock()
			c.put(k, v, false)
			c.Unlock()
		}
		return
	}

	// 过期标识需要占用4个byte
	if c.expired > 0 {
		pack.Seek(4, -1)
	}
	v.Decode(pack)

	c.RUnlock()
	return
}

// loadFromDisk 从磁盘加载数据
func (c *chunk) loadFromDisk(v packet.Decoder, k string) bool {
	if c.name == "" {
		return false
	}
	return store.Load(v, c.name, k)
}

func (c *chunk) Put(k string, v packet.Packable) {
	c.Lock()
	c.put(k, v, true)
	c.Unlock()
}

// Put 设置数据到chunk中
func (c *chunk) put(k string, v packet.Packable, saved bool) {
	var (
		pack   *packet.Packet
		ok     bool
		oriKey uint32 = 0
		nowKey uint32 = 0
	)

	if pack, ok = c.m[k]; ok {
		if saved && c.name != "" {
			if c.expired > 0 {
				oriKey = xutils.HashCodeBS(pack.Slice(4, -1))
			} else {
				oriKey = xutils.HashCodeBS(pack.Slice(0, -1))
			}
		}
		pack.Reset()
	} else {
		pack = packet.New(512)
	}

	// 设置过期时间
	if c.expired > 0 {
		binary.BigEndian.PutUint32(pack.Allocate(4), uint32(time.Now().Add(c.expired).Unix()))
	}

	// 编码
	v.Encode(pack)
	c.m[k] = pack
	if saved && c.name != "" {
		if c.expired > 0 {
			nowKey = xutils.HashCodeBS(pack.Slice(4, -1))
		} else {
			nowKey = xutils.HashCodeBS(pack.Slice(0, -1))
		}
	}

	// 更新数据到磁盘上
	if saved && oriKey != nowKey {
		store.Save(v, c.name, k)
	}
}

// Del 从chunk中删除数据
func (c *chunk) Del(k string) {
	c.Lock()
	if pack, ok := c.m[k]; ok {
		delete(c.m, k)
		packet.Free(pack)
	}
	c.Unlock()
}

// isExpired 是否过期
func (c *chunk) isExpired(pack *packet.Packet) bool {
	// 没有设置过期时长
	if c.expired <= 0 {
		return false
	}

	now := time.Now()
	if binary.BigEndian.Uint32(pack.Slice(0, 4)) <= uint32(now.Unix()) {
		return true
	}

	// 更新过期时间
	binary.BigEndian.PutUint32(pack.Slice(0, 4), uint32(now.Add(c.expired).Unix()))
	return false
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
		if binary.BigEndian.Uint32(pack.Slice(0, 4)) <= now {
			delete(c.m, k)
			packet.Free(pack)
		}
	}
	c.Unlock()
}

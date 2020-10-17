package cache

import (
	"runtime"
	"time"

	"github.com/micro/packet"
	"github.com/micro/xutils"
)

// New 创建缓存区
func New(name string, expired time.Duration) Cache {
	const SIZE = 512

	var c cache
	c.cze = uint32(runtime.NumCPU() * 2)
	c.cks = make([]chunk, c.cze)
	for i := uint32(0); i < c.cze; i++ {
		c.cks[i].Init(name, expired, SIZE)
	}

	return &c
}

// Cache 数据缓存
type Cache interface {
	Has(string) bool
	Load(packet.Packable, string) bool
	Put(string, packet.Packable)
	Del(string)
}

type cache struct {
	cks        []chunk
	cze        uint32
}

// Has 指定的数据是否存在
func (c *cache) Has(key string) bool {
	if key == "" {
		return false
	}
	return c.cks[xutils.HashCode32(key)%c.cze].Has(key)
}

// Load 从缓存中加载数据
func (c *cache) Load(data packet.Packable, key string) bool {
	if key == "" {
		return false
	}
	return c.cks[xutils.HashCode32(key)%c.cze].Load(data, key)
}

// Store 将数据放入缓存中
func (c *cache) Put(key string, data packet.Packable) {
	if key == "" || data == nil {
		return
	}
	c.cks[xutils.HashCode32(key)%c.cze].Put(key, data)
}

// Del 从缓存中删除数据
func (c *cache) Del(key string) {
	if key == "" {
		return
	}
	c.cks[xutils.HashCode32(key)%c.cze].Del(key)
}

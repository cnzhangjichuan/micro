package micro

import (
	"strings"
	"sync"

	"github.com/micro/packet"
)

// Dpo 数据处理对象(data-process-object)
type Dpo interface {
	// Parse 解析数据
	Parse(interface{})

	// LoadUser 加载角色数据
	LoadUser(packet.PackIdentifier) bool

	// SetUser 设置角色数据
	SetUser(packet.PackIdentifier)

	// GetUID 获取UID
	GetUID() string

	// GetCache 设置缓存数据
	// 该数据在会话中一直有效，直到会话结束或被删除
	SetCache(string, interface{})

	// GetString 从设置的缓存中获取数据
	GetString(string) (string, bool)

	// GetI32 从设置的缓存中获取数据
	GetI32(string) (int32, bool)

	// GetI64 从设置的缓存中获取数据
	GetI64(string) (int64, bool)

	// GetU32 从设置的缓存中获取数据
	GetU32(string) (uint32, bool)

	// GetU64 从设置的缓存中获取数据
	GetU64(string) (uint64, bool)

	// GetFloat32 从设置的缓存中获取数据
	GetF32(string) (float32, bool)

	// GetFloat64 从设置的缓存中获取数据
	GetF64(string) (float64, bool)

	// GetCache 从设置的缓存中获取数据
	GetCache(string) (interface{}, bool)

	// DelCache 删除缓存
	DelCache(string)

	// Remote 远端地址
	Remote() string

	// SetGroup 设置分组
	SetGroup(uint8, string)

	// ClearGroup 清空分组
	ClearGroup(uint8)

	// GetGroup 获取分组
	GetGroup(uint8) string
}

// userGroups 分组
type tUserDpoGroup [16]string

func (t *tUserDpoGroup) clear() {
	for i := 0; i < len(t); i++ {
		t[i] = ""
	}
}
func (t *tUserDpoGroup) Match(flag uint8, v string) bool {
	if flag >= 16 {
		return false
	}
	return t[flag] == v
}

type baseDpo struct {
	uid   string
	rem   string
	cache dpoCache
	group *tUserDpoGroup
}

// LoadUser 加载用户数据
func (b *baseDpo) LoadUser(u packet.PackIdentifier) bool {
	uid := u.GetUID()
	if uid == "" {
		uid = b.uid
	}
	if uid == "" {
		return false
	}
	return env.cache.Load(u, uid)
}

// SetUser 设置用户数据
func (b *baseDpo) SetUser(u packet.PackIdentifier) {
	b.uid = u.GetUID()
	env.cache.Put(b.uid, u)
}

// GetUID 角色ID
func (b *baseDpo) GetUID() string {
	return b.uid
}

func (b *baseDpo) SetCache(key string, v interface{}) {
	b.cache[key] = v
}
func (b *baseDpo) GetString(key string) (string, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
func (b *baseDpo) GetI32(key string) (int32, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return 0, false
	}
	u, ok := v.(int32)
	return u, ok
}
func (b *baseDpo) GetI64(key string) (int64, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return 0, false
	}
	u, ok := v.(int64)
	return u, ok
}
func (b *baseDpo) GetU32(key string) (uint32, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return 0, false
	}
	u, ok := v.(uint32)
	return u, ok
}
func (b *baseDpo) GetU64(key string) (uint64, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return 0, false
	}
	u, ok := v.(uint64)
	return u, ok
}
func (b *baseDpo) GetF32(key string) (float32, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return 0, false
	}
	f, ok := v.(float32)
	return f, ok
}
func (b *baseDpo) GetF64(key string) (float64, bool) {
	v, ok := b.GetCache(key)
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}
func (b *baseDpo) GetCache(key string) (interface{}, bool) {
	if b.cache == nil {
		return nil, false
	}
	v, ok := b.cache[key]
	return v, ok
}
func (b *baseDpo) DelCache(key string) {
	delete(b.cache, key)
}
func (b *baseDpo) SetRemote(rem string) {
	if i := strings.LastIndex(rem, ":"); i > 0 {
		b.rem = rem[:i]
	} else {
		b.rem = rem
	}
}
func (b *baseDpo) Remote() string {
	return b.rem
}
func (b *baseDpo) SetGroup(flag uint8, v string) {
	if b.group == nil || flag >= 16 {
		return
	}
	b.group[flag] = v
}
func (b *baseDpo) ClearGroup(flag uint8) {
	if b.group != nil {
		b.group[flag] = ""
	}
}
func (b *baseDpo) GetGroup(flag uint8) string {
	if b.group != nil && flag < 16 {
		return b.group[flag]
	}
	return ""
}
func (b *baseDpo) release() {
	b.uid = ""
	b.rem = ""
	b.cache = nil
	b.group = nil
}

// dpoCache 数据缓存器
type dpoCache map[string]interface{}

// dpoCachePool dpo数据缓存器池
var dpoCachePool = sync.Pool{
	New: func() interface{} {
		return make(dpoCache, 16)
	},
}

// createDpoCache 获取数据缓存器
func createDpoCache() dpoCache {
	return dpoCachePool.Get().(dpoCache)
}

// freeDpoCache 归还数据缓存器
func freeDpoCache(m dpoCache) {
	if m == nil {
		return
	}
	for key := range m {
		delete(m, key)
	}
	dpoCachePool.Put(m)
}

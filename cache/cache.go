package cache

import (
	"encoding/binary"
	"sync"
	"time"
)

// New 创建Cache
func New(bucketSize uint64) *cache {
	var c cache
	c.Init(bucketSize)
	return &c
}

type Cache interface {
	Set(key string, data []byte, expired time.Duration)
	Get(key string) (data []byte, ok bool)
	Del(key string)
}

type cache struct {
	locks   [256]sync.Mutex
	buckets [256]bucket
}

// Init 初始化缓存
func (c *cache) Init(size uint64) {
	for i := 0; i < 256; i++ {
		c.buckets[i].Init(size)
	}
}

// Set 设置数据
func (c *cache) Set(key string, data []byte, expired time.Duration) {
	ks := []byte(key)
	hash := c.hashCode(ks)
	i := hash & 255
	c.locks[i].Lock()
	c.buckets[i].Set(ks, data, hash, expired)
	c.locks[i].Unlock()
	return
}

// Get 获取数据
func (c *cache) Get(key string) (data []byte, ok bool) {
	ks := []byte(key)
	hash := c.hashCode(ks)
	i := hash & 255
	c.locks[i].Lock()
	data, ok = c.buckets[i].Get(ks, hash)
	c.locks[i].Unlock()
	return
}

// Del 删除数据(仅删除缓存)
func (c *cache) Del(key string) {
	ks := []byte(key)
	hash := c.hashCode(ks)
	i := hash & 255
	c.locks[i].Lock()
	c.buckets[i].Del(ks, hash)
	c.locks[i].Unlock()
}

// hashCode
func (c *cache) hashCode(s []byte) uint32 {
	var code uint32
	for i, l := 0, len(s); i < l; i++ {
		code = code<<5 - code + uint32(s[i])
	}
	return code
}

// bucket 数据存储单元
type bucket struct {
	// 存储区
	data struct {
		begin   uint64
		size    uint64
		index   uint64
		maxSize uint64
		limit   uint64
		buf     []byte
	}
	// 索引区
	index struct {
		len [256]uint32
		cap uint32
		ptr []indexPtr
	}
	count uint64
	total uint64
}

// Init 初始化bucket
func (b *bucket) Init(buffSize uint64) {
	const indexCap = 128

	b.data.begin = 0
	b.data.size = 0
	b.data.maxSize = buffSize
	b.data.limit = buffSize >> 2
	b.data.buf = make([]byte, buffSize)
	for i := 0; i < 256; i++ {
		b.index.len[i] = 0
	}
	b.index.cap = indexCap
	b.index.ptr = make([]indexPtr, indexCap*256)
}

// 0 timestamp 4 byte
// 4 expired 4 byte
// 8 hash 2 byte
// 10 key-length 2 byte
// 12 value-length 4 byte
// 16 value-cap 4 byte
// 20 deleted 1 byte
// 21 chunk-index 1 byte
const headerSize = 22

// Set 设置数据
func (b *bucket) Set(key, data []byte, hash uint32, expired time.Duration) {
	const capMore = 16

	nedCap := uint64(headerSize + len(key) + len(data) + capMore)
	if nedCap > b.data.limit {
		return
	}

	var v [headerSize]byte
	offset, idx, code, exists := b.findData(key, hash)
	now := time.Now()
	nowStamp := uint32(now.Unix())

	if exists {
		b.data.index = offset
		b.readBytes(v[:])
		cp := binary.LittleEndian.Uint32(v[16:20])
		ds := uint32(len(data))
		// 如果空间足够时,直接放置在原位置
		if cp >= ds {
			keyLen := binary.LittleEndian.Uint16(v[10:12])
			b.total += uint64(nowStamp - binary.LittleEndian.Uint32(v[:4]))
			binary.LittleEndian.PutUint32(v[:4], nowStamp)
			if expired <= 0 {
				binary.LittleEndian.PutUint32(v[4:8], 0)
			} else {
				binary.LittleEndian.PutUint32(v[4:8], uint32(now.Add(expired).Unix()))
			}
			binary.LittleEndian.PutUint32(v[12:16], ds)
			v[20] = 0
			b.data.index = offset
			b.writeBytes(v[:])
			b.data.index = offset
			b.skip(headerSize + uint64(keyLen))
			b.writeBytes(data)
			return
		}
		// 标记为删除状态
		b.data.buf[offset+20] = 1
		// 删除索引
		b.removeIndex(v[:], offset)
	}

	// 清理空间
	b.clean(nedCap, v[:], nowStamp)

	// 添加数据
	binary.LittleEndian.PutUint32(v[:4], nowStamp)
	if expired > 0 {
		binary.LittleEndian.PutUint32(v[4:8], uint32(now.Add(expired).Unix()))
	} else {
		binary.LittleEndian.PutUint32(v[4:8], 0)
	}
	binary.LittleEndian.PutUint16(v[8:10], code)
	binary.LittleEndian.PutUint16(v[10:12], uint16(len(key)))
	ds := uint32(len(data))
	binary.LittleEndian.PutUint32(v[12:16], ds)
	binary.LittleEndian.PutUint32(v[16:20], ds+capMore)
	v[20] = 0
	v[21] = idx
	b.data.index = b.data.begin
	b.skip(b.data.size)
	offset = b.data.index
	b.writeBytes(v[:])
	b.writeBytes(key)
	b.writeBytes(data)
	b.addIndex(v[:], offset)
}

// clean 清理空间
func (b *bucket) clean(nedCap uint64, v []byte, nowStamp uint32) {
	times := 0
	free := b.data.maxSize - b.data.size
	for nedCap > free {
		b.data.index = b.data.begin
		b.readBytes(v[:])
		del := v[20] == 1
		if !del {
			// 删除过期
			expired := binary.LittleEndian.Uint32(v[4:8])
			del = expired > 0 && expired < nowStamp
		}
		if !del {
			// 删除太久未访问
			stamp := uint64(binary.LittleEndian.Uint32(v[:4])) * b.count
			del = stamp <= b.total
		}
		if !del {
			times += 1
			del = times > 4
		}
		oldOffset := b.data.begin
		s := b.getChuckSize(v[:])
		b.data.begin += s
		if b.data.begin >= b.data.maxSize {
			b.data.begin -= b.data.maxSize
		}
		if del {
			// 删除数据
			b.removeIndex(v[:], oldOffset)
			free = b.data.maxSize - b.data.size
		} else {
			// 将数据移到环尾
			newOffset := oldOffset + b.data.size
			if newOffset >= b.data.maxSize {
				newOffset -= b.data.maxSize
			}
			b.data.index = newOffset
			more := oldOffset + s - b.data.maxSize
			if more <= 0 {
				b.writeBytes(b.data.buf[oldOffset : oldOffset+s])
			} else {
				b.writeBytes(b.data.buf[oldOffset:b.data.maxSize])
				b.writeBytes(b.data.buf[:more])
			}
			b.updateIndex(v, newOffset, oldOffset)
		}
	}
}

// getChuckSize 获取数据块总长度
func (b *bucket) getChuckSize(v []byte) uint64 {
	// header
	size := uint64(headerSize)
	// key size
	size += uint64(binary.LittleEndian.Uint16(v[10:12]))
	// data size
	size += uint64(binary.LittleEndian.Uint32(v[16:20]))
	return size
}

// addIndex 添加索引
func (b *bucket) addIndex(v []byte, offset uint64) {
	idx := v[21]
	code := binary.LittleEndian.Uint16(v[8:10])
	keyLen := binary.LittleEndian.Uint16(v[10:12])
	if b.index.len[idx] >= b.index.cap {
		b.growIndexCap()
	}
	ptr := b.getIndexChunk(idx)
	x := b.findIndex(ptr, code)
	ptr = ptr[:len(ptr)+1]
	copy(ptr[x+1:], ptr[x:])
	ptr[x].hash = code
	ptr[x].keyLen = keyLen
	ptr[x].offset = offset
	b.data.size += b.getChuckSize(v)
	b.index.len[idx] += 1
	b.count += 1
	b.total += uint64(binary.LittleEndian.Uint32(v[:4]))
}

// removeIndex 删除索引
func (b *bucket) removeIndex(v []byte, offset uint64) {
	ptr := b.getIndexChunk(v[21])
	code := binary.LittleEndian.Uint16(v[8:])
	for i, l := b.findIndex(ptr, code), len(ptr); i < l; i++ {
		if ptr[i].hash != code {
			break
		}
		if ptr[i].offset != offset {
			continue
		}
		b.data.size -= b.getChuckSize(v)
		b.index.len[v[21]] -= 1
		b.count -= 1
		b.total -= uint64(binary.LittleEndian.Uint32(v[:4]))
		copy(ptr[:i], ptr[i+1:])
		break
	}
}

// updateIndex 更新索引
func (b *bucket) updateIndex(v []byte, newOffset, oldOffset uint64) {
	idx, hash := v[21], binary.LittleEndian.Uint16(v[8:10])
	ptr := b.getIndexChunk(idx)
	x := b.findIndex(ptr, hash)
	for l := len(ptr); x < l; x++ {
		if ptr[x].hash != hash {
			break
		}
		if ptr[x].offset == oldOffset {
			ptr[x].offset = newOffset
			break
		}
	}
}

// Get 获取key值映射的数据
func (b *bucket) Get(key []byte, hash uint32) (data []byte, ok bool) {
	offset, _, _, exists := b.findData(key, hash)
	if !exists {
		return
	}

	b.data.index = offset
	var v [headerSize]byte
	b.readBytes(v[:])
	// delete?
	if v[20] == 1 {
		ok = false
		return
	}
	// expired?
	now := uint32(time.Now().Unix())
	expired := binary.LittleEndian.Uint32(v[4:8])
	if expired > 0 && expired < now {
		ok = false
		return
	}
	ok = true
	oldStamp := binary.LittleEndian.Uint32(v[:4])
	keyLen := binary.LittleEndian.Uint16(v[10:12])
	valLen := binary.LittleEndian.Uint32(v[12:16])
	b.total += uint64(now - oldStamp)
	binary.LittleEndian.PutUint32(v[:4], now)
	b.data.index = offset
	b.writeBytes(v[:4])
	b.data.index = offset
	b.skip(uint64(headerSize + keyLen))
	data = make([]byte, valLen)
	b.readBytes(data)
	return
}

// Del 删除数据(仅删除缓存)
func (b *bucket) Del(key []byte, hash uint32) {
	offset, _, _, exists := b.findData(key, hash)
	if !exists {
		return
	}
	var v [headerSize]byte
	b.data.index = offset
	b.readBytes(v[:])
	// 删除索引
	b.removeIndex(v[:], offset)
	// 标记删除
	b.data.index = offset
	b.skip(20)
	b.data.buf[b.data.index] = 1
}

// findData 查找数据位置
func (b *bucket) findData(key []byte, hash uint32) (offset uint64, idx uint8, code uint16, ok bool) {
	idx, code = b.splitHashCode(hash)
	ptr := b.getIndexChunk(idx)
	keyLen := uint16(len(key))
	x := b.findIndex(ptr, code)
	for i, l := x, len(ptr); i < l; i++ {
		if ptr[i].hash != code {
			break
		}
		ok = ptr[i].keyLen == keyLen && b.equalAt(key, ptr[i].offset+headerSize)
		if ok {
			offset = ptr[i].offset
			x = i
			break
		}
	}
	return
}

// skip
func (b *bucket) skip(offset uint64) {
	b.data.index += offset
	for b.data.index >= b.data.maxSize {
		b.data.index -= b.data.maxSize
	}
}

// readBytes 从offset位置读取指定大小的数据
func (b *bucket) readBytes(data []byte) {
	n := copy(data, b.data.buf[b.data.index:])
	b.data.index += uint64(n)
	if n < len(data) {
		n = copy(data[n:], b.data.buf[0:])
		b.data.index = uint64(n)
	}
}

// writeBytes 写入数据
func (b *bucket) writeBytes(v []byte) {
	n := copy(b.data.buf[b.data.index:], v)
	b.data.index += uint64(n)
	if n < len(v) {
		n = copy(b.data.buf[0:], v[n:])
		b.data.index = uint64(n)
	}
}

// splitHashCode 拆分hashCode
func (b *bucket) splitHashCode(hash uint32) (idx uint8, code uint16) {
	const lowBit = 255

	hash >>= 8
	idx = uint8(hash & lowBit)
	code = uint16(hash >> 8)
	return
}

// equalAt 是否与传入的值相同
func (b *bucket) equalAt(v []byte, idx uint64) bool {
	for i, l := 0, len(v); i < l; i++ {
		if v[i] != b.data.buf[idx] {
			return false
		}
		idx += 1
		if idx >= b.data.maxSize {
			idx = 0
		}
	}
	return true
}

// indexPtr 索引
type indexPtr struct {
	hash   uint16
	keyLen uint16
	offset uint64
}

// getIndexChunk 获取索引块
func (b *bucket) getIndexChunk(idx uint8) []indexPtr {
	s := uint32(idx) * b.index.cap
	e := s + b.index.len[idx]
	return b.index.ptr[s:e]
}

// findIndex 查找索引
func (b *bucket) findIndex(ptr []indexPtr, hash uint16) (idx int) {
	height := len(ptr)
	for idx < height {
		mid := (idx + height) >> 1
		if ptr[mid].hash < hash {
			idx = mid + 1
		} else {
			height = mid
		}
	}
	return
}

// growIndexCap 扩容
func (b *bucket) growIndexCap() {
	sCap := b.index.cap
	newPtr := make([]indexPtr, len(b.index.ptr)<<1)
	for i := uint32(0); i < 256; i++ {
		off := i * sCap
		copy(newPtr[off<<1:], b.index.ptr[off:off+b.index.len[i]])
	}
	b.index.cap = sCap << 1
	b.index.ptr = newPtr
}

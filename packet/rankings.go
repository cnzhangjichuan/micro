package packet

import (
	"sync"
)

// Rankings 排行榜
type Rankings struct {
	sync.RWMutex

	name  string
	saver SingleSaver
	items []rankingItem
}

// Init 初始化
func (r *Rankings) Init(name string, capacity int, saver SingleSaver) {
	r.name = name
	r.items = make([]rankingItem, capacity)
	for i := 0; i < capacity; i++ {
		r.items[i].Rank = int32(i + 1)
	}
	r.saver = saver
	r.loadData()
}

// RankingItem 数据结果
type RankingItem interface {
	Encoder
	Decoder

	GetID() string
	GetValue() int32
	SetRank(int32)
}

// SingleSaver 数据加载器
type SingleSaver interface {
	// Save 将数据保存到指定的数据表中
	Save(id string, data []byte) (ok bool)

	// Find 从数据表中查询数据
	Find(id string) (data []byte, ok bool)
}

// rankingItem 排名序列内部结构
type rankingItem struct {
	Id    string
	Rank  int32
	Value int32
	Data  []byte
}

// updateItem 更新榜中数据
func (item *rankingItem) updateItem(src RankingItem) {
	item.Id = src.GetID()
	item.Value = src.GetValue()
	pack := NewWithData(item.Data)
	pack.Reset()
	src.Encode(pack)
	item.Data = pack.Data()
	pack.buf = nil
	Free(pack)
}

// 查找ID位置
func (r *Rankings) findIndex(id string) int {
	for i, j := 0, len(r.items)-1; i <= j; i, j = i+1, j-1 {
		if r.items[i].Id == id {
			return i
		}
		if r.items[j].Id == id {
			return j
		}
	}
	// 未上榜
	return -1
}

// GetRank 查找ID对应的排名
func (r *Rankings) GetRank(id string) int32 {
	r.RLock()
	i := r.findIndex(id)
	r.RUnlock()

	if i == -1 {
		return -1
	}
	return r.items[i].Rank
}

// LoadRankings 加载榜单
func (r *Rankings) LoadRankings(creator func(i int) RankingItem, offset, count int, id string) (n int) {
	e := offset + count
	if e > len(r.items) {
		e = len(r.items)
	}
	if offset >= e {
		return 0
	}
	idIdx := -1
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	r.RLock()
	if id != "" {
		idIdx = r.findIndex(id)
		if idIdx >= offset && idIdx < e {
			// 在查找范围内
			idIdx = -1
		} else if idIdx != -1 {
			// 不在查找范围内
			e = -1
		}
	}
	if idIdx != -1 && idIdx < offset {
		pack.buf = r.items[idIdx].Data
		pack.r, pack.w = 0, len(pack.buf)
		item := creator(n)
		n += 1
		item.Decode(pack)
		item.SetRank(r.items[idIdx].Rank)
	}
	for i := offset; i < e; i++ {
		if r.items[i].Id == "" {
			break
		}
		pack.buf = r.items[i].Data
		pack.r, pack.w = 0, len(pack.buf)
		item := creator(n)
		n += 1
		item.Decode(pack)
		item.SetRank(r.items[i].Rank)
	}
	if idIdx >= e {
		pack.buf = r.items[idIdx].Data
		pack.r, pack.w = 0, len(pack.buf)
		item := creator(n)
		n += 1
		item.Decode(pack)
		item.SetRank(r.items[idIdx].Rank)
	}
	r.RUnlock()
	pack.buf = nil
	Free(pack)
	return
}

// LoadNears 加载与自身相邻的榜单数据
func (r *Rankings) LoadNears(creator func(i int) RankingItem, id string, window []int, top3 bool) int {
	var (
		max  = len(r.items) - 1
		s    int
		last = window[len(window)-1]
		n    = 0
	)
	r.RLock()
	i := r.findIndex(id)
	r.RUnlock()
	if i == -1 {
		// 自身没有入榜
		s = max - last
	} else {
		// 自身已入榜
		s = i - window[len(window)/2]
		if s < 3 {
			s = 3
		} else if s+last > max {
			s = max - last
		}
	}
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	r.RLock()
	if top3 {
		for i := 0; i < 3; i++ {
			pack.buf = r.items[i].Data
			pack.r, pack.w = 0, len(pack.buf)
			item := creator(n)
			n += 1
			item.Decode(pack)
			item.SetRank(r.items[i].Rank)
		}
	}
	for i := 0; i < len(window); i++ {
		idx := window[i] + s
		pack.buf = r.items[idx].Data
		pack.r, pack.w = 0, len(pack.buf)
		item := creator(n)
		n += 1
		item.Decode(pack)
		item.SetRank(r.items[idx].Rank)
	}
	r.RUnlock()
	pack.buf = nil
	Free(pack)
	return n
}

// Add 添加数据
func (r *Rankings) Add(item RankingItem) (rank, delta int32) {
	var (
		idx   = -1
		rid   string
		oRank int32
		id    = item.GetID()
	)

	// 没有设置ID
	if id == "" {
		rank, delta = -1, 0
		return
	}

	r.Lock()

	// 如果在榜中, 直接更新榜中数据
	for i := 0; i < len(r.items); i++ {
		rid = r.items[i].Id
		if rid == "" {
			idx, oRank = i, r.items[i].Rank
			break
		}
		if rid == id {
			r.items[i].updateItem(item)
			idx, oRank = i, r.items[i].Rank
			break
		}
	}

	// 不在榜中, 设置到榜尾
	if idx == -1 {
		if item.GetValue() <= r.items[len(r.items)-1].Value {
			// 未上榜
			r.Unlock()
			rank, delta = -1, 0
			return
		}
		idx = len(r.items) - 1
		oRank = r.items[idx].Rank
	}

	// 更新数据
	rank = oRank
	r.items[idx].updateItem(item)

	if idx > 0 && r.items[idx].Value > r.items[idx-1].Value {
		// 如果比上一个数据大，提升排名
		for provIdx := idx - 1; provIdx >= 0; provIdx-- {
			idx = provIdx + 1
			if r.items[idx].Value <= r.items[provIdx].Value {
				break
			}
			rank = r.items[provIdx].Rank
			r.swap(idx, provIdx)
		}
	} else if idx < len(r.items)-1 && r.items[idx].Value < r.items[idx+1].Value {
		// 如果比下一个数据小，下降排名
		for nextIdx := idx + 1; nextIdx < len(r.items); nextIdx++ {
			idx = nextIdx - 1
			if r.items[idx].Value >= r.items[nextIdx].Value {
				break
			}
			rank = r.items[nextIdx].Rank
			r.swap(idx, nextIdx)
		}
	}

	r.Unlock()
	delta = oRank - rank
	return
}

// Swap 交换数据
//func (r *Rankings) Swap(i, j int) {
//	r.Lock()
//	r.swap(i, j)
//	r.Unlock()
//}

// swap 交换数据
func (r *Rankings) swap(i, j int) {
	r.items[j].Id, r.items[i].Id = r.items[i].Id, r.items[j].Id
	r.items[j].Value, r.items[i].Value = r.items[i].Value, r.items[j].Value
	r.items[j].Data, r.items[i].Data = r.items[i].Data, r.items[j].Data
}

// Close 将数据保存到磁盘上。
// 在服务器关之前调用，保存到数据库中。
func (r *Rankings) Close() {
	const initCacheSize = 4096

	if r.saver == nil {
		return
	}

	pack := New(initCacheSize)
	r.RLock()
	for i := 0; i < len(r.items); i++ {
		if r.items[i].Id == "" {
			break
		}
		pack.WriteString(r.items[i].Id)
		pack.WriteI32(r.items[i].Value)
		pack.WriteBytes(r.items[i].Data)
	}
	r.RUnlock()
	pack.Compress(0)
	r.saver.Save(r.name, pack.Data())
	Free(pack)
}

// loadData 从数据库中初始化数据
func (r *Rankings) loadData() {
	if r.saver == nil {
		return
	}
	data, ok := r.saver.Find(r.name)
	if !ok || len(data) == 0 {
		return
	}
	pack := NewWithData(data)
	pack.UnCompress(0)
	r.Lock()
	for i := 0; i < len(r.items); i++ {
		r.items[i].Id = pack.ReadString()
		if r.items[i].Id == "" {
			break
		}
		r.items[i].Value = pack.ReadI32()
		r.items[i].Data = pack.ReadBytes()
	}
	r.Unlock()
	pack.buf = nil
	Free(pack)
}

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

// RankingItem 数据结果
type RankingItem interface {
	Encoder
	Decoder

	GetID() string
	GetValue() int32
	SetRank(int32)
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

// LoadRankings 加载榜单
func (r *Rankings) LoadRankings(creator func(i int) RankingItem, s, count int) int {
	e := s + count
	if e > len(r.items) {
		e = len(r.items)
	}
	if s >= e {
		return 0
	}
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	for i := s; i < e; i++ {
		if r.items[i].Id == "" {
			break
		}
		pack.buf = r.items[i].Data
		pack.r, pack.w = 0, len(pack.buf)
		item := creator(i - s)
		item.Decode(pack)
		item.SetRank(r.items[i].Rank)
	}
	pack.buf = nil
	Free(pack)
	return e - s
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
		return -1, -1
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

// swap 交换数据
func (r *Rankings) swap(i, j int) {
	r.items[j].Id, r.items[i].Id = r.items[i].Id, r.items[j].Id
	r.items[j].Value, r.items[i].Value = r.items[i].Value, r.items[j].Value
	r.items[j].Data, r.items[i].Data = r.items[i].Data, r.items[j].Data
}

// Close 将数据保存到磁盘上
// 一般在服务器关之前调用，保存到数据库中
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

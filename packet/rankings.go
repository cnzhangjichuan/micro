package packet

import (
	"sync"
)

// Rankings 排行榜
type Rankings struct {
	sync.RWMutex

	name  string
	saver Saver
	items []rankingItem
}

type rankingItem struct {
	Id    string
	Rank  int32
	Value int32
	Data  []byte
}

type RankingItem interface {
	Encoder
	Decoder

	GetID() string
	GetValue() int32
	SetRank(int32)
}

// Init 初始化
func (r *Rankings) Init(name string, capacity int, saver Saver) {
	r.name = name
	r.items = make([]rankingItem, capacity)
	for i := 0; i < capacity; i++ {
		r.items[i].Rank = int32(i + 1)
	}
	r.saver = saver
	if saver != nil {
		saver.Load(r, name)
	}
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
		idx   int = -1
		rid   string
		oRank int32
		id    = item.GetID()
		value = item.GetValue()
	)

	r.Lock()

	// 比榜尾还要小, 不更新数据
	if value <= r.items[len(r.items)-1].Value {
		r.Unlock()
		return
	}

	// 如果在榜中, 直接更新榜中数据
	for i := 0; i < len(r.items); i++ {
		rid = r.items[i].Id
		if rid == "" || rid == id {
			// value没有榜中的大
			if value <= r.items[i].Value {
				r.Unlock()
				return
			}
			idx, oRank = i, r.items[i].Rank
			break
		}
	}

	// 不在榜中, 设置到榜尾
	if idx == -1 {
		idx = len(r.items) - 1
		oRank = r.items[idx].Rank
	}

	// 在指定位置更新数据
	rank = oRank
	r.items[idx].Id = id
	r.items[idx].Value = value
	pack := NewWithData(r.items[idx].Data)
	pack.Reset()
	item.Encode(pack)
	r.items[idx].Data = pack.Data()
	pack.buf = nil
	Free(pack)

	// 将数据移动到相应的榜位上
	for i := idx - 1; i >= 0; i-- {
		idx = i + 1
		if r.items[idx].Value <= r.items[i].Value {
			break
		}
		rank = r.items[i].Rank
		r.items[idx].Id, r.items[i].Id = r.items[i].Id, r.items[idx].Id
		r.items[idx].Value, r.items[i].Value = r.items[i].Value, r.items[idx].Value
		r.items[idx].Data, r.items[i].Data = r.items[i].Data, r.items[idx].Data
	}

	r.Unlock()
	delta = oRank - rank
	return
}

// Save 将数据保存到磁盘上
func (r *Rankings) Save() {
	if r.saver != nil {
		r.RLock()
		r.saver.Save(r, r.name)
		r.RUnlock()
	}
}

// Encode 序列化Rankings
func (r *Rankings) Encode(pack *Packet) {
	for i := 0; i < len(r.items); i++ {
		if r.items[i].Id == "" {
			break
		}
		pack.WriteString(r.items[i].Id)
		pack.WriteI32(r.items[i].Value)
		pack.WriteBytes(r.items[i].Data)
	}
}

// Decode 从流中解析数据
func (r *Rankings) Decode(pack *Packet) {
	for i := 0; i < len(r.items); i++ {
		r.items[i].Id = pack.ReadString()
		if r.items[i].Id == "" {
			break
		}
		r.items[i].Value = pack.ReadI32()
		r.items[i].Data = pack.ReadBytes()
	}
}

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
	edIns RankingItem
}

// Init 初始化
func (r *Rankings) Init(name string, capacity int, edIns RankingItem, saver SingleSaver) {
	r.name = name
	r.items = make([]rankingItem, capacity)
	r.edIns = edIns
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
func (r *Rankings) LoadRankings(creator func(int) RankingItem, offset, count int, id string) (n int) {
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
			idIdx = -1
		} else if idIdx != -1 {
			e -= 1
		}
	}
	if idIdx != -1 && idIdx < offset {
		n = r.setRankingItem(pack, creator, idIdx, n)
	}
	for i := offset; i < e; i++ {
		n = r.setRankingItem(pack, creator, i, n)
	}
	if idIdx >= e {
		n = r.setRankingItem(pack, creator, idIdx, n)
	}
	r.RUnlock()
	pack.buf = nil
	Free(pack)
	return
}

// LoadNears 加载与自身相邻的榜单数据
func (r *Rankings) LoadNears(creator func(i int) RankingItem, id string, window []int, top3 bool) (n int) {
	r.RLock()
	cdx := r.findIndex(id)
	r.RUnlock()

	pack := packetPool.Get().(*Packet)
	pack.freed = 0

	// 装载数据
	ids := r.calNearCoordinate(window, cdx, top3)
	r.RLock()
	for _, i := range ids {
		n = r.setRankingItem(pack, creator, i, n)
	}
	r.RUnlock()

	pack.buf = nil
	Free(pack)
	return n
}

// Update 更新数据
func (r *Rankings) Update(item RankingItem) (rank, delta int32) {
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
			rank = oRank
			break
		}
		if rid == id {
			idx, oRank = i, r.items[i].Rank
			rank = oRank
			if r.items[i].Value == item.GetValue() {
				r.updateItem(i, item)
				r.Unlock()
				delta = 0
				return
			}
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
		oRank = r.getOutRankingsRank(item.GetValue())
		rank = oRank
	}

	// 更新数据
	r.updateItem(idx, item)

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

// Replace 将指定位置的数据替换掉
func (r *Rankings) Replace(mine RankingItem, dest string) (rank, delta int32) {
	r.Lock()
	destIdx := r.findIndex(dest)
	if destIdx == -1 {
		r.Unlock()
		rank, delta = -1, 0
		return
	}
	rank = r.items[destIdx].Rank
	mineIdx := r.findIndex(mine.GetID())
	oRank := rank
	if mineIdx >= 0 {
		r.swap(destIdx, mineIdx)
		oRank = r.items[mineIdx].Rank
	} else {
		oRank = r.getOutRankingsRank(mine.GetValue())
	}
	r.updateItem(destIdx, mine)
	r.Unlock()
	delta = oRank - rank
	return
}

// getOutRankingsRank 获取排行榜之外的排名
func (r *Rankings) getOutRankingsRank(value int32) (rank int32) {
	rank = int32(len(r.items))
	offset := value - r.items[rank-1].Value
	if offset > 0 {
		rank += offset
	} else {
		rank -= offset
	}
	return
}

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
	encoder := packetPool.Get().(*Packet)
	encoder.freed = 0
	r.RLock()
	for i := 0; i < len(r.items); i++ {
		if r.items[i].Id == "" {
			break
		}
		pack.WriteString(r.items[i].Id)
		pack.WriteI32(r.items[i].Value)
		encoder.buf = r.items[i].Data
		encoder.r, encoder.w = 0, len(encoder.buf)
		r.edIns.Decode(encoder)
		encoder.Reset()
		encoder.EncodeJSON(r.edIns, false, false)
		pack.WriteBytes(encoder.Data())
	}
	r.RUnlock()
	pack.Compress(0)
	r.saver.Save(r.name, pack.Data())
	encoder.buf = nil
	Free(encoder)
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
	decoder := packetPool.Get().(*Packet)
	decoder.freed = 0
	r.Lock()
	for i := 0; i < len(r.items); i++ {
		r.items[i].Id = pack.ReadString()
		if r.items[i].Id == "" {
			break
		}
		r.items[i].Value = pack.ReadI32()
		decoder.buf = pack.ReadBytes()
		decoder.r, decoder.w = 0, len(decoder.buf)
		decoder.DecodeJSON(r.edIns)
		decoder.Reset()
		r.edIns.Encode(decoder)
		r.items[i].Data = decoder.Data()
	}
	r.Unlock()
	decoder.buf = nil
	Free(decoder)
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

// updateItem 更新榜中数据
func (r *Rankings) updateItem(i int, src RankingItem) {
	r.items[i].Id = src.GetID()
	r.items[i].Value = src.GetValue()
	pack := NewWithData(r.items[i].Data)
	pack.Reset()
	src.Encode(pack)
	r.items[i].Data = pack.Data()
	pack.buf = nil
	Free(pack)
}

// setRankingItem 填充数据
func (r *Rankings) setRankingItem(pack *Packet, creator func(int) RankingItem, i, n int) int {
	if r.items[i].Id == "" {
		return n
	}
	pack.buf = r.items[i].Data
	pack.r, pack.w = 0, len(pack.buf)
	item := creator(n)
	item.Decode(pack)
	item.SetRank(r.items[i].Rank)
	return n + 1
}

// calNearCoordinate 计算邻近的下标列表
func (r *Rankings) calNearCoordinate(window []int, cdx int, top3 bool) []int {
	coordinates := make([]int, 0, len(window)+4)

	// 加入top3
	if top3 {
		coordinates = append(coordinates, 0, 1, 2)
	}

	rMax := len(r.items) - 1
	wMax := window[len(window)-1]

	// 未入榜
	if cdx == -1 {
		s := rMax - wMax
		for _, i := range window {
			coordinates = append(coordinates, i+s)
		}
		return coordinates
	}

	// 计算榜中位置
	ofx := cdx + window[0] - 3
	s := cdx
	if ofx < 0 {
		s -= ofx
	}
	ofx = cdx + wMax - rMax
	if ofx > 0 {
		s -= ofx
	}
	if top3 && cdx < 3 {
		cdx = -1
	}
	for _, i := range window {
		ofx = i + s
		if cdx == ofx {
			cdx = -1
		} else if cdx != -1 && cdx < ofx {
			coordinates = append(coordinates, cdx)
			cdx = -1
		}
		coordinates = append(coordinates, ofx)
	}
	return coordinates
}

package packet

import (
	"math/rand"
	"sync"
)

// Rankings 排行榜
type Rankings struct {
	sync.RWMutex

	name     string
	saver    SingleSaver
	items    []rankingItem
	edCreate func() RankingItem
}

// Init 初始化
func (r *Rankings) Init(name string, capacity int, edCreate func() RankingItem, saver SingleSaver) {
	r.name = name
	r.items = make([]rankingItem, capacity)
	for i := 0; i < capacity; i++ {
		r.items[i].Rank = int32(i + 1)
	}
	r.saver = saver
	r.edCreate = edCreate
	r.loadData()
}

// RankingItem 数据结果
type RankingItem interface {
	Encoder
	Decoder

	GetID() string
	GetValue() int32
	SetRank(int32)
	GetRank() int32
}

// SingleSaver 数据加载器
type SingleSaver interface {
	// Save 将数据保存到指定的数据表中
	Save(id string, data []byte) (ok bool)

	// Find 从数据表中查询数据
	Find(id string, call func(*Packet)) (ok bool)
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

type RankingResult struct {
	List []RankingItem
}

// Encode 实现packet.Encoder接口
func (resp *RankingResult) Encode(p *Packet) {
	cList := uint64(len(resp.List))
	p.WriteU64(cList)
	for i := uint64(0); i < cList; i++ {
		resp.List[i].Encode(p)
	}
}

// GetRankings 获取榜单列表
func (r *Rankings) GetRankings(offset, count int, id string) (ret *RankingResult) {
	ret = &RankingResult{}
	e := offset + count
	if e > len(r.items) {
		e = len(r.items)
	}
	if offset >= e {
		return
	}
	idIdx := -1
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	r.RLock()

	// 加载自身的排名
	ls := e - offset
	if id != "" {
		idIdx = r.findIndex(id)
		if idIdx >= offset && idIdx < e {
			// 如果自身排名在列表中，不再单独加载
			idIdx = -1
		}
		if idIdx != -1 {
			ls += 1
		}
	}

	// 将数据装入结果列表中
	ret.List = make([]RankingItem, 0, ls)
	if idIdx != -1 && idIdx < offset {
		r.loadRankingItem(ret, pack, idIdx)
	}
	for i := offset; i < e; i++ {
		r.loadRankingItem(ret, pack, i)
	}
	if idIdx >= e {
		r.loadRankingItem(ret, pack, idIdx)
	}
	r.RUnlock()

	pack.buf = nil
	Free(pack)
	return
}

// GetNearWindow 获取附近的滑动窗体
func (r *Rankings) GetNearWindow(leftCount, rightCount, step int) []int32 {
	window := make([]int32, leftCount+rightCount)
	v := int32(0)
	stp := int32(step)
	if stp < 2 {
		stp = 2
	}
	for i := leftCount - 1; i >= 0; i-- {
		v -= rand.Int31n(stp) + 1
		window[i] = v
	}
	ls := leftCount + rightCount
	v = 0
	for i := leftCount; i < ls; i++ {
		v += rand.Int31n(stp) + 1
		window[i] = v
	}
	return window
}

// GetNears 加载与自身相邻的榜单数据
// id 自身ID
// window 查找的窗口范围
// top 加载前几名
// mine 是否包含自已
func (r *Rankings) GetNears(id string, window []int32, top int, mine bool) (ret *RankingResult) {
	ret = &RankingResult{}

	// 查找自身所在的索引位置
	r.RLock()
	cdx := r.findIndex(id)
	r.RUnlock()

	pack := packetPool.Get().(*Packet)
	pack.freed = 0

	// 装载数据
	ids := r.calNearCoordinate(window, cdx, top)
	r.RLock()
	skip := false
	for _, i := range ids {
		if !mine && i == cdx {
			skip = true
			continue
		}
		r.loadRankingItem(ret, pack, i)
	}
	if skip && len(ids) > 0 {
		// 由于跳过自身的数据，结果中少一个，在这里补上
		last := ids[len(ids)-1] + 1
		if last < len(r.items) {
			r.loadRankingItem(ret, pack, last)
		}
	}
	r.RUnlock()

	pack.buf = nil
	Free(pack)
	return
}

// AddRobot 添加机器人
func (r *Rankings) AddRobot(item RankingItem) (rank int32) {
	r.Lock()
	i := r.findIndex(item.GetID())
	if i >= 0 {
		rank = r.items[i].Rank
		r.Unlock()
		return
	}
	rank, _ = r.update(item)
	r.Unlock()
	return
}

// Update 更新数据
func (r *Rankings) Update(item RankingItem) (rank, delta int32) {
	r.Lock()
	rank, delta = r.update(item)
	r.Unlock()
	return
}

// UpdateTwo 更新数据
func (r *Rankings) UpdateTwo(mine, opp RankingItem, syncFunc func(int32, int32)) (mineRank, mineDelta, oppRank, oppDelta int32) {
	r.Lock()
	if syncFunc != nil {
		var (
			mv int32
			ov int32
		)
		mdx := r.findIndex(mine.GetID())
		if mdx >= 0 {
			mv = r.items[mdx].Value
		} else {
			mv = mine.GetValue()
		}
		odx := r.findIndex(opp.GetID())
		if odx >= 0 {
			ov = r.items[odx].Value
		} else {
			ov = opp.GetValue()
		}
		syncFunc(mv, ov)
	}
	mineRank, mineDelta = r.update(mine)
	oppRank, oppDelta = r.update(opp)
	r.Unlock()
	return
}

func (r *Rankings) update(item RankingItem) (rank, delta int32) {
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

	delta = oRank - rank
	return
}

// Replace 将指定位置的数据替换掉
func (r *Rankings) Replace(mine RankingItem, dest string) (rank, raise int32) {
	r.Lock()
	destIdx := r.findIndex(dest)
	if destIdx == -1 {
		r.Unlock()
		rank, raise = -1, 0
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
	raise = oRank - rank
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

// Save 将数据保存到磁盘上。
// 在服务器关闭之前调用，保存到数据库中。
func (r *Rankings) Save() {
	const (
		initCacheSize = 4096
		buffItemSize  = 1024
	)

	if r.saver == nil && r.edCreate != nil {
		return
	}

	pack := New(initCacheSize)
	buff := New(buffItemSize)
	dec := r.edCreate()
	r.RLock()
	for i := 0; i < len(r.items); i++ {
		if r.items[i].Id == "" {
			break
		}
		pack.WriteString(r.items[i].Id)
		pack.WriteI32(r.items[i].Value)
		rs := pack.w
		pack.r = rs
		pack.Write(r.items[i].Data)
		dec.Decode(pack)
		pack.r, pack.w = 0, rs
		buff.Reset()
		buff.EncodeJSON(dec, false, false)
		pack.WriteBytes(buff.Data())
	}
	r.RUnlock()
	r.saver.Save(r.name, pack.Data())
	Free(buff)
	Free(pack)
}

// loadData 从数据库中初始化数据
func (r *Rankings) loadData() {
	if r.saver == nil && r.edCreate != nil {
		return
	}
	r.saver.Find(r.name, func(pack *Packet) {
		buff := packetPool.Get().(*Packet)
		buff.freed = 0
		enc := r.edCreate()
		r.Lock()
		for i := 0; i < len(r.items); i++ {
			r.items[i].Id = pack.ReadString()
			if r.items[i].Id == "" {
				break
			}
			r.items[i].Value = pack.ReadI32()
			buff.buf = pack.ReadBytes()
			buff.r, buff.w = 0, len(buff.buf)
			buff.DecodeJSON(enc)
			buff.Reset()
			enc.Encode(buff)
			r.items[i].Data = buff.Data()
		}
		r.Unlock()
		buff.buf = nil
		Free(buff)
	})
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
func (r *Rankings) loadRankingItem(ret *RankingResult, pack *Packet, i int) {
	if r.items[i].Id == "" {
		return
	}
	pack.buf = r.items[i].Data
	pack.r, pack.w = 0, len(pack.buf)
	item := r.edCreate()
	item.Decode(pack)
	item.SetRank(r.items[i].Rank)
	ret.List = append(ret.List, item)
}

// calNearCoordinate 计算邻近的下标列表
func (r *Rankings) calNearCoordinate(window []int32, cdx int, top int) []int {
	if top < 0 {
		top = 0
	}
	coordinates := make([]int, 0, len(window)+top)

	// 加入top
	for i := 0; i < top; i++ {
		coordinates = append(coordinates, i)
	}

	rMax := len(r.items) - 1
	wMax := int(window[len(window)-1])

	mdx := cdx
	// 未入榜时定位到榜未
	if mdx == -1 {
		mdx = rMax - wMax
	}

	if mdx+int(window[0]) < top {
		// 在top中时，定位到top榜外
		mdx = top - int(window[0])
	} else if mdx+wMax > rMax {
		// 超出榜尾时，定位到榜尾
		mdx = rMax - wMax
	}

	// 去除top榜中数据
	if cdx < top {
		cdx = -1
	}

	// 放入数据索引
	for _, i := range window {
		rdx := int(i) + mdx
		if rdx < 0 || rdx > rMax {
			continue
		}
		if cdx != -1 {
			if rdx == cdx {
				cdx = -1
			} else if rdx > cdx {
				coordinates = append(coordinates, cdx)
				cdx = -1
			}
		}
		coordinates = append(coordinates, rdx)
	}

	return coordinates
}

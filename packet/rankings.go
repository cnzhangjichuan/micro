package packet

import (
	"math/rand"
	"sync"
)

// Rankings 排行榜
type Rankings struct {
	rm       sync.RWMutex
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

// Remove 移除榜中数据
func (r *Rankings) Remove(id string) {
	r.rm.Lock()
	start := false
	l := len(r.items)
	for i := 0; i < l; i++ {
		if !start && r.items[i].Id == id {
			start = true
			continue
		}
		if start {
			r.items[i-1].Id = r.items[i].Id
			r.items[i-1].Value = r.items[i].Value
			r.items[i-1].Data = r.items[i].Data
		}
	}
	if start {
		r.items[l-1].Id = ""
		r.items[l-1].Value = 0
		r.items[l-1].Data = nil
	}
	r.rm.Unlock()
}

// Clear 清除榜单返回第一条数据
func (r *Rankings) Clear(first RankingItem) (ok bool) {
	r.rm.Lock()
	l := len(r.items)
	if l > 0 && r.items[0].Id != "" {
		ok = true
		pack := packetPool.Get().(*Packet)
		pack.freed = 0
		pack.buf = r.items[0].Data
		pack.r, pack.w = 0, len(pack.buf)
		first.Decode(pack)
		first.SetRank(r.items[0].Rank)
		pack.buf = nil
		Free(pack)
	}
	r.items = make([]rankingItem, l)
	for i := 0; i < l; i++ {
		r.items[i].Rank = int32(i + 1)
	}
	//if l > 0 && r.saver != nil {
	//	r.saver.Save(r.name, []byte{})
	//}
	r.rm.Unlock()
	return
}

// GetRank 查找ID对应的排名
func (r *Rankings) GetRank(id string) int32 {
	r.rm.RLock()
	i := r.findIndex(id, 0)
	r.rm.RUnlock()

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
func (r *Rankings) GetRankings(offset, count int, id string, minValue int32) (ret *RankingResult) {
	ret = &RankingResult{}
	pack := packetPool.Get().(*Packet)
	pack.freed = 0
	r.rm.RLock()

	found := id == ""
	// 加载列表
	for i := 0; i < len(r.items); i++ {
		if r.items[i].Value < minValue {
			break
		}
		// 加载传入的值
		if r.items[i].Id == id {
			found = true
			r.loadRankingItem(ret, pack, i)
			continue
		}
		if i < offset {
			continue
		}
		r.loadRankingItem(ret, pack, i)
		count -= 1
		if count <= 0 {
			break
		}
	}

	// 加载传入的值
	if !found {
		x := r.findIndex(id, minValue)
		if x >= 0 {
			r.loadRankingItem(ret, pack, x)
		}
	}
	r.rm.RUnlock()
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

// RandomID 随机加载一个相领的元素ID
func (r *Rankings) RandomID(id string, offset int) (string, int32) {
	var (
		maxCount = offset * 2
		ids      = make([]string, 0, maxCount)
		vls      = make([]int32, 0, maxCount)
		i        = 0
	)
	r.rm.RLock()
	for ; i < len(r.items); i++ {
		// 没有数据时跳出
		if r.items[i].Id == "" {
			break
		}
		// 找到传入ID位置
		if r.items[i].Id != id {
			continue
		}
		// 向下取数
		for j := i + 1; j < len(r.items) && len(ids) < offset; j++ {
			if r.items[j].Id == "" {
				break
			}
			ids = append(ids, r.items[j].Id)
			vls = append(vls, r.items[j].Value)
		}
		// 向上取数
		for j := i - 1; j >= 0 && len(ids) < maxCount; j-- {
			ids = append(ids, r.items[j].Id)
			vls = append(vls, r.items[j].Value)
		}
		break
	}
	// 不在榜中，从最未取数
	if len(ids) == 0 {
		for j := i - 1; j >= 0 && len(ids) < maxCount; j-- {
			ids = append(ids, r.items[j].Id)
			vls = append(vls, r.items[j].Value)
		}
	}
	r.rm.RUnlock()

	// 从结果中随机一个数值
	l := len(ids)
	if l == 0 {
		return "", 0
	}
	if l == 1 {
		return ids[0], vls[0]
	}
	idx := rand.Intn(l)
	return ids[idx], vls[idx]
}

// GetNears 加载与自身相邻的榜单数据
// id 自身ID
// window 查找的窗口范围
// top 加载前几名
// mine 是否包含自已
func (r *Rankings) GetNears(id string, window []int32, top int, mine bool) (ret *RankingResult) {
	ret = &RankingResult{}

	// 查找自身所在的索引位置
	r.rm.RLock()
	cdx := r.findIndex(id, 0)
	r.rm.RUnlock()

	pack := packetPool.Get().(*Packet)
	pack.freed = 0

	// 装载数据
	ids := r.calNearCoordinate(window, cdx, top)
	r.rm.RLock()
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
	r.rm.RUnlock()

	pack.buf = nil
	Free(pack)
	return
}

// AddRobot 添加机器人
func (r *Rankings) AddRobot(item RankingItem) (rank int32) {
	r.rm.Lock()
	i := r.findIndex(item.GetID(), 0)
	if i >= 0 {
		rank = r.items[i].Rank
		r.rm.Unlock()
		return
	}
	rank, _ = r.update(item)
	r.rm.Unlock()
	return
}

// Update 更新数据
func (r *Rankings) Update(item RankingItem) (rank, delta int32) {
	r.rm.Lock()
	rank, delta = r.update(item)
	r.rm.Unlock()
	return
}

// update 更新数据
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
	r.rm.Lock()
	destIdx := r.findIndex(dest, 0)
	if destIdx == -1 {
		r.rm.Unlock()
		rank, raise = -1, 0
		return
	}
	rank = r.items[destIdx].Rank
	mineIdx := r.findIndex(mine.GetID(), 0)
	oRank := rank
	if mineIdx >= 0 {
		r.swap(destIdx, mineIdx)
		oRank = r.items[mineIdx].Rank
	} else {
		oRank = r.getOutRankingsRank(mine.GetValue())
	}
	r.updateItem(destIdx, mine)
	r.rm.Unlock()
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
	r.rm.RLock()
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
	r.rm.RUnlock()
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
		r.rm.Lock()
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
		r.rm.Unlock()
		buff.buf = nil
		Free(buff)
	})
}

// 查找ID位置
func (r *Rankings) findIndex(id string, minValue int32) int {
	for i, j := 0, len(r.items)-1; i <= j; i, j = i+1, j-1 {
		if r.items[i].Id == id {
			if r.items[i].Value < minValue {
				return -1
			}
			return i
		}
		if r.items[j].Id == id {
			if r.items[j].Value < minValue {
				return -1
			}
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

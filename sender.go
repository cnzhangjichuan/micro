package micro

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/micro/packet"
)

const workerSize = 16

// sender 数据发送
type sender struct {
	sync.RWMutex
	valid bool
	seq   uint32
	wgo   [workerSize]chan *wkConnData

	// 共享池
	autoDataPool sync.Pool
	connDataPool sync.Pool
}

// Init 初始化
func (w *sender) Init() {
	w.valid = true
	w.autoDataPool.New = func() interface{} {
		return &wkAutoData{}
	}
	w.connDataPool.New = func() interface{} {
		return &wkConnData{}
	}
	for i := 0; i < workerSize; i++ {
		w.wgo[i] = make(chan *wkConnData, 512)
		go func(c <-chan *wkConnData) {
			const WT = time.Second * 10

			for wk := range c {
				wk.ad.pack.SetTimeout(0, WT)
				wk.ad.pack.FlushToConn(wk.conn)
				w.FreeConnData(wk)
			}
		}(w.wgo[i])
	}
}

type wkAutoData struct {
	ref  int64
	pack *packet.Packet
}

// FreeAutoData 释放AutoData
func (w *sender) FreeAutoData(ad *wkAutoData) {
	if ad == nil {
		return
	}
	if atomic.AddInt64(&ad.ref, -1) == 0 {
		packet.Free(ad.pack)
		w.autoDataPool.Put(ad)
	}
}

type wkConnData struct {
	conn net.Conn
	ad   *wkAutoData
}

// FreeConnData 释放ConnData
func (w *sender) FreeConnData(wc *wkConnData) {
	w.FreeAutoData(wc.ad)
	wc.conn, wc.ad = nil, nil
	w.connDataPool.Put(wc)
}

// Close 关闭
func (w *sender) Close() {
	w.Lock()
	w.valid = false
	for i := 0; i < workerSize; i++ {
		close(w.wgo[i])
	}
	w.Unlock()
}

// NewAutoData 创建释放器
func (w *sender) NewAutoData(pack *packet.Packet) *wkAutoData {
	ad := w.autoDataPool.Get().(*wkAutoData)
	ad.pack, ad.ref = pack, 1
	return ad
}

// AddConnData 添加数据
func (w *sender) AddConnData(conn net.Conn, ad *wkAutoData) {
	w.RLock()
	if w.valid {
		wk := w.connDataPool.Get().(*wkConnData)
		wk.conn, wk.ad = conn, ad
		atomic.AddInt64(&ad.ref, 1)
		w.wgo[atomic.AddUint32(&w.seq, 1)%workerSize] <- wk
	}
	w.RUnlock()
}

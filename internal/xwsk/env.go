package xwsk

import (
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"github.com/cnzhangjichuan/micro/xutils"
	"sync"
)

var env struct {
	handlers its.Handlers
	log      types.Logger
	sds      wsDps
}

func InitEnv(config *types.EnvConfig, handlers its.Handlers, logger types.Logger) {
	const MAP_CAP = 64

	env.handlers = handlers
	env.log = logger
	env.sds.g1 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g2 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g3 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g4 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g5 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g6 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g7 = make(map[string]*wsDpo, MAP_CAP)
	env.sds.g8 = make(map[string]*wsDpo, MAP_CAP)
}

// log for error
func logError(fmt string, v ...interface{}) {
	if env.log != nil {
		env.log.Error(fmt, v...)
	}
}

type wsDps struct {
	m1 sync.RWMutex
	g1 map[string]*wsDpo

	m2 sync.RWMutex
	g2 map[string]*wsDpo

	m3 sync.RWMutex
	g3 map[string]*wsDpo

	m4 sync.RWMutex
	g4 map[string]*wsDpo

	m5 sync.RWMutex
	g5 map[string]*wsDpo

	m6 sync.RWMutex
	g6 map[string]*wsDpo

	m7 sync.RWMutex
	g7 map[string]*wsDpo

	m8 sync.RWMutex
	g8 map[string]*wsDpo
}

func (w *wsDps) Get(uid string) (dpo *wsDpo, ok bool) {
	switch xutils.HashCode32(uid) % 8 {
	default:
		w.m8.RLock()
		dpo, ok = w.g8[uid]
		w.m8.RUnlock()

	case 0:
		w.m1.RLock()
		dpo, ok = w.g1[uid]
		w.m1.RUnlock()

	case 1:
		w.m2.RLock()
		dpo, ok = w.g2[uid]
		w.m2.RUnlock()

	case 2:
		w.m3.RLock()
		dpo, ok = w.g3[uid]
		w.m3.RUnlock()

	case 3:
		w.m4.RLock()
		dpo, ok = w.g4[uid]
		w.m4.RUnlock()

	case 4:
		w.m5.RLock()
		dpo, ok = w.g5[uid]
		w.m5.RUnlock()

	case 5:
		w.m6.RLock()
		dpo, ok = w.g6[uid]
		w.m6.RUnlock()

	case 6:
		w.m7.RLock()
		dpo, ok = w.g7[uid]
		w.m7.RUnlock()

	}
	return
}

func (w *wsDps) Set(dpo *wsDpo) {
	if dpo.user == nil {
		return
	}
	var (
		ori *wsDpo
		ok  bool
		uid = dpo.user.GetUserId()
	)
	switch xutils.HashCode32(uid) % 8 {
	default:
		w.m8.Lock()
		ori, ok = w.g8[uid]
		w.g8[uid] = dpo
		w.m8.Unlock()

	case 0:
		w.m1.Lock()
		ori, ok = w.g1[uid]
		w.g1[uid] = dpo
		w.m1.Unlock()

	case 1:
		w.m2.Lock()
		ori, ok = w.g2[uid]
		w.g2[uid] = dpo
		w.m2.Unlock()

	case 2:
		w.m3.Lock()
		ori, ok = w.g3[uid]
		w.g3[uid] = dpo
		w.m3.Unlock()

	case 3:
		w.m4.Lock()
		ori, ok = w.g4[uid]
		w.g4[uid] = dpo
		w.m4.Unlock()

	case 4:
		w.m5.Lock()
		ori, ok = w.g5[uid]
		w.g5[uid] = dpo
		w.m5.Unlock()

	case 5:
		w.m6.Lock()
		ori, ok = w.g6[uid]
		w.g6[uid] = dpo
		w.m6.Unlock()

	case 6:
		w.m7.Lock()
		ori, ok = w.g7[uid]
		w.g7[uid] = dpo
		w.m7.Unlock()

	}
	if ok {
		ori.Close()
	}
}

func (w *wsDps) Del(dpo *wsDpo) (success bool) {
	if dpo.user == nil {
		return success
	}
	uid := dpo.user.GetUserId()
	switch xutils.HashCode32(uid) % 8 {
	default:
		w.m8.Lock()
		if o, ok := w.g8[uid]; ok && o == dpo {
			delete(w.g8, uid)
			success = true
		}
		w.m8.Unlock()

	case 0:
		w.m1.Lock()
		if o, ok := w.g1[uid]; ok && o == dpo {
			delete(w.g1, uid)
			success = true
		}
		w.m1.Unlock()

	case 1:
		w.m2.Lock()
		if o, ok := w.g2[uid]; ok && o == dpo {
			delete(w.g2, uid)
			success = true
		}
		w.m2.Unlock()

	case 2:
		w.m3.Lock()
		if o, ok := w.g3[uid]; ok && o == dpo {
			delete(w.g3, uid)
			success = true
		}
		w.m3.Unlock()

	case 3:
		w.m4.Lock()
		if o, ok := w.g4[uid]; ok && o == dpo {
			delete(w.g4, uid)
			success = true
		}
		w.m4.Unlock()

	case 4:
		w.m5.Lock()
		if o, ok := w.g5[uid]; ok && o == dpo {
			delete(w.g5, uid)
			success = true
		}
		w.m5.Unlock()

	case 5:
		w.m6.Lock()
		if o, ok := w.g6[uid]; ok && o == dpo {
			delete(w.g6, uid)
			success = true
		}
		w.m6.Unlock()

	case 6:
		w.m7.Lock()
		if o, ok := w.g7[uid]; ok && o == dpo {
			delete(w.g7, uid)
			success = true
		}
		w.m7.Unlock()

	}
	return success
}

func (w *wsDps) Foreach(f func(dpo *wsDpo)) {
	w.m1.RLock()
	for _, d := range w.g1 {
		f(d)
	}
	w.m1.RUnlock()

	w.m2.RLock()
	for _, d := range w.g2 {
		f(d)
	}
	w.m2.RUnlock()

	w.m3.RLock()
	for _, d := range w.g3 {
		f(d)
	}
	w.m3.RUnlock()

	w.m4.RLock()
	for _, d := range w.g4 {
		f(d)
	}
	w.m4.RUnlock()

	w.m5.RLock()
	for _, d := range w.g5 {
		f(d)
	}
	w.m5.RUnlock()

	w.m6.RLock()
	for _, d := range w.g6 {
		f(d)
	}
	w.m6.RUnlock()

	w.m7.RLock()
	for _, d := range w.g7 {
		f(d)
	}
	w.m7.RUnlock()

	w.m8.RLock()
	for _, d := range w.g8 {
		f(d)
	}
	w.m8.RUnlock()
}

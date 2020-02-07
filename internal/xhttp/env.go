package xhttp

import (
	"github.com/cnzhangjichuan/micro/internal/its"
	"github.com/cnzhangjichuan/micro/types"
	"github.com/cnzhangjichuan/micro/xutils"
	"net/http"
	"sync"
	"time"
)

var env struct {
	handlers its.Handlers
	log      types.Logger
	usp      userPool
	assets   string
	fs       http.Handler
}

// init env
func InitEnv(config *types.EnvConfig, handlers its.Handlers, log types.Logger) {
	const MaxTimeout = 60 * 30

	initUserPool(&env.usp, MaxTimeout)
	env.assets = "./assets"
	env.fs = http.FileServer(http.Dir(env.assets))
	env.log = log
	env.handlers = handlers
}

// Init user pool
func initUserPool(up *userPool, maxTimeout int64) *userPool {
	const MAP_CAP = 64

	up.p1 = make(map[string]types.User, MAP_CAP)
	up.p2 = make(map[string]types.User, MAP_CAP)
	up.p3 = make(map[string]types.User, MAP_CAP)
	up.p4 = make(map[string]types.User, MAP_CAP)
	up.p5 = make(map[string]types.User, MAP_CAP)
	up.p6 = make(map[string]types.User, MAP_CAP)
	up.p7 = make(map[string]types.User, MAP_CAP)
	up.p8 = make(map[string]types.User, MAP_CAP)
	go up.gc(maxTimeout)

	return up
}

// user pool
type userPool struct {
	m1 sync.RWMutex
	p1 map[string]types.User

	m2 sync.RWMutex
	p2 map[string]types.User

	m3 sync.RWMutex
	p3 map[string]types.User

	m4 sync.RWMutex
	p4 map[string]types.User

	m5 sync.RWMutex
	p5 map[string]types.User

	m6 sync.RWMutex
	p6 map[string]types.User

	m7 sync.RWMutex
	p7 map[string]types.User

	m8 sync.RWMutex
	p8 map[string]types.User
}

func (up *userPool) Del(uid string) {
	switch xutils.HashCode32(uid) % 8 {
	default:
		up.m8.Lock()
		delete(up.p8, uid)
		up.m8.Unlock()
	case 0:
		up.m1.Lock()
		delete(up.p1, uid)
		up.m1.Unlock()
	case 1:
		up.m2.Lock()
		delete(up.p2, uid)
		up.m2.Unlock()
	case 2:
		up.m3.Lock()
		delete(up.p3, uid)
		up.m3.Unlock()
	case 3:
		up.m4.Lock()
		delete(up.p4, uid)
		up.m4.Unlock()
	case 4:
		up.m5.Lock()
		delete(up.p5, uid)
		up.m5.Unlock()
	case 5:
		up.m6.Lock()
		delete(up.p6, uid)
		up.m6.Unlock()
	case 6:
		up.m7.Lock()
		delete(up.p7, uid)
		up.m7.Unlock()
	}
}

func (up *userPool) Set(user types.User) {
	uid := user.GetUserId()
	user.SetTimeStamp(time.Now().Unix())
	switch xutils.HashCode32(uid) % 8 {
	default:
		up.m8.Lock()
		up.p8[uid] = user
		up.m8.Unlock()
	case 0:
		up.m1.Lock()
		up.p1[uid] = user
		up.m1.Unlock()
	case 1:
		up.m2.Lock()
		up.p2[uid] = user
		up.m2.Unlock()
	case 2:
		up.m3.Lock()
		up.p3[uid] = user
		up.m3.Unlock()
	case 3:
		up.m4.Lock()
		up.p4[uid] = user
		up.m4.Unlock()
	case 4:
		up.m5.Lock()
		up.p5[uid] = user
		up.m5.Unlock()
	case 5:
		up.m6.Lock()
		up.p6[uid] = user
		up.m6.Unlock()
	case 6:
		up.m7.Lock()
		up.p7[uid] = user
		up.m7.Unlock()
	}
}

func (up *userPool) Get(uid string) (u types.User, ok bool) {
	switch xutils.HashCode32(uid) % 8 {
	default:
		up.m8.RLock()
		u, ok = up.p8[uid]
		up.m8.RUnlock()
	case 0:
		up.m1.RLock()
		u, ok = up.p1[uid]
		up.m1.RUnlock()
	case 1:
		up.m2.RLock()
		u, ok = up.p2[uid]
		up.m2.RUnlock()
	case 2:
		up.m3.RLock()
		u, ok = up.p3[uid]
		up.m3.RUnlock()
	case 3:
		up.m4.RLock()
		u, ok = up.p4[uid]
		up.m4.RUnlock()
	case 4:
		up.m5.RLock()
		u, ok = up.p5[uid]
		up.m5.RUnlock()
	case 5:
		up.m6.RLock()
		u, ok = up.p6[uid]
		up.m6.RUnlock()
	case 6:
		up.m7.RLock()
		u, ok = up.p7[uid]
		up.m7.RUnlock()
	}
	if u != nil {
		u.SetTimeStamp(time.Now().Unix())
	}
	return
}

func (up *userPool) gc(maxTimeout int64) {
	const period = time.Second * 30

	now := time.Now().Unix()

	// group 1
	up.m1.Lock()
	for id, u := range up.p1 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p1, id)
		}
	}
	up.m1.Unlock()

	// group 2
	up.m2.Lock()
	for id, u := range up.p2 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p2, id)
		}
	}
	up.m2.Unlock()

	// group 3
	up.m3.Lock()
	for id, u := range up.p3 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p3, id)
		}
	}
	up.m3.Unlock()

	// group 4
	up.m4.Lock()
	for id, u := range up.p4 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p4, id)
		}
	}
	up.m4.Unlock()

	// group 5
	up.m5.Lock()
	for id, u := range up.p5 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p5, id)
		}
	}
	up.m5.Unlock()

	// group 6
	up.m6.Lock()
	for id, u := range up.p6 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p6, id)
		}
	}
	up.m6.Unlock()

	// group 7
	up.m7.Lock()
	for id, u := range up.p7 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p7, id)
		}
	}
	up.m7.Unlock()

	// group 8
	up.m8.Lock()
	for id, u := range up.p8 {
		if now-u.GetTimeStamp() > maxTimeout {
			delete(up.p8, id)
		}
	}
	up.m8.Unlock()

	// pause period
	time.Sleep(period)

	// new round gc
	up.gc(maxTimeout)
}

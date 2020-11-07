package xutils

import (
	"math/rand"
	"sync"
)

var aliasMethodsPool = sync.Pool{
	New: func() interface{} {
		return &AliasMethods{}
	},
}

// RandIndexWithProb 通过概率获取随机索引
func RandIndexWithProb(prob []float32) int {
	switch len(prob) {
	default:
		as := aliasMethodsPool.Get().(*AliasMethods)
		as.InitWithProb(prob)
		rdx := as.RandIndex()
		aliasMethodsPool.Put(as)
		return rdx
	case 0:
		return -1
	case 1:
		if rand.Float32() < prob[0] {
			return 0
		}
		return -1
	}
}

// RandIndexWithWeight 通过权重获取随机索引
func RandIndexWithWeight(weight []int32) int {
	switch len(weight) {
	default:
		as := aliasMethodsPool.Get().(*AliasMethods)
		as.InitWithWeight(weight)
		rdx := as.RandIndex()
		aliasMethodsPool.Put(as)
		return rdx
	case 0:
		return -1
	case 1:
		return 0
	}
}

// AliasMethods 别名采样
type AliasMethods struct {
	prob  []float32
	alias []int
	sms   []int
	bgs   []int
}

// RandIndex 随机产生索引
func (a *AliasMethods) RandIndex() int {
	if len(a.prob) == 0 {
		return -1
	}
	idx := rand.Intn(len(a.prob))
	if rand.Float32() > a.prob[idx] {
		idx = a.alias[idx]
	}
	return idx
}

// InitWithWeight 能过权重初始化采样表
func (a *AliasMethods) InitWithWeight(weights []int32) {
	if len(weights) == 0 {
		return
	}

	var (
		prob  = make([]float32, len(weights))
		total = float32(0)
	)
	for i, v := range weights {
		prob[i] = float32(v)
		total += prob[i]
	}
	for i := range prob {
		prob[i] /= total
	}

	a.InitWithProb(prob)
}

// InitWithProb 通过概率初始化采样表
func (a *AliasMethods) InitWithProb(prob []float32) {
	l := len(prob)
	if l == 0 {
		return
	}
	if cap(a.prob) >= l {
		a.prob = a.prob[:l]
		a.alias = a.alias[:l]
		a.sms = a.sms[:0]
		a.bgs = a.bgs[:0]
	} else {
		a.prob = make([]float32, l)
		a.alias = make([]int, l)
		a.sms = make([]int, 0, l)
		a.bgs = make([]int, 0, 1)
	}

	lf := float32(l)
	for i := 0; i < l; i++ {
		prob[i] *= lf
		if prob[i] <= 1 {
			a.sms = append(a.sms, i)
		} else {
			a.bgs = append(a.bgs, i)
		}
		a.prob[i] = -1
		a.alias[i] = -1
	}

	for len(a.sms) > 0 && len(a.bgs) > 0 {
		sdx, bdx := a.sms[0], a.bgs[0]
		a.sms = append(a.sms[:0], a.sms[1:]...)
		a.bgs = append(a.bgs[:0], a.bgs[1:]...)
		a.prob[sdx] = prob[sdx]
		a.alias[sdx] = bdx
		prob[bdx] += prob[sdx] - 1
		if prob[bdx] <= 1 {
			a.sms = append(a.sms, bdx)
		} else {
			a.bgs = append(a.bgs, bdx)
		}
	}

	for _, idx := range a.sms {
		a.prob[idx] = prob[idx]
	}
	for _, idx := range a.bgs {
		a.prob[idx] = prob[idx]
	}
}

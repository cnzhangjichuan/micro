package xutils

import "math/rand"

// AliasSample 别名采样
type AliasSample struct {
	prob  []float32
	alias []int
}

// RandIndex 随机产生索引
func (a *AliasSample) RandIndex() int {
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
func (a *AliasSample) InitWithWeight(weights []int32) {
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
func (a *AliasSample) InitWithProb(prob []float32) {
	l := len(prob)
	if l == 0 {
		return
	}
	a.prob = make([]float32, l)
	a.alias = make([]int, l)
	sms := make([]int, 0, l)
	bgs := make([]int, 0, 1)
	for i := 0; i < l; i++ {
		a.alias[i] = -1
	}

	lf := float32(l)
	for i := 0; i < l; i++ {
		prob[i] *= lf
		if prob[i] <= 1 {
			sms = append(sms, i)
		} else {
			bgs = append(bgs, i)
		}
	}

	var sdx, bdx int
	for len(sms) > 0 && len(bgs) > 0 {
		sdx, bdx = sms[0], bgs[0]
		sms = append(sms[:0], sms[1:]...)
		bgs = append(bgs[:0], bgs[1:]...)
		a.prob[sdx] = prob[sdx]
		a.alias[sdx] = bdx
		prob[bdx] += prob[sdx] - 1
		if prob[bdx] <= 1 {
			sms = append(sms, bdx)
		} else {
			bgs = append(bgs, bdx)
		}
	}

	for _, idx := range sms {
		a.prob[idx] = prob[idx]
	}
	for _, idx := range bgs {
		a.prob[idx] = prob[idx]
	}
}

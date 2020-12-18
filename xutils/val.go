package xutils

// IncreaseTimes 获取数值的增长次数
func IncreaseTimes(cur *int64, max, step int64) (times int32) {
	if max <= *cur {
		times = 0
		*cur = max
		return
	}

	offset := max - *cur
	t := offset / step
	if t < 1 {
		times = 0
		return
	}

	times = int32(t)
	m := offset % step
	*cur = max - m
	return
}

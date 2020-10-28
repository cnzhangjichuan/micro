package xutils

import "time"

// Days 距离1970-01-01日的天数
func Days(t time.Time) int {
	const DAY = 86400

	return int(t.Unix() / DAY)
}

// DaysNow 当前距离1970-01-01日的天数
func DaysNow() int {
	return Days(time.Now())
}

// 当前时间
func Now() time.Time {
	return time.Now()
}

// NowSec 当前秒数
func NowSec() int64 {
	return time.Now().Unix()
}

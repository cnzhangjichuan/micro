package xutils

import "time"

// Days 距离1970-01-01日的天数
func Days(t time.Time) int {
	return DaysWithStamp(t.Unix())
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

// DaysWithStamp 距离1970-01-01日的天数
func DaysWithStamp(stamp int64) int {
	const DAY = 86400

	return int(stamp / DAY)
}

// WeeksWithStamp 距离1970-01-01日的周数
func WeeksWithStamp(stamp int64) int {
	const WEEK = 604800 //86400*7

	return int(stamp / WEEK)
}
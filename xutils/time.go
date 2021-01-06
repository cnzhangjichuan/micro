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

// DaysWithDateString 距离1970-01-01日的天数
func DaysWithDateString(value string) int {
	const layout = `2006/01/02`

	t, err := time.ParseInLocation(layout, value, time.UTC)
	if err != nil {
		return 0
	}
	return DaysWithStamp(t.Unix())
}

// ParseTime 时间转换
func ParseTime(value string) (int64, error) {
	const layout = `2006/01/02 15:04`

	return ParseTimeWithLayout(layout, value)
}

// ParseTimeWithLayout 时间转换
func ParseTimeWithLayout(layout, value string) (int64, error) {
	t, err := time.ParseInLocation(layout, value, time.UTC)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// DayInMonth 月中第几天
func DaysOfMonth() int {
	return time.Now().Day()
}

// SETime 起止时段
type SETime struct {
	s int64
	e int64
}

// Parse 从字符串中解析
// s格式[yyyy/MM/dd HH:mm,...]
func (t *SETime) Parse(s string) {
	ss := SplitN(s)
	if len(ss) > 0 {
		t.s, _ = ParseTime(ss[0])
	}
	if len(ss) > 1 {
		t.e, _ = ParseTime(ss[1])
	}
}

// Contains 是否在包含指定的时间
func (t *SETime) Contains(at int64) (bool, bool) {
	if t.s > 0 && at < t.s {
		return false, false
	}
	if t.e > 0 && at >= t.e {
		return true, false
	}
	return false, true
}
package xutils

import "time"

var (
	zero        = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	offsetInDay = 0 - zero.Unix()
)

const (
	daySec  = 86400
	weekSec = 604800
)

// DaysNow 已过天数
func DaysNow() int32 {
	return DaysWithStamp(NowSec())
}

// Days 已过天数
func Days(t time.Time) int32 {
	return DaysWithStamp(t.Unix())
}

// DaysWithStamp 已过天数
func DaysWithStamp(stamp int64) int32 {
	return int32(stamp / daySec)
}

// DaysWithDateString 已过天数
func DaysWithDateString(value string) int32 {
	const layout = `2006/01/02`

	t, err := ParseTimeWithLayout(layout, value)
	if err != nil {
		return 0
	}
	return DaysWithStamp(t)
}

// 当前时间
func Now() time.Time {
	return time.Now()
}

// NowSec 当前秒数
func NowSec() int64 {
	return Now().Unix() + offsetInDay
}

// TodaySec 今天已过的秒数
func TodaySec() int32 {
	return int32(NowSec() % daySec)
}

// TodayDuration 今天余下的秒数
func TodayDuration() int32 {
	return daySec-TodaySec()
}

// WeekNow 当前周
func WeekNow() int32 {
	return WeeksWithStamp(NowSec())
}

// WeekDays 当前周几
func WeekDays() int32 {
	return int32(NowSec()%weekSec)/daySec + 1
}

// WeeksWithStamp 已过周数
func WeeksWithStamp(stamp int64) int32 {
	return int32(stamp / weekSec)
}

// ParseTime 时间转换
func ParseTime(value string) (int64, error) {
	const layout = `2006/01/02 15:04`

	return ParseTimeWithLayout(layout, value)
}

// ParseTimeWithLayout 时间转换
func ParseTimeWithLayout(layout, value string) (int64, error) {
	t, err := time.ParseInLocation(layout, value, time.Local)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// FormatNow 当前日期YMD
func FormatNow() string {
	const layout = `2006/01/02`

	return time.Now().Format(layout)
}

// DayInMonth 月中第几天
func DaysOfMonth() int32 {
	return int32(Now().Day())
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
func (t *SETime) Contains(at int64) bool {
	if t.s > 0 && at < t.s {
		return false
	}
	if t.e > 0 && at >= t.e {
		return false
	}
	return true
}

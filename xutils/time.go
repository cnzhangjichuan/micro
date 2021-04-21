package xutils

import "time"

// 开服时间
var openingTime = time.Date(1970, 1, 1, 6, 0, 0, 0, time.Local)

// SetOpeningTime 设置开服时间
func SetOpeningTime(year, month, day, hour int) {
	openingTime = time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.Local)
}

const (
	daySec       = 86400
	dayDuration  = time.Hour * 24
	weekDuration = 7 * dayDuration
)

// 当前时间
func Now() time.Time {
	return time.Now()
}

// Days 已过天数
func Days(t time.Time) int32 {
	return int32(t.Sub(openingTime) / dayDuration)
}

// DaysNow 当前已过天数
func DaysNow() int32 {
	return int32(Now().Sub(openingTime) / dayDuration)
}

// DaysWithDateString 已过天数
func DaysWithDateString(value string) int32 {
	const layout = `2006/01/02`

	t, err := time.Parse(layout, value)
	if err != nil {
		return 0
	}
	return Days(t)
}

// Weeks 已过周数
func Weeks(t time.Time) int32 {
	return int32(t.Sub(openingTime) / weekDuration)
}

// WeeksNow 当前已过周数
func WeeksNow() int32 {
	return int32(Now().Sub(openingTime) / weekDuration)
}

// NowSec 当前秒数
func NowSec() int64 {
	return Now().Unix()
}

// TodaySec 今天已过的秒数
func TodaySec() int32 {
	now := Now()
	return int32(now.Hour()*3600 + now.Minute()*60 + now.Second())
}

// SecondSineDay 从过去指定的天数始，距离现在已过的秒数
func SecondSineDay(day int32) int32 {
	return int32(NowSec() - openingTime.Add(time.Duration(day)*dayDuration).Unix())
}

// SecondToNextOpeningTime 距离下一个开放时间点的秒数
func SecondToNextOpeningTime(day int32) (sec int32) {
	nowSec := TodaySec()
	refreshAt := int32(openingTime.Hour()*3600 + openingTime.Minute()*60 + openingTime.Second())
	if day <= 0 {
		if nowSec <= refreshAt {
			sec = refreshAt - nowSec
			return
		}
		sec = refreshAt + daySec - nowSec
		return
	}
	sec = refreshAt + daySec - nowSec
	sec += daySec * (day - 1)
	return
}

// ParseTime 时间转换
func ParseTime(value string) (int64, error) {
	const layout = `2006/01/02 15:04`
	t, err := time.Parse(layout, value)
	if err != nil {
		return 0, err
	}
	return t.Unix(), err
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

// MonthOfYear 一年中的月份
func MonthOfYear() int32 {
	return int32(Now().Month())
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

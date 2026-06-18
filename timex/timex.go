package timex

import (
	"fmt"
	"time"
)

const (
	// DateTimeMillisISO 日期时间到毫秒，ISO 风格但不带时区。
	DateTimeMillisISO = "2006-01-02T15:04:05.000"
	// DateTimeISO 日期时间，ISO 风格但不带时区。
	DateTimeISO = "2006-01-02T15:04:05"
	// DateTimeCommon 常见日期时间格式（空格分隔）。
	DateTimeCommon = "2006-01-02 15:04:05"
	// DateTimeMillisCommon 常见日期时间格式（空格分隔，毫秒精度）。
	DateTimeMillisCommon = "2006-01-02 15:04:05.000"
	// DateOnly 仅日期格式。
	DateOnly = "2006-01-02"
	// DateOnlySlash 仅日期格式（斜杠分隔）。
	DateOnlySlash = "2006/01/02"
	// DateCompact 紧凑日期格式。
	DateCompact = "20060102"
	// DateTimeSlashCommon 常见斜杠日期时间格式。
	DateTimeSlashCommon = "2006/01/02 - 15:04:05"
	// DateTimeCompact 紧凑日期时间格式。
	DateTimeCompact = "20060102150405"
	// TimeOnly 仅时间格式。
	TimeOnly = "15:04:05"
	// TimeOnlyMillis 仅时间格式（毫秒精度）。
	TimeOnlyMillis = "15:04:05.000"
	// TimeOnlyNano 仅时间格式（纳秒精度）。
	TimeOnlyNano = "15:04:05.000000000"
)

var defaultParseLayouts = []string{
	time.RFC3339,
	DateTimeMillisISO,
	DateTimeISO,
	DateTimeMillisCommon,
	DateTimeCommon,
	DateOnly,
	DateOnlySlash,
	DateTimeCompact,
}

// ParseAny 依次尝试 layouts 解析时间字符串，全部失败时返回错误。
// layouts 为空时使用包内常用默认格式。
func ParseAny(value string, layouts ...string) (time.Time, error) {
	if len(layouts) == 0 {
		layouts = defaultParseLayouts
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("timex: cannot parse %q", value)
}

// StartOfDay 返回当天 00:00:00.000000000，保留原时区。
func StartOfDay(t time.Time) time.Time {
	return atDayTime(t, 0, 0, 0, 0)
}

// EndOfDay 返回当天 23:59:59.999999999，保留原时区。
func EndOfDay(t time.Time) time.Time {
	return atDayTime(t, 23, 59, 59, int(time.Second-time.Nanosecond))
}

// StartOfMonth 返回当月第一天 00:00:00，保留原时区。
func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 返回当月最后一天 23:59:59.999999999，保留原时区。
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// StartOfYear 返回当年第一天 00:00:00，保留原时区。
func StartOfYear(t time.Time) time.Time {
	y, _, _ := t.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear 返回当年最后一天 23:59:59.999999999，保留原时区。
func EndOfYear(t time.Time) time.Time {
	return StartOfYear(t).AddDate(1, 0, 0).Add(-time.Nanosecond)
}

// DaysBetween 计算从 a 到 b 相差的完整日历天数（b 在 a 之后时为正）。
// 比较基于各自时区下的年月日，忽略时分秒。
func DaysBetween(a, b time.Time) int {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	aDay := time.Date(ay, am, ad, 0, 0, 0, 0, time.UTC)
	bDay := time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	return int(bDay.Sub(aDay).Hours() / 24)
}

// Between 判断 t 是否在 [start, end] 闭区间内。
func Between(t, start, end time.Time) bool {
	return InRange(t, start, end, true, true)
}

// InRange 判断 t 是否在 (start, end) 区间内，可通过 startInclusive/endInclusive 控制端点是否包含。
func InRange(t, start, end time.Time, startInclusive, endInclusive bool) bool {
	if startInclusive {
		if t.Before(start) {
			return false
		}
	} else if !t.After(start) {
		return false
	}
	if endInclusive {
		return !t.After(end)
	}
	return t.Before(end)
}

// IsSameDay 判断两个时间是否在同一天（同时区下年月日相同）。
func IsSameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// IsSameMonth 判断两个时间是否在同一个月（同时区下年月相同）。
func IsSameMonth(a, b time.Time) bool {
	ay, am, _ := a.Date()
	by, bm, _ := b.Date()
	return ay == by && am == bm
}

// StartOfWeek 返回当周周一 00:00:00（ISO 8601，周一为一周起始），保留原时区。
func StartOfWeek(t time.Time) time.Time {
	start := StartOfDay(t)
	weekday := int(start.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return start.AddDate(0, 0, -(weekday - 1))
}

// EndOfWeek 返回当周周日 23:59:59.999999999，保留原时区。
func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// FormatAny 按 layout 格式化时间；layouts 为空时根据时间精度自动选择常用 layout。
func FormatAny(t time.Time, layouts ...string) string {
	if len(layouts) > 0 {
		return t.Format(layouts[0])
	}
	return t.Format(pickFormatLayout(t))
}

func pickFormatLayout(t time.Time) string {
	h, m, s := t.Clock()
	if h == 0 && m == 0 && s == 0 && t.Nanosecond() == 0 {
		return DateOnly
	}
	if t.Nanosecond() == 0 {
		return DateTimeCommon
	}
	if t.Nanosecond()%1_000_000 == 0 {
		return DateTimeMillisCommon
	}
	return DateTimeCommon
}

func atDayTime(t time.Time, hour, min, sec, nsec int) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, hour, min, sec, nsec, t.Location())
}

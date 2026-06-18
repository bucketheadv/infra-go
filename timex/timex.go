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

// Between 判断 t 是否在 [start, end] 闭区间内。
func Between(t, start, end time.Time) bool {
	return !t.Before(start) && !t.After(end)
}

func atDayTime(t time.Time, hour, min, sec, nsec int) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, hour, min, sec, nsec, t.Location())
}

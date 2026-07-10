package timezone

import "time"

var TimeZones = []*time.Location{
	time.FixedZone("UTC-12:00", -12*3600),
	time.FixedZone("UTC-11:00", -11*3600),
	time.FixedZone("UTC-10:00", -10*3600),
	time.FixedZone("UTC-09:00", -9*3600),
	time.FixedZone("UTC-08:00", -8*3600),
	time.FixedZone("UTC-07:00", -7*3600),
	time.FixedZone("UTC-06:00", -6*3600),
	time.FixedZone("UTC-05:00", -5*3600),
	time.FixedZone("UTC-04:00", -4*3600),
	time.FixedZone("UTC-03:00", -3*3600),
	time.FixedZone("UTC-02:00", -2*3600),
	time.FixedZone("UTC-01:00", -1*3600),
	time.FixedZone("UTC+00:00", 0),
	time.FixedZone("UTC+01:00", 3600),
	time.FixedZone("UTC+02:00", 2*3600),
	time.FixedZone("UTC+03:00", 3*3600),
	time.FixedZone("UTC+04:00", 4*3600),
	time.FixedZone("UTC+04:30", 4*3600+1800),
	time.FixedZone("UTC+05:00", 5*3600),
	time.FixedZone("UTC+05:30", 5*3600+1800),
	time.FixedZone("UTC+05:45", 5*3600+2700),
	time.FixedZone("UTC+06:00", 6*3600),
	time.FixedZone("UTC+06:30", 6*3600+1800),
	time.FixedZone("UTC+07:00", 7*3600),
	time.FixedZone("UTC+08:00", 8*3600),
	time.FixedZone("UTC+09:00", 9*3600),
	time.FixedZone("UTC+09:30", 9*3600+1800),
	time.FixedZone("UTC+10:00", 10*3600),
	time.FixedZone("UTC+11:00", 11*3600),
	time.FixedZone("UTC+12:00", 12*3600),
	time.FixedZone("UTC+13:00", 13*3600),
}

var timeZoneMap = map[string]*time.Location{}

func init() {
	for _, zone := range TimeZones {
		timeZoneMap[zone.String()] = zone
	}
}

// GetTimeZone 按固定偏移名（如 "UTC+08:00"）查询时区；未找到返回 nil。
func GetTimeZone(name string) *time.Location {
	return timeZoneMap[name]
}

// Lookup 查询时区：先匹配固定偏移名，再尝试 IANA 名称（如 "Asia/Shanghai"）。
func Lookup(name string) (*time.Location, bool) {
	if loc := GetTimeZone(name); loc != nil {
		return loc, true
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, false
	}
	return loc, true
}

// WithZone 将同一时刻转换到目标时区表示（保留 instant）。
func WithZone(now time.Time, location *time.Location) time.Time {
	if location == nil {
		return now
	}
	return now.In(location)
}

// WithZoneRetainFields 仅替换时区标签，年月日时分秒字段保持不变。
func WithZoneRetainFields(now time.Time, location *time.Location) time.Time {
	if location == nil {
		return now
	}
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), location)
}

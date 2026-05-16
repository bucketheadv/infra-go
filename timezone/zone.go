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

func GetTimeZone(name string) *time.Location {
	return timeZoneMap[name]
}

// WithZone 时区转换，并计算对应大区的时间
func WithZone(now time.Time, location *time.Location) time.Time {
	if location == nil {
		return now
	}
	return now.In(location)
}

// WithZoneRetainFields 仅时区转换，时间不变
func WithZoneRetainFields(now time.Time, location *time.Location) time.Time {
	if location == nil {
		return now
	}
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), location)
}

package timezone

import (
	"testing"
	"time"
)

func TestGetTimeZone(t *testing.T) {
	for _, zone := range TimeZones {
		got := GetTimeZone(zone.String())
		if got == nil {
			t.Fatalf("expect zone %s exists in map", zone.String())
		}
		if got.String() != zone.String() {
			t.Fatalf("zone mismatch: got=%s want=%s", got.String(), zone.String())
		}
	}

	if got := GetTimeZone("NonExistentTimeZone"); got != nil {
		t.Fatalf("expect nil for unknown zone, got=%s", got.String())
	}
}

func TestLookup(t *testing.T) {
	loc, ok := Lookup("UTC+08:00")
	if !ok || loc == nil {
		t.Fatal("Lookup fixed zone failed")
	}
	loc, ok = Lookup("Asia/Shanghai")
	if !ok || loc == nil {
		t.Fatal("Lookup IANA zone failed")
	}
	if _, ok := Lookup("Not/A/Zone"); ok {
		t.Fatal("Lookup unknown should fail")
	}
}

func TestInitTimeZoneMapSize(t *testing.T) {
	if len(timeZoneMap) != len(TimeZones) {
		t.Fatalf("map size mismatch: got=%d want=%d", len(timeZoneMap), len(TimeZones))
	}
}

func TestWithZone(t *testing.T) {
	utc := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	target := GetTimeZone("UTC+08:00")
	if target == nil {
		t.Fatalf("target zone should not be nil")
	}

	got := WithZone(utc, target)

	if got.Location().String() != "UTC+08:00" {
		t.Fatalf("unexpected location: %s", got.Location().String())
	}
	if got.Hour() != 21 || got.Minute() != 0 {
		t.Fatalf("unexpected local time: %s", got.Format("15:04:05"))
	}
	if got.Unix() != utc.Unix() {
		t.Fatalf("WithZone should keep instant")
	}
}

func TestWithZoneRetainFields(t *testing.T) {
	utc := time.Date(2026, 5, 16, 13, 0, 0, 123, time.UTC)
	target := GetTimeZone("UTC+08:00")
	if target == nil {
		t.Fatalf("target zone should not be nil")
	}

	got := WithZoneRetainFields(utc, target)

	if got.Location().String() != "UTC+08:00" {
		t.Fatalf("unexpected location: %s", got.Location().String())
	}
	if got.Year() != 2026 || got.Month() != 5 || got.Day() != 16 {
		t.Fatalf("date fields should be retained")
	}
	if got.Hour() != 13 || got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 123 {
		t.Fatalf("time fields should be retained, got=%s", got.Format("15:04:05.000000000"))
	}
	if got.Unix() == utc.Unix() {
		t.Fatalf("WithZoneRetainFields should change instant in most cases")
	}
}

func TestWithZoneNilLocation(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	got := WithZone(now, nil)
	if !got.Equal(now) || got.Location() != now.Location() {
		t.Fatalf("nil location should keep original time, got=%v want=%v", got, now)
	}
}

func TestWithZoneRetainFieldsNilLocation(t *testing.T) {
	now := time.Date(2026, 5, 16, 13, 0, 0, 0, time.UTC)
	got := WithZoneRetainFields(now, nil)
	if !got.Equal(now) || got.Location() != now.Location() {
		t.Fatalf("nil location should keep original time, got=%v want=%v", got, now)
	}
}

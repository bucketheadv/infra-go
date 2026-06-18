package timex

import (
	"testing"
	"time"
)

func TestParseAny(t *testing.T) {
	t.Run("without layouts", func(t *testing.T) {
		cases := []struct {
			name  string
			input string
			want  string
		}{
			{name: "RFC3339", input: "2026-05-17T12:30:45Z", want: "2026-05-17 12:30:45"},
			{name: "DateTimeMillisISO", input: "2026-05-17T12:30:45.123", want: "2026-05-17 12:30:45"},
			{name: "DateTimeISO", input: "2026-05-17T12:30:45", want: "2026-05-17 12:30:45"},
			{name: "DateTimeMillisCommon", input: "2026-05-17 12:30:45.123", want: "2026-05-17 12:30:45"},
			{name: "DateTimeCommon", input: "2026-05-17 12:30:45", want: "2026-05-17 12:30:45"},
			{name: "DateOnly", input: "2026-05-17", want: "2026-05-17 00:00:00"},
			{name: "DateOnlySlash", input: "2026/05/17", want: "2026-05-17 00:00:00"},
			{name: "DateTimeCompact", input: "20260517123045", want: "2026-05-17 12:30:45"},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got, err := ParseAny(tc.input)
				if err != nil {
					t.Fatalf("ParseAny(%q) err=%v", tc.input, err)
				}
				if got.UTC().Format(DateTimeCommon) != tc.want {
					t.Fatalf("ParseAny(%q) = %v, want %s", tc.input, got, tc.want)
				}
			})
		}
	})

	t.Run("with custom layout", func(t *testing.T) {
		got, err := ParseAny("17/05/2026", "02/01/2006")
		if err != nil {
			t.Fatalf("ParseAny() err=%v", err)
		}
		if got.UTC().Format(DateOnly) != "2026-05-17" {
			t.Fatalf("ParseAny() = %v", got)
		}
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := ParseAny("invalid")
		if err == nil {
			t.Fatalf("ParseAny() should fail for invalid input")
		}
	})
}

func TestBoundary(t *testing.T) {
	loc := time.UTC
	tm := time.Date(2026, 5, 17, 15, 30, 45, 123, loc)

	start := StartOfDay(tm)
	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Fatalf("StartOfDay() = %v", start)
	}

	end := EndOfDay(tm)
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Fatalf("EndOfDay() = %v", end)
	}

	mStart := StartOfMonth(tm)
	if mStart.Day() != 1 {
		t.Fatalf("StartOfMonth() = %v", mStart)
	}

	mEnd := EndOfMonth(tm)
	if mEnd.Month() != 5 || mEnd.Day() != 31 {
		t.Fatalf("EndOfMonth() = %v", mEnd)
	}

	yStart := StartOfYear(tm)
	if yStart.Month() != time.January || yStart.Day() != 1 {
		t.Fatalf("StartOfYear() = %v", yStart)
	}
	yEnd := EndOfYear(tm)
	if yEnd.Month() != time.December || yEnd.Day() != 31 {
		t.Fatalf("EndOfYear() = %v", yEnd)
	}

	if DaysBetween(
		time.Date(2026, 5, 17, 23, 0, 0, 0, loc),
		time.Date(2026, 5, 20, 1, 0, 0, 0, loc),
	) != 3 {
		t.Fatalf("DaysBetween() mismatch")
	}
	if DaysBetween(
		time.Date(2026, 5, 20, 0, 0, 0, 0, loc),
		time.Date(2026, 5, 17, 0, 0, 0, 0, loc),
	) != -3 {
		t.Fatalf("DaysBetween() negative mismatch")
	}

	if !Between(tm, start, end) {
		t.Fatalf("Between() should be true inside same day")
	}
	if Between(tm.Add(24*time.Hour), start, end) {
		t.Fatalf("Between() should be false outside range")
	}
}

func TestFormatAnyAndCompare(t *testing.T) {
	loc := time.UTC
	day := time.Date(2026, 5, 17, 0, 0, 0, 0, loc)
	dt := time.Date(2026, 5, 17, 12, 30, 45, 0, loc)

	if FormatAny(day) != "2026-05-17" {
		t.Fatalf("FormatAny(day) = %q", FormatAny(day))
	}
	if FormatAny(dt, DateTimeISO) != "2026-05-17T12:30:45" {
		t.Fatalf("FormatAny(custom) = %q", FormatAny(dt, DateTimeISO))
	}

	a := time.Date(2026, 5, 17, 10, 0, 0, 0, loc)
	b := time.Date(2026, 5, 17, 20, 0, 0, 0, loc)
	c := time.Date(2026, 5, 1, 0, 0, 0, 0, loc)
	if !IsSameDay(a, b) || IsSameDay(a, time.Date(2026, 6, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("IsSameDay mismatch")
	}
	if !IsSameMonth(a, c) || IsSameMonth(b, time.Date(2026, 4, 1, 0, 0, 0, 0, loc)) {
		t.Fatalf("IsSameMonth mismatch")
	}

	weekMid := time.Date(2026, 5, 14, 15, 0, 0, 0, loc) // Thursday
	wStart := StartOfWeek(weekMid)
	if wStart.Weekday() != time.Monday || wStart.Day() != 11 {
		t.Fatalf("StartOfWeek() = %v", wStart)
	}
	wEnd := EndOfWeek(weekMid)
	if wEnd.Weekday() != time.Sunday || wEnd.Day() != 17 {
		t.Fatalf("EndOfWeek() = %v", wEnd)
	}

	mid := time.Date(2026, 5, 17, 12, 0, 0, 0, loc)
	start := time.Date(2026, 5, 17, 12, 0, 0, 0, loc)
	end := time.Date(2026, 5, 17, 13, 0, 0, 0, loc)
	if !InRange(mid, start, end, true, true) {
		t.Fatalf("InRange closed failed")
	}
	if InRange(start, start, end, false, true) {
		t.Fatalf("InRange open start should exclude start")
	}
}

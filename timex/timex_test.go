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

	if !Between(tm, start, end) {
		t.Fatalf("Between() should be true inside same day")
	}
	if Between(tm.Add(24*time.Hour), start, end) {
		t.Fatalf("Between() should be false outside range")
	}
}

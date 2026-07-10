package timex

import (
	"fmt"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		lang Language
		want string
	}{
		{"zh zero", 0, LangZH, "0秒"},
		{"zh composite", 2*time.Hour + 30*time.Minute, LangZH, "2小时30分钟"},
		{"ja zero", 0, LangJA, "0秒"},
		{"ja composite", 2*time.Hour + 5*time.Minute, LangJA, "2時間5分"},
		{"ko zero", 0, LangKO, "0초"},
		{"ko composite", 1*time.Hour + 30*time.Second, LangKO, "1시간30초"},
		{"en zero", 0, LangEN, "0 seconds"},
		{"en composite", 90 * time.Second, LangEN, "1 minute 30 seconds"},
		{"en millis", 500 * time.Millisecond, LangEN, "500 milliseconds"},
		{"es zero", 0, LangES, "0 segundos"},
		{"es composite", 2*time.Hour + 30*time.Minute, LangES, "2 horas 30 minutos"},
		{"es seconds", 90 * time.Second, LangES, "1 minuto 30 segundos"},
		{"de zero", 0, LangDE, "0 Sekunden"},
		{"de composite", 2*time.Hour + 30*time.Minute, LangDE, "2 Stunden 30 Minuten"},
		{"de seconds", 90 * time.Second, LangDE, "1 Minute 30 Sekunden"},
		{"fallback unknown lang", 2 * time.Hour, Language("fr"), "2 hours"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := FormatDuration(tc.d, tc.lang)
			if got != tc.want {
				t.Fatalf("FormatDuration() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDurationFormatterRegisterAndUnregister(t *testing.T) {
	formatter := NewDurationFormatter(LangEN, map[Language]DurationLocale{
		LangEN: durationLocalePlural("0 seconds", map[DurationUnit][2]string{
			DurationHour: {"hour", "hours"},
		}),
	})

	const langFR Language = "fr"
	formatter.Register(langFR, durationLocalePlural("0 secondes", map[DurationUnit][2]string{
		DurationHour: {"heure", "heures"},
	}))

	got := formatter.FormatDuration(2*time.Hour, langFR)
	if got != "2 heures" {
		t.Fatalf("custom locale = %q, want %q", got, "2 heures")
	}

	formatter.Unregister(langFR)
	got = formatter.FormatDuration(2*time.Hour, langFR)
	if got != "2 hours" {
		t.Fatalf("fallback locale = %q, want %q", got, "2 hours")
	}
}

func TestRegisterDurationLocaleOnDefaultFormatter(t *testing.T) {
	const langFR Language = "fr"
	RegisterDurationLocale(langFR, DurationLocale{
		Zero: "0 secondes",
		FormatUnit: func(count int64, unit DurationUnit) string {
			return fmt.Sprintf("%d heures", count)
		},
		JoinParts: func(parts []string) string {
			return parts[0]
		},
	})
	t.Cleanup(func() {
		UnregisterDurationLocale(langFR)
	})

	got := FormatDuration(2*time.Hour, langFR)
	if got != "2 heures" {
		t.Fatalf("default formatter custom locale = %q, want %q", got, "2 heures")
	}
}

func TestDaysBetweenTimezone(t *testing.T) {
	loc := time.FixedZone("UTC+09:00", 9*3600)
	a := time.Date(2026, 2, 28, 23, 0, 0, 0, time.UTC)
	b := time.Date(2026, 3, 1, 8, 0, 0, 0, loc) // same instant as a
	if !IsSameDay(a, b) {
		t.Fatal("IsSameDay should be true for same instant")
	}
	if DaysBetween(a, b) != 0 {
		t.Fatalf("DaysBetween should be 0, got %d", DaysBetween(a, b))
	}
}

func TestNowIn(t *testing.T) {
	loc := time.FixedZone("UTC+08:00", 8*3600)
	now := NowIn(loc)
	if now.Location().String() != loc.String() {
		t.Fatalf("NowIn() location = %v", now.Location())
	}
	if NowIn(nil).Location().String() == "" {
		t.Fatalf("NowIn(nil) should return local time")
	}
}

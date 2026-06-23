package timex

import (
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

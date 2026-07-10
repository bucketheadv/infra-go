package timex

import (
	"testing"
	"time"
)

func TestFormatRelativeSince(t *testing.T) {
	base := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		t    time.Time
		lang Language
		want string
	}{
		// ZH Past
		{"zh just now", base.Add(-30 * time.Second), LangZH, "刚刚"},
		{"zh 1 min ago", base.Add(-1 * time.Minute), LangZH, "1分钟前"},
		{"zh 5 mins ago", base.Add(-5 * time.Minute), LangZH, "5分钟前"},
		{"zh 1 hour ago", base.Add(-1 * time.Hour), LangZH, "1小时前"},
		{"zh 2 hours ago", base.Add(-2 * time.Hour), LangZH, "2小时前"},
		{"zh 1 day ago", base.Add(-24 * time.Hour), LangZH, "1天前"},
		{"zh 5 days ago", base.Add(-5 * 24 * time.Hour), LangZH, "5天前"},
		{"zh 1 month ago", base.Add(-30 * 24 * time.Hour), LangZH, "1个月前"},
		{"zh 2 months ago", base.Add(-65 * 24 * time.Hour), LangZH, "2个月前"},
		{"zh 1 year ago", base.Add(-365 * 24 * time.Hour), LangZH, "1年前"},
		{"zh 2 years ago", base.Add(-800 * 24 * time.Hour), LangZH, "2年前"},

		// ZH Future
		{"zh in few secs", base.Add(30 * time.Second), LangZH, "几秒后"},
		{"zh in 1 min", base.Add(1 * time.Minute), LangZH, "1分钟后"},
		{"zh in 5 mins", base.Add(5 * time.Minute), LangZH, "5分钟后"},

		// EN Past
		{"en just now", base.Add(-30 * time.Second), LangEN, "just now"},
		{"en 1 min ago", base.Add(-1 * time.Minute), LangEN, "1 minute ago"},
		{"en 5 mins ago", base.Add(-5 * time.Minute), LangEN, "5 minutes ago"},
		{"en 1 hour ago", base.Add(-1 * time.Hour), LangEN, "1 hour ago"},
		{"en 2 hours ago", base.Add(-2 * time.Hour), LangEN, "2 hours ago"},
		{"en 1 day ago", base.Add(-24 * time.Hour), LangEN, "1 day ago"},
		{"en 5 days ago", base.Add(-5 * 24 * time.Hour), LangEN, "5 days ago"},
		{"en 1 month ago", base.Add(-30 * 24 * time.Hour), LangEN, "1 month ago"},
		{"en 2 months ago", base.Add(-65 * 24 * time.Hour), LangEN, "2 months ago"},
		{"en 1 year ago", base.Add(-365 * 24 * time.Hour), LangEN, "1 year ago"},
		{"en 2 years ago", base.Add(-800 * 24 * time.Hour), LangEN, "2 years ago"},

		// EN Future
		{"en in few secs", base.Add(30 * time.Second), LangEN, "in a few seconds"},
		{"en in 1 min", base.Add(1 * time.Minute), LangEN, "in 1 minute"},
		{"en in 5 mins", base.Add(5 * time.Minute), LangEN, "in 5 minutes"},

		// JA Past
		{"ja just now", base.Add(-30 * time.Second), LangJA, "たった今"},
		{"ja 1 min ago", base.Add(-1 * time.Minute), LangJA, "1分前"},
		{"ja 5 mins ago", base.Add(-5 * time.Minute), LangJA, "5分前"},
		{"ja 1 hour ago", base.Add(-1 * time.Hour), LangJA, "1時間前"},
		{"ja 1 day ago", base.Add(-24 * time.Hour), LangJA, "1日前"},
		{"ja 1 month ago", base.Add(-30 * 24 * time.Hour), LangJA, "1ヶ月前"},
		{"ja 1 year ago", base.Add(-365 * 24 * time.Hour), LangJA, "1年前"},

		// JA Future
		{"ja in few secs", base.Add(30 * time.Second), LangJA, "数秒後"},
		{"ja in 1 min", base.Add(1 * time.Minute), LangJA, "1分後"},

		// KO Past
		{"ko just now", base.Add(-30 * time.Second), LangKO, "방금 전"},
		{"ko 1 min ago", base.Add(-1 * time.Minute), LangKO, "1분 전"},
		{"ko 5 mins ago", base.Add(-5 * time.Minute), LangKO, "5분 전"},
		{"ko 1 hour ago", base.Add(-1 * time.Hour), LangKO, "1시간 전"},
		{"ko 1 day ago", base.Add(-24 * time.Hour), LangKO, "1일 전"},
		{"ko 1 month ago", base.Add(-30 * 24 * time.Hour), LangKO, "1개월 전"},
		{"ko 1 year ago", base.Add(-365 * 24 * time.Hour), LangKO, "1년 전"},

		// KO Future
		{"ko in few secs", base.Add(30 * time.Second), LangKO, "몇 초 후"},
		{"ko in 1 min", base.Add(1 * time.Minute), LangKO, "1분 후"},

		// ES Past
		{"es just now", base.Add(-30 * time.Second), LangES, "justo ahora"},
		{"es 1 min ago", base.Add(-1 * time.Minute), LangES, "hace 1 minuto"},
		{"es 5 mins ago", base.Add(-5 * time.Minute), LangES, "hace 5 minutos"},
		{"es 1 hour ago", base.Add(-1 * time.Hour), LangES, "hace 1 hora"},
		{"es 2 hours ago", base.Add(-2 * time.Hour), LangES, "hace 2 horas"},
		{"es 1 day ago", base.Add(-24 * time.Hour), LangES, "hace 1 día"},
		{"es 5 days ago", base.Add(-5 * 24 * time.Hour), LangES, "hace 5 días"},
		{"es 1 month ago", base.Add(-30 * 24 * time.Hour), LangES, "hace 1 mes"},
		{"es 2 months ago", base.Add(-65 * 24 * time.Hour), LangES, "hace 2 meses"},
		{"es 1 year ago", base.Add(-365 * 24 * time.Hour), LangES, "hace 1 año"},
		{"es 2 years ago", base.Add(-800 * 24 * time.Hour), LangES, "hace 2 años"},

		// ES Future
		{"es in few secs", base.Add(30 * time.Second), LangES, "en unos segundos"},
		{"es in 1 min", base.Add(1 * time.Minute), LangES, "en 1 minuto"},
		{"es in 5 mins", base.Add(5 * time.Minute), LangES, "en 5 minutos"},

		// DE Past
		{"de just now", base.Add(-30 * time.Second), LangDE, "gerade eben"},
		{"de 1 min ago", base.Add(-1 * time.Minute), LangDE, "vor 1 Minute"},
		{"de 5 mins ago", base.Add(-5 * time.Minute), LangDE, "vor 5 Minuten"},
		{"de 1 hour ago", base.Add(-1 * time.Hour), LangDE, "vor 1 Stunde"},
		{"de 2 hours ago", base.Add(-2 * time.Hour), LangDE, "vor 2 Stunden"},
		{"de 1 day ago", base.Add(-24 * time.Hour), LangDE, "vor 1 Tag"},
		{"de 5 days ago", base.Add(-5 * 24 * time.Hour), LangDE, "vor 5 Tagen"},
		{"de 1 month ago", base.Add(-30 * 24 * time.Hour), LangDE, "vor 1 Monat"},
		{"de 2 months ago", base.Add(-65 * 24 * time.Hour), LangDE, "vor 2 Monaten"},
		{"de 1 year ago", base.Add(-365 * 24 * time.Hour), LangDE, "vor 1 Jahr"},
		{"de 2 years ago", base.Add(-800 * 24 * time.Hour), LangDE, "vor 2 Jahren"},

		// DE Future
		{"de in few secs", base.Add(30 * time.Second), LangDE, "in wenigen Sekunden"},
		{"de in 1 min", base.Add(1 * time.Minute), LangDE, "in 1 Minute"},
		{"de in 5 mins", base.Add(5 * time.Minute), LangDE, "in 5 Minuten"},

		// Unknown lang defaults to EN
		{"unknown lang", base.Add(-5 * time.Minute), "fr", "5 minutes ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatRelativeSince(tt.t, base, tt.lang)
			if got != tt.want {
				t.Errorf("FormatRelativeSince() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRelativeFormatterRegisterAndUnregister(t *testing.T) {
	base := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	formatter := NewRelativeFormatter(LangZH, map[Language]RelativeLocale{
		LangZH: localeWithAffix("刚刚", "几秒后", "前", "后", map[RelativePeriod]string{
			PeriodMinute: "分钟",
		}),
	})

	const langFR Language = "fr"
	formatter.Register(langFR, localeWithPlural(
		"à l'instant", "dans quelques secondes",
		"il y a", "dans",
		map[RelativePeriod][2]string{
			PeriodMinute: {"minute", "minutes"},
		},
		false,
	))

	got := formatter.FormatRelativeSince(base.Add(-5*time.Minute), base, langFR)
	if got != "il y a 5 minutes" {
		t.Fatalf("custom locale = %q, want %q", got, "il y a 5 minutes")
	}

	formatter.Unregister(langFR)
	got = formatter.FormatRelativeSince(base.Add(-5*time.Minute), base, langFR)
	if got != "5分钟前" {
		t.Fatalf("fallback locale = %q, want %q", got, "5分钟前")
	}
}

func TestRegisterRelativeLocaleOnDefaultFormatter(t *testing.T) {
	base := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	const langFR Language = "fr"

	RegisterRelativeLocale(langFR, localeWithPlural(
		"à l'instant", "dans quelques secondes",
		"il y a", "dans",
		map[RelativePeriod][2]string{
			PeriodMinute: {"minute", "minutes"},
		},
		false,
	))
	t.Cleanup(func() {
		UnregisterRelativeLocale(langFR)
	})

	got := FormatRelativeSince(base.Add(-1*time.Minute), base, langFR)
	if got != "il y a 1 minute" {
		t.Fatalf("default formatter custom locale = %q, want %q", got, "il y a 1 minute")
	}
}

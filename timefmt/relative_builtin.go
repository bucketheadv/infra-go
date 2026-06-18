package timefmt

import "fmt"

func builtinRelativeLocales() map[Language]RelativeLocale {
	return map[Language]RelativeLocale{
		LangZH: localeWithAffix("刚刚", "几秒后", "前", "后", map[RelativePeriod]string{
			PeriodMinute: "分钟",
			PeriodHour:   "小时",
			PeriodDay:    "天",
			PeriodMonth:  "个月",
			PeriodYear:   "年",
		}),
		LangEN: localeWithSpacedPlural(
			"just now", "in a few seconds",
			"ago", "in",
			map[RelativePeriod][2]string{
				PeriodMinute: {"minute", "minutes"},
				PeriodHour:   {"hour", "hours"},
				PeriodDay:    {"day", "days"},
				PeriodMonth:  {"month", "months"},
				PeriodYear:   {"year", "years"},
			},
		),
		LangJA: localeWithAffix("たった今", "数秒後", "前", "後", map[RelativePeriod]string{
			PeriodMinute: "分",
			PeriodHour:   "時間",
			PeriodDay:    "日",
			PeriodMonth:  "ヶ月",
			PeriodYear:   "年",
		}),
		LangKO: localeWithSpacedAffix("방금 전", "몇 초 후", " 전", " 후", map[RelativePeriod]string{
			PeriodMinute: "분",
			PeriodHour:   "시간",
			PeriodDay:    "일",
			PeriodMonth:  "개월",
			PeriodYear:   "년",
		}),
		LangES: localeWithPrefixPlural(
			"justo ahora", "en unos segundos",
			"hace", "en",
			map[RelativePeriod][2]string{
				PeriodMinute: {"minuto", "minutos"},
				PeriodHour:   {"hora", "horas"},
				PeriodDay:    {"día", "días"},
				PeriodMonth:  {"mes", "meses"},
				PeriodYear:   {"año", "años"},
			},
		),
		LangDE: localeGerman(),
	}
}

func localeWithAffix(justNow, inFewSec, pastAffix, futureAffix string, units map[RelativePeriod]string) RelativeLocale {
	return RelativeLocale{
		JustNow:      justNow,
		InFewSeconds: inFewSec,
		FormatInterval: func(count int64, period RelativePeriod, isFuture bool) string {
			unit := units[period]
			if isFuture {
				return fmt.Sprintf("%d%s%s", count, unit, futureAffix)
			}
			return fmt.Sprintf("%d%s%s", count, unit, pastAffix)
		},
	}
}

func localeWithSpacedAffix(justNow, inFewSec, pastAffix, futureAffix string, units map[RelativePeriod]string) RelativeLocale {
	return RelativeLocale{
		JustNow:      justNow,
		InFewSeconds: inFewSec,
		FormatInterval: func(count int64, period RelativePeriod, isFuture bool) string {
			unit := units[period]
			if isFuture {
				return fmt.Sprintf("%d%s%s", count, unit, futureAffix)
			}
			return fmt.Sprintf("%d%s%s", count, unit, pastAffix)
		},
	}
}

func localeWithSpacedPlural(justNow, inFewSec, pastWord, futureWord string, units map[RelativePeriod][2]string) RelativeLocale {
	return RelativeLocale{
		JustNow:      justNow,
		InFewSeconds: inFewSec,
		FormatInterval: func(count int64, period RelativePeriod, isFuture bool) string {
			unit := pluralUnit(units[period], count)
			if isFuture {
				return fmt.Sprintf("%s %d %s", futureWord, count, unit)
			}
			return fmt.Sprintf("%d %s %s", count, unit, pastWord)
		},
	}
}

func localeWithPrefixPlural(justNow, inFewSec, pastPrefix, futurePrefix string, units map[RelativePeriod][2]string) RelativeLocale {
	return RelativeLocale{
		JustNow:      justNow,
		InFewSeconds: inFewSec,
		FormatInterval: func(count int64, period RelativePeriod, isFuture bool) string {
			unit := pluralUnit(units[period], count)
			if isFuture {
				return fmt.Sprintf("%s %d %s", futurePrefix, count, unit)
			}
			return fmt.Sprintf("%s %d %s", pastPrefix, count, unit)
		},
	}
}

func localeGerman() RelativeLocale {
	units := map[RelativePeriod][2]string{
		PeriodMinute: {"Minute", "Minuten"},
		PeriodHour:   {"Stunde", "Stunden"},
		PeriodDay:    {"Tag", "Tagen"},
		PeriodMonth:  {"Monat", "Monaten"},
		PeriodYear:   {"Jahr", "Jahren"},
	}
	return RelativeLocale{
		JustNow:      "gerade eben",
		InFewSeconds: "in wenigen Sekunden",
		FormatInterval: func(count int64, period RelativePeriod, isFuture bool) string {
			pair := units[period]
			unit := pair[0]
			if count > 1 {
				unit = pair[1]
			}
			if isFuture {
				return fmt.Sprintf("in %d %s", count, unit)
			}
			return fmt.Sprintf("vor %d %s", count, unit)
		},
	}
}

func pluralUnit(pair [2]string, count int64) string {
	if count == 1 {
		return pair[0]
	}
	return pair[1]
}

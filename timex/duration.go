package timex

import (
	"fmt"
	"time"
)

// FormatDuration 将时长格式化为人类可读字符串。
// lang 支持 LangZH、LangEN、LangJA、LangKO、LangES、LangDE；未知语言回退为英文。
func FormatDuration(d time.Duration, lang Language) string {
	if d < 0 {
		d = -d
	}
	loc := durationLocaleFor(lang)
	if d == 0 {
		return loc.zero
	}

	ms := d.Milliseconds()
	days := ms / (24 * 60 * 60 * 1000)
	ms %= 24 * 60 * 60 * 1000
	hours := ms / (60 * 60 * 1000)
	ms %= 60 * 60 * 1000
	minutes := ms / (60 * 1000)
	ms %= 60 * 1000
	seconds := ms / 1000
	millis := ms % 1000

	parts := make([]string, 0, 5)
	if days > 0 {
		parts = append(parts, loc.formatUnit(days, durationDay))
	}
	if hours > 0 {
		parts = append(parts, loc.formatUnit(hours, durationHour))
	}
	if minutes > 0 {
		parts = append(parts, loc.formatUnit(minutes, durationMinute))
	}
	if seconds > 0 {
		parts = append(parts, loc.formatUnit(seconds, durationSecond))
	}
	if millis > 0 && days == 0 && hours == 0 && minutes == 0 && seconds == 0 {
		parts = append(parts, loc.formatUnit(millis, durationMillis))
	}

	if len(parts) == 0 {
		return loc.zero
	}
	return loc.join(parts)
}

type durationPart int

const (
	durationDay durationPart = iota
	durationHour
	durationMinute
	durationSecond
	durationMillis
)

type durationUnit struct {
	compact string
	pair    [2]string
}

type durationLocale struct {
	zero        string
	joinCompact bool
	day         durationUnit
	hour        durationUnit
	minute      durationUnit
	second      durationUnit
	millisecond durationUnit
}

func (l durationLocale) formatUnit(count int64, part durationPart) string {
	u := l.unitFor(part)
	if u.compact != "" {
		return fmt.Sprintf("%d%s", count, u.compact)
	}
	return pluralPair(count, u.pair)
}

func (l durationLocale) unitFor(part durationPart) durationUnit {
	switch part {
	case durationDay:
		return l.day
	case durationHour:
		return l.hour
	case durationMinute:
		return l.minute
	case durationSecond:
		return l.second
	case durationMillis:
		return l.millisecond
	default:
		return l.second
	}
}

func (l durationLocale) join(parts []string) string {
	if l.joinCompact {
		var b string
		for _, p := range parts {
			b += p
		}
		return b
	}
	var b string
	for i, p := range parts {
		if i > 0 {
			b += " "
		}
		b += p
	}
	return b
}

func pluralPair(count int64, pair [2]string) string {
	return fmt.Sprintf("%d %s", count, pluralUnit(pair, count))
}

func durationLocaleFor(lang Language) durationLocale {
	if loc, ok := builtinDurationLocales[lang]; ok {
		return loc
	}
	return builtinDurationLocales[LangEN]
}

func initDurationLocales() map[Language]durationLocale {
	return map[Language]durationLocale{
		LangZH: {
			zero:        "0秒",
			joinCompact: true,
			day:         durationUnit{compact: "天"},
			hour:        durationUnit{compact: "小时"},
			minute:      durationUnit{compact: "分钟"},
			second:      durationUnit{compact: "秒"},
			millisecond: durationUnit{compact: "毫秒"},
		},
		LangEN: {
			zero:        "0 seconds",
			day:         durationUnit{pair: [2]string{"day", "days"}},
			hour:        durationUnit{pair: [2]string{"hour", "hours"}},
			minute:      durationUnit{pair: [2]string{"minute", "minutes"}},
			second:      durationUnit{pair: [2]string{"second", "seconds"}},
			millisecond: durationUnit{pair: [2]string{"millisecond", "milliseconds"}},
		},
		LangJA: {
			zero:        "0秒",
			joinCompact: true,
			day:         durationUnit{compact: "日"},
			hour:        durationUnit{compact: "時間"},
			minute:      durationUnit{compact: "分"},
			second:      durationUnit{compact: "秒"},
			millisecond: durationUnit{compact: "ミリ秒"},
		},
		LangKO: {
			zero:        "0초",
			joinCompact: true,
			day:         durationUnit{compact: "일"},
			hour:        durationUnit{compact: "시간"},
			minute:      durationUnit{compact: "분"},
			second:      durationUnit{compact: "초"},
			millisecond: durationUnit{compact: "밀리초"},
		},
		LangES: {
			zero:        "0 segundos",
			day:         durationUnit{pair: [2]string{"día", "días"}},
			hour:        durationUnit{pair: [2]string{"hora", "horas"}},
			minute:      durationUnit{pair: [2]string{"minuto", "minutos"}},
			second:      durationUnit{pair: [2]string{"segundo", "segundos"}},
			millisecond: durationUnit{pair: [2]string{"milisegundo", "milisegundos"}},
		},
		LangDE: {
			zero:        "0 Sekunden",
			day:         durationUnit{pair: [2]string{"Tag", "Tagen"}},
			hour:        durationUnit{pair: [2]string{"Stunde", "Stunden"}},
			minute:      durationUnit{pair: [2]string{"Minute", "Minuten"}},
			second:      durationUnit{pair: [2]string{"Sekunde", "Sekunden"}},
			millisecond: durationUnit{pair: [2]string{"Millisekunde", "Millisekunden"}},
		},
	}
}

var builtinDurationLocales = initDurationLocales()

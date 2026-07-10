package timex

import (
	"fmt"
	"sync"
	"time"
)

// Language 相对时间语言标识。
type Language string

const (
	LangZH Language = "zh"
	LangEN Language = "en"
	LangJA Language = "ja"
	LangKO Language = "ko"
	LangES Language = "es"
	LangDE Language = "de"
)

// RelativePeriod 相对时间粒度。
type RelativePeriod int

const (
	PeriodMinute RelativePeriod = iota
	PeriodHour
	PeriodDay
	PeriodMonth
	PeriodYear
)

// RelativeLocale 单种语言的相对时间配置。
type RelativeLocale struct {
	// JustNow 刚刚发生时的文案。
	JustNow string
	// InFewSeconds 即将发生（数秒内）的文案。
	InFewSeconds string
	// FormatInterval 格式化相对时间间隔。
	FormatInterval func(count int64, period RelativePeriod, isFuture bool) string
}

// RelativeFormatter 相对时间格式化器，支持动态注册语言。
type RelativeFormatter struct {
	// mu 保护 locales 并发读写。
	mu sync.RWMutex
	// locales 已注册语言配置。
	locales map[Language]RelativeLocale
	// fallback 未命中语言时的回退语言。
	fallback Language
}

// NewRelativeFormatter 创建格式化器；fallback 为未注册语言时的回退语言。
func NewRelativeFormatter(fallback Language, locales map[Language]RelativeLocale) *RelativeFormatter {
	if locales == nil {
		locales = make(map[Language]RelativeLocale)
	}
	return &RelativeFormatter{
		locales:  locales,
		fallback: fallback,
	}
}

// Register 注册或覆盖一种语言配置。
func (f *RelativeFormatter) Register(lang Language, locale RelativeLocale) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.locales[lang] = locale
}

// Unregister 移除一种语言配置。
func (f *RelativeFormatter) Unregister(lang Language) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.locales, lang)
}

// SetFallback 设置未命中语言时的回退语言。
func (f *RelativeFormatter) SetFallback(lang Language) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fallback = lang
}

// Has 判断语言是否已注册。
func (f *RelativeFormatter) Has(lang Language) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.locales[lang]
	return ok
}

// Languages 返回已注册语言列表。
func (f *RelativeFormatter) Languages() []Language {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]Language, 0, len(f.locales))
	for lang := range f.locales {
		out = append(out, lang)
	}
	return out
}

// FormatRelative 以当前时间为基准格式化相对时间。
func (f *RelativeFormatter) FormatRelative(t time.Time, lang Language) string {
	return f.FormatRelativeSince(t, time.Now(), lang)
}

// FormatRelativeSince 以 base 时间为基准格式化相对时间。
// 月/年粒度按 30 天/月、365 天/年近似计算，非精确日历差。
func (f *RelativeFormatter) FormatRelativeSince(t, base time.Time, lang Language) string {
	locale, ok := f.locale(lang)
	if !ok {
		locale, ok = f.locale(f.fallbackLang())
	}
	if !ok || locale.FormatInterval == nil {
		locale = builtinRelativeLocales()[LangEN]
	}

	diff := base.Sub(t)
	isFuture := diff < 0
	if isFuture {
		diff = -diff
	}

	seconds := int64(diff.Seconds())
	if seconds < 60 {
		if isFuture {
			return locale.InFewSeconds
		}
		return locale.JustNow
	}

	minutes := seconds / 60
	if minutes < 60 {
		return locale.FormatInterval(minutes, PeriodMinute, isFuture)
	}

	hours := minutes / 60
	if hours < 24 {
		return locale.FormatInterval(hours, PeriodHour, isFuture)
	}

	days := hours / 24
	if days < 30 {
		return locale.FormatInterval(days, PeriodDay, isFuture)
	}

	months := days / 30
	if months < 12 {
		return locale.FormatInterval(months, PeriodMonth, isFuture)
	}

	years := days / 365
	return locale.FormatInterval(years, PeriodYear, isFuture)
}

func (f *RelativeFormatter) locale(lang Language) (RelativeLocale, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	locale, ok := f.locales[lang]
	return locale, ok
}

func (f *RelativeFormatter) fallbackLang() Language {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fallback
}

var defaultRelativeFormatter = NewRelativeFormatter(LangEN, builtinRelativeLocales())

// DefaultRelativeFormatter 返回默认相对时间格式化器。
func DefaultRelativeFormatter() *RelativeFormatter {
	return defaultRelativeFormatter
}

// RegisterRelativeLocale 向默认格式化器注册语言。
func RegisterRelativeLocale(lang Language, locale RelativeLocale) {
	defaultRelativeFormatter.Register(lang, locale)
}

// UnregisterRelativeLocale 从默认格式化器移除语言。
func UnregisterRelativeLocale(lang Language) {
	defaultRelativeFormatter.Unregister(lang)
}

// FormatRelative 使用默认格式化器，以当前时间为基准。
func FormatRelative(t time.Time, lang Language) string {
	return defaultRelativeFormatter.FormatRelative(t, lang)
}

// FormatRelativeSince 使用默认格式化器，以 base 时间为基准。
func FormatRelativeSince(t, base time.Time, lang Language) string {
	return defaultRelativeFormatter.FormatRelativeSince(t, base, lang)
}

func builtinRelativeLocales() map[Language]RelativeLocale {
	return map[Language]RelativeLocale{
		LangZH: localeWithAffix("刚刚", "几秒后", "前", "后", map[RelativePeriod]string{
			PeriodMinute: "分钟",
			PeriodHour:   "小时",
			PeriodDay:    "天",
			PeriodMonth:  "个月",
			PeriodYear:   "年",
		}),
		LangEN: localeWithPlural(
			"just now", "in a few seconds",
			"ago", "in",
			map[RelativePeriod][2]string{
				PeriodMinute: {"minute", "minutes"},
				PeriodHour:   {"hour", "hours"},
				PeriodDay:    {"day", "days"},
				PeriodMonth:  {"month", "months"},
				PeriodYear:   {"year", "years"},
			},
			true,
		),
		LangJA: localeWithAffix("たった今", "数秒後", "前", "後", map[RelativePeriod]string{
			PeriodMinute: "分",
			PeriodHour:   "時間",
			PeriodDay:    "日",
			PeriodMonth:  "ヶ月",
			PeriodYear:   "年",
		}),
		LangKO: localeWithAffix("방금 전", "몇 초 후", " 전", " 후", map[RelativePeriod]string{
			PeriodMinute: "분",
			PeriodHour:   "시간",
			PeriodDay:    "일",
			PeriodMonth:  "개월",
			PeriodYear:   "년",
		}),
		LangES: localeWithPlural(
			"justo ahora", "en unos segundos",
			"hace", "en",
			map[RelativePeriod][2]string{
				PeriodMinute: {"minuto", "minutos"},
				PeriodHour:   {"hora", "horas"},
				PeriodDay:    {"día", "días"},
				PeriodMonth:  {"mes", "meses"},
				PeriodYear:   {"año", "años"},
			},
			false,
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

// localeWithPlural 构建带复数单位的相对时间文案。
// spaced=true 时 past 为 "5 minutes ago"、future 为 "in 5 minutes"；
// spaced=false 时 past/future 均为 "hace 5 minutos" / "en 5 minutos" 风格。
func localeWithPlural(justNow, inFewSec, pastWord, futureWord string, units map[RelativePeriod][2]string, spaced bool) RelativeLocale {
	return RelativeLocale{
		JustNow:      justNow,
		InFewSeconds: inFewSec,
		FormatInterval: func(count int64, period RelativePeriod, isFuture bool) string {
			unit := pluralUnit(units[period], count)
			if isFuture {
				return fmt.Sprintf("%s %d %s", futureWord, count, unit)
			}
			if spaced {
				return fmt.Sprintf("%d %s %s", count, unit, pastWord)
			}
			return fmt.Sprintf("%s %d %s", pastWord, count, unit)
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

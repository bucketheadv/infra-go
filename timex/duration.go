package timex

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DurationUnit 时长格式化粒度。
type DurationUnit int

const (
	DurationDay DurationUnit = iota
	DurationHour
	DurationMinute
	DurationSecond
	DurationMillisecond
)

// DurationLocale 单种语言的时长格式化配置。
type DurationLocale struct {
	// Zero 时长为零时的文案。
	Zero string
	// FormatUnit 格式化单个时间单位。
	FormatUnit func(count int64, unit DurationUnit) string
	// JoinParts 拼接各单位片段。
	JoinParts func(parts []string) string
}

// DurationFormatter 时长格式化器，支持动态注册语言。
type DurationFormatter struct {
	// mu 保护 locales 并发读写。
	mu sync.RWMutex
	// locales 已注册语言配置。
	locales map[Language]DurationLocale
	// fallback 未命中语言时的回退语言。
	fallback Language
}

// NewDurationFormatter 创建格式化器；fallback 为未注册语言时的回退语言。
func NewDurationFormatter(fallback Language, locales map[Language]DurationLocale) *DurationFormatter {
	if locales == nil {
		locales = make(map[Language]DurationLocale)
	}
	return &DurationFormatter{
		locales:  locales,
		fallback: fallback,
	}
}

// Register 注册或覆盖一种语言配置。
func (f *DurationFormatter) Register(lang Language, locale DurationLocale) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.locales[lang] = locale
}

// Unregister 移除一种语言配置。
func (f *DurationFormatter) Unregister(lang Language) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.locales, lang)
}

// SetFallback 设置未命中语言时的回退语言。
func (f *DurationFormatter) SetFallback(lang Language) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.fallback = lang
}

// Has 判断语言是否已注册。
func (f *DurationFormatter) Has(lang Language) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.locales[lang]
	return ok
}

// Languages 返回已注册语言列表。
func (f *DurationFormatter) Languages() []Language {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]Language, 0, len(f.locales))
	for lang := range f.locales {
		out = append(out, lang)
	}
	return out
}

// FormatDuration 将时长格式化为人类可读字符串。
func (f *DurationFormatter) FormatDuration(d time.Duration, lang Language) string {
	locale, ok := f.locale(lang)
	if !ok {
		locale, ok = f.locale(f.fallbackLang())
	}
	if !ok || locale.FormatUnit == nil {
		locale = builtinDurationLocales()[LangEN]
	}

	if d < 0 {
		d = -d
	}
	if d == 0 {
		return locale.Zero
	}

	const (
		hoursPerDay       = 24
		minutesPerHour    = 60
		secondsPerMinute  = 60
		millisPerSecond   = int64(time.Second / time.Millisecond)
		millisPerMinute   = int64(time.Minute / time.Millisecond)
		millisPerHour     = int64(time.Hour / time.Millisecond)
		millisPerDay      = hoursPerDay * millisPerHour
	)

	ms := d.Milliseconds()
	values := [5]int64{
		ms / millisPerDay,
		(ms / millisPerHour) % hoursPerDay,
		(ms / millisPerMinute) % minutesPerHour,
		(ms / millisPerSecond) % secondsPerMinute,
		ms % millisPerSecond,
	}
	units := [5]DurationUnit{
		DurationDay, DurationHour, DurationMinute, DurationSecond, DurationMillisecond,
	}

	parts := make([]string, 0, 5)
	for i, count := range values {
		if count == 0 {
			continue
		}
		// 毫秒仅在更高粒度全为 0 时输出。
		if units[i] == DurationMillisecond && len(parts) > 0 {
			continue
		}
		parts = append(parts, locale.FormatUnit(count, units[i]))
	}
	if len(parts) == 0 {
		return locale.Zero
	}
	if locale.JoinParts != nil {
		return locale.JoinParts(parts)
	}
	return strings.Join(parts, " ")
}

func (f *DurationFormatter) locale(lang Language) (DurationLocale, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	locale, ok := f.locales[lang]
	return locale, ok
}

func (f *DurationFormatter) fallbackLang() Language {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.fallback
}

var defaultDurationFormatter = NewDurationFormatter(LangEN, builtinDurationLocales())

// DefaultDurationFormatter 返回默认时长格式化器。
func DefaultDurationFormatter() *DurationFormatter {
	return defaultDurationFormatter
}

// RegisterDurationLocale 向默认格式化器注册语言。
func RegisterDurationLocale(lang Language, locale DurationLocale) {
	defaultDurationFormatter.Register(lang, locale)
}

// UnregisterDurationLocale 从默认格式化器移除语言。
func UnregisterDurationLocale(lang Language) {
	defaultDurationFormatter.Unregister(lang)
}

// FormatDuration 使用默认格式化器格式化时长。
func FormatDuration(d time.Duration, lang Language) string {
	return defaultDurationFormatter.FormatDuration(d, lang)
}

func builtinDurationLocales() map[Language]DurationLocale {
	return map[Language]DurationLocale{
		LangZH: durationLocaleCompact("0秒", map[DurationUnit]string{
			DurationDay:          "天",
			DurationHour:         "小时",
			DurationMinute:       "分钟",
			DurationSecond:       "秒",
			DurationMillisecond:  "毫秒",
		}),
		LangJA: durationLocaleCompact("0秒", map[DurationUnit]string{
			DurationDay:          "日",
			DurationHour:         "時間",
			DurationMinute:       "分",
			DurationSecond:       "秒",
			DurationMillisecond:  "ミリ秒",
		}),
		LangKO: durationLocaleCompact("0초", map[DurationUnit]string{
			DurationDay:          "일",
			DurationHour:         "시간",
			DurationMinute:       "분",
			DurationSecond:       "초",
			DurationMillisecond:  "밀리초",
		}),
		LangEN: durationLocalePlural("0 seconds", map[DurationUnit][2]string{
			DurationDay:         {"day", "days"},
			DurationHour:        {"hour", "hours"},
			DurationMinute:      {"minute", "minutes"},
			DurationSecond:      {"second", "seconds"},
			DurationMillisecond: {"millisecond", "milliseconds"},
		}),
		LangES: durationLocalePlural("0 segundos", map[DurationUnit][2]string{
			DurationDay:         {"día", "días"},
			DurationHour:        {"hora", "horas"},
			DurationMinute:      {"minuto", "minutos"},
			DurationSecond:      {"segundo", "segundos"},
			DurationMillisecond: {"milisegundo", "milisegundos"},
		}),
		LangDE: durationLocalePlural("0 Sekunden", map[DurationUnit][2]string{
			DurationDay:         {"Tag", "Tagen"},
			DurationHour:        {"Stunde", "Stunden"},
			DurationMinute:      {"Minute", "Minuten"},
			DurationSecond:      {"Sekunde", "Sekunden"},
			DurationMillisecond: {"Millisekunde", "Millisekunden"},
		}),
	}
}

// durationLocaleCompact 构建紧凑书写语言（中日韩）："2小时30分钟"。
func durationLocaleCompact(zero string, units map[DurationUnit]string) DurationLocale {
	return DurationLocale{
		Zero: zero,
		FormatUnit: func(count int64, unit DurationUnit) string {
			return fmt.Sprintf("%d%s", count, units[unit])
		},
		JoinParts: func(parts []string) string {
			return strings.Join(parts, "")
		},
	}
}

// durationLocalePlural 构建带复数单位的语言："2 hours 30 minutes"。
func durationLocalePlural(zero string, units map[DurationUnit][2]string) DurationLocale {
	return DurationLocale{
		Zero: zero,
		FormatUnit: func(count int64, unit DurationUnit) string {
			return fmt.Sprintf("%d %s", count, pluralUnit(units[unit], count))
		},
		JoinParts: func(parts []string) string {
			return strings.Join(parts, " ")
		},
	}
}

package timefmt

import (
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
	JustNow       string
	InFewSeconds  string
	FormatInterval func(count int64, period RelativePeriod, isFuture bool) string
}

// RelativeFormatter 相对时间格式化器，支持动态注册语言。
type RelativeFormatter struct {
	mu       sync.RWMutex
	locales  map[Language]RelativeLocale
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
func (f *RelativeFormatter) FormatRelativeSince(t, base time.Time, lang Language) string {
	locale, ok := f.locale(lang)
	if !ok {
		locale, _ = f.locale(f.fallbackLang())
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

var defaultRelativeFormatter = NewRelativeFormatter(LangZH, builtinRelativeLocales())

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

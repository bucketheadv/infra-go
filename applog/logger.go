package applog

import (
	"context"
	"fmt"
	"time"
)

// 常用 logger 名称，与配置中 loggers 键对应。
const (
	NameRoot   = "root"
	NameApp    = "app"
	NameAccess = "access"
)

// Logger 命名日志器，类似 logback 中按 name 区分的 logger。
type Logger struct {
	name      string
	level     Level
	appenders []Appender
	reg       *Registry
}

func (l *Logger) enabled(lv Level) bool {
	return lv.enabled(l.level)
}

func (l *Logger) log(ctx context.Context, lv Level, format string, args ...any) {
	if l == nil || !l.enabled(lv) {
		return
	}
	file, line := getCallerBeyondApplog()
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	rec := &Record{
		Time:   time.Now(),
		Level:  lv,
		Logger: l.name,
		Msg:    msg,
		File:   file,
		Line:   line,
	}
	for _, a := range l.appenders {
		a.Append(rec)
	}
}

// LogFrom 使用调用方给出的源码 file、line（如 GORM 提供的触发 SQL 的位置），其余与常规日志相同（pattern/appenders）。
func LogFrom(ctx context.Context, loggerName string, lv Level, file string, line int, msg string) {
	_ = ctx
	l := Get(loggerName)
	if l == nil || !l.enabled(lv) {
		return
	}
	f := file
	if f != "" && f != "?" {
		f = shortenCallerPath(f)
	}
	rec := &Record{
		Time:   time.Now(),
		Level:  lv,
		Logger: l.name,
		Msg:    msg,
		File:   f,
		Line:   line,
	}
	for _, a := range l.appenders {
		a.Append(rec)
	}
}

// Tracef 输出 TRACE。
func (l *Logger) Tracef(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelTrace, format, args...)
}

// Debugf 输出 DEBUG。
func (l *Logger) Debugf(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelDebug, format, args...)
}

// Infof 输出 INFO。
func (l *Logger) Infof(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelInfo, format, args...)
}

// Warnf 输出 WARN。
func (l *Logger) Warnf(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelWarn, format, args...)
}

// Errorf 输出 ERROR。
func (l *Logger) Errorf(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelError, format, args...)
}

// Fatalf 输出 FATAL。
func (l *Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelFatal, format, args...)
}

// --- 包级便捷方法：使用命名 logger ---

// C 返回指定名称的 Logger（未配置则回退 root）。
func C(name string) *Logger {
	return Get(name)
}

// Tracef 使用命名 logger。
func Tracef(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).log(ctx, LevelTrace, format, args...)
}

// Debugf 使用命名 logger。
func Debugf(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).log(ctx, LevelDebug, format, args...)
}

// Infof 使用命名 logger。
func Infof(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).log(ctx, LevelInfo, format, args...)
}

// Warnf 使用命名 logger。
func Warnf(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).log(ctx, LevelWarn, format, args...)
}

// Errorf 使用命名 logger。
func Errorf(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).log(ctx, LevelError, format, args...)
}

// Fatalf 使用命名 logger。
func Fatalf(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).log(ctx, LevelFatal, format, args...)
}

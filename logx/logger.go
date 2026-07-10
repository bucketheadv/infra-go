package logx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
	// name logger 名称。
	name string
	// level 最低输出级别。
	level Level
	// appenders 绑定的输出器列表。
	appenders []Appender
	// reg 所属注册表。
	reg *Registry
}

// fatalExit 在 Fatalf 写完日志后调用；测试可替换。
var fatalExit = os.Exit

func (l *Logger) enabled(lv Level) bool {
	return lv.enabled(l.level)
}

// live 从当前默认注册表按名称解析，确保 Load 后缓存的 *Logger 仍写到新 appender。
func (l *Logger) live() *Logger {
	if l == nil {
		return nil
	}
	return Get(l.name)
}

func (l *Logger) log(ctx context.Context, lv Level, format string, args ...any) {
	cur := l.live()
	if cur == nil || !cur.enabled(lv) {
		return
	}
	file, line := getCallerBeyondLogx()
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	appendRecord(cur.name, lv, &Record{
		Time:   time.Now(),
		Level:  lv,
		Logger: cur.name,
		Msg:    msg,
		File:   file,
		Line:   line,
		Fields: FieldsFrom(ctx),
	})
}

// LogFrom 使用调用方给出的源码 file、line（如 GORM 提供的触发 SQL 的位置），其余与常规日志相同（pattern/appenders）。
func LogFrom(ctx context.Context, loggerName string, lv Level, file string, line int, msg string) {
	l := Get(loggerName)
	if l == nil || !l.enabled(lv) {
		return
	}
	f := file
	if f != "" && f != "?" {
		if filepath.IsAbs(f) {
			f = filepath.ToSlash(filepath.Clean(f))
		} else {
			f = shortenCallerPath(f)
		}
	}
	appendRecord(loggerName, lv, &Record{
		Time:   time.Now(),
		Level:  lv,
		Logger: l.name,
		Msg:    msg,
		File:   f,
		Line:   line,
		Fields: FieldsFrom(ctx),
	})
}

// appendRecord 在稳定的 registry 代际上写入：先 Add inFlight，确认未被 Load/Close 换掉后再 Append，
// 避免写到已关闭的旧 appender。
func appendRecord(loggerName string, lv Level, rec *Record) {
	for {
		reg := defaultRegistry.Load()
		reg.inFlight.Add(1)
		if defaultRegistry.Load() != reg {
			reg.inFlight.Done()
			continue
		}
		func() {
			defer reg.inFlight.Done()
			cur := lookupLogger(reg, loggerName)
			if cur == nil || !cur.enabled(lv) {
				return
			}
			rec.Logger = cur.name
			for _, a := range cur.appenders {
				a.Append(rec)
			}
		}()
		return
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

// Fatalf 输出 FATAL 后刷盘并退出进程（exit code 1）。
func (l *Logger) Fatalf(ctx context.Context, format string, args ...any) {
	l.log(ctx, LevelFatal, format, args...)
	flushRegistry(defaultRegistry.Load())
	fatalExit(1)
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

// Fatalf 使用命名 logger，写完后退出进程。
func Fatalf(ctx context.Context, loggerName string, format string, args ...any) {
	Get(loggerName).Fatalf(ctx, format, args...)
}

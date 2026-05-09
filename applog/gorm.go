package applog

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	glog "gorm.io/gorm/logger"
)

// NameGorm 为 GORM SQL 桥接使用的 logger 名称，需在 applog.yaml 的 loggers 中配置（或继承 root）。
const NameGorm = "gorm"

// GormLoggerConfig 将 GORM 日志写入 applog（与业务日志同一套 pattern / 文件滚动）。
type GormLoggerConfig struct {
	LoggerName string
	// SlowThreshold 慢 SQL 阈值，0 则使用 200ms（与 gorm logger.Default 一致）。
	SlowThreshold time.Duration
	// IgnoreRecordNotFoundError 为 true 时 ErrRecordNotFound 不记 ERROR。
	IgnoreRecordNotFoundError bool
}

type gormApplogLogger struct {
	name                 string
	level                glog.LogLevel
	slowThreshold        time.Duration
	ignoreRecordNotFound bool
}

// NewGormLogger 返回实现 gorm/logger.Interface 的适配器，输出格式与命名 logger NameGorm 一致。
func NewGormLogger(cfg GormLoggerConfig) glog.Interface {
	if cfg.LoggerName == "" {
		cfg.LoggerName = NameGorm
	}
	if cfg.SlowThreshold == 0 {
		cfg.SlowThreshold = 200 * time.Millisecond
	}
	return &gormApplogLogger{
		name:                 cfg.LoggerName,
		level:                glog.Info,
		slowThreshold:        cfg.SlowThreshold,
		ignoreRecordNotFound: cfg.IgnoreRecordNotFoundError,
	}
}

func (l *gormApplogLogger) LogMode(level glog.LogLevel) glog.Interface {
	n := *l
	n.level = level
	return &n
}

func (l *gormApplogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level < glog.Info {
		return
	}
	f, ln := gormBusinessCaller()
	body := msg
	if len(data) > 0 {
		body = fmt.Sprintf(msg, data...)
	}
	LogFrom(ctx, l.name, LevelInfo, f, ln, body)
}

func (l *gormApplogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level < glog.Warn {
		return
	}
	f, ln := gormBusinessCaller()
	body := msg
	if len(data) > 0 {
		body = fmt.Sprintf(msg, data...)
	}
	LogFrom(ctx, l.name, LevelWarn, f, ln, body)
}

func (l *gormApplogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level < glog.Error {
		return
	}
	f, ln := gormBusinessCaller()
	body := msg
	if len(data) > 0 {
		body = fmt.Sprintf(msg, data...)
	}
	LogFrom(ctx, l.name, LevelError, f, ln, body)
}

func (l *gormApplogLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= glog.Silent {
		return
	}
	sqlStr, rows := fc()
	f, ln := gormBusinessCaller()
	elapsed := time.Since(begin)
	elapsedMs := float64(elapsed.Nanoseconds()) / 1e6
	rowsLabel := formatGormRows(rows)

	switch {
	case err != nil && l.level >= glog.Error &&
		(!errors.Is(err, glog.ErrRecordNotFound) || !l.ignoreRecordNotFound):
		msg := fmt.Sprintf("[%.3fms] [rows:%s] %s | %v", elapsedMs, rowsLabel, sqlStr, err)
		LogFrom(ctx, l.name, LevelError, f, ln, msg)
	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.level >= glog.Warn:
		msg := fmt.Sprintf("SLOW SQL >= %v [%.3fms] [rows:%s] %s", l.slowThreshold, elapsedMs, rowsLabel, sqlStr)
		LogFrom(ctx, l.name, LevelWarn, f, ln, msg)
	case l.level == glog.Info:
		msg := fmt.Sprintf("[%.3fms] [rows:%s] %s", elapsedMs, rowsLabel, sqlStr)
		LogFrom(ctx, l.name, LevelInfo, f, ln, msg)
	}
}

func formatGormRows(rows int64) string {
	if rows == -1 {
		return "-"
	}
	return strconv.FormatInt(rows, 10)
}

// gormBusinessCaller 跳过本库 applog 与 gorm.io 栈帧，取业务里发起 DB 调用的 file:line（如 repository）。
func gormBusinessCaller() (file string, line int) {
	var pcs [48]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return "?", 0
	}
	frames := runtime.CallersFrames(pcs[:n])
	for {
		f, more := frames.Next()
		if f.PC == 0 {
			return "?", 0
		}
		if skipGormOrAdapterFrame(f) {
			if !more {
				return "?", 0
			}
			continue
		}
		return shortenCallerPath(f.File), f.Line
	}
}

func skipGormOrAdapterFrame(f runtime.Frame) bool {
	if f.Function != "" && strings.HasPrefix(f.Function, modulePrefix) {
		return true
	}
	slash := filepath.ToSlash(f.File)
	if strings.Contains(slash, "/infra-go/applog/") {
		return true
	}
	low := strings.ToLower(slash)
	if strings.Contains(low, "gorm.io/gorm") {
		return true
	}
	if strings.Contains(low, "gorm.io/driver") {
		return true
	}
	if strings.Contains(low, "gorm.io/plugin") {
		return true
	}
	return false
}

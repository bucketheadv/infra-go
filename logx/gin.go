package logx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bucketheadv/infra-go/timex"
	"github.com/gin-gonic/gin"
)

// NameGinWriter 用于 InstallGinWriters 写入的 logger 名（若 YAML 未配置则回退 root）。
const NameGinWriter = "gin"

// GinLoggerConfig 配置与 gin.LoggerWithConfig 对齐的访问日志（输出走 logx，不再写 os.Stdout）。
type GinLoggerConfig struct {
	// LoggerName 命名 logger，默认 NameAccess（建议在 YAML 配置 loggers.access）。
	LoggerName string
	// SkipPaths 跳过记录的路径列表。
	SkipPaths []string
	// Skip 自定义跳过判断；返回 true 时不记访问日志。
	Skip func(c *gin.Context) bool
	// SkipQueryString 为 true 时日志里的 path 不含 query（与 gin 默认 false 一致时带 query）。
	SkipQueryString bool
}

// GinRecoveryConfig 配置 panic 恢复；行为对齐 gin.Recovery，日志走 logx。
type GinRecoveryConfig struct {
	// LoggerName 默认 NameApp。
	LoggerName string
}

// GinWritersConfig 用于 InstallGinWriters，将 gin 包内直接写入 DefaultWriter / DefaultErrorWriter 的内容导入 logx。
type GinWritersConfig struct {
	// OutLoggerName 标准输出对应的 logger 名，默认 NameGinWriter。
	OutLoggerName string
	// ErrLoggerName 错误输出对应的 logger 名，默认 NameGinWriter。
	ErrLoggerName string
}

// GinLogger 返回等价于 gin.Logger() 的中间件，但每条访问日志经 logx 输出（无 ANSI；级别按 HTTP 状态码：5xx error、4xx warn、其它 info）。
func GinLogger(cfg GinLoggerConfig) gin.HandlerFunc {
	name := cfg.LoggerName
	if name == "" {
		name = NameAccess
	}
	formatter := func(param gin.LogFormatterParams) string {
		msg := formatGinAccessLine(param)
		logGinAccessLine(param.Request.Context(), name, param.StatusCode, "[GIN] "+msg)
		return ""
	}
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter:       formatter,
		Output:          io.Discard,
		SkipPaths:       cfg.SkipPaths,
		Skip:            cfg.Skip,
		SkipQueryString: cfg.SkipQueryString,
	})
}

// GinRecovery 返回等价于 gin.Recovery() 的中间件；panic 与断连场景写入 logx（含 Authorization 脱敏的请求摘要；Debug 模式下附加请求 dump）。
func GinRecovery(cfg GinRecoveryConfig) gin.HandlerFunc {
	name := cfg.LoggerName
	if name == "" {
		name = NameApp
	}
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				err, ok := rec.(error)
				isBrokenPipe := ok && (errors.Is(err, syscall.EPIPE) ||
					errors.Is(err, syscall.ECONNRESET) ||
					errors.Is(err, http.ErrAbortHandler))
				ctx := c.Request.Context()
				if isBrokenPipe {
					Warnf(ctx, name, "[GIN] recovery broken pipe: %v\n%s", rec, ginSecureRequestDump(c.Request))
					_ = c.Error(err)
					c.Abort()
					return
				}
				var msg string
				if gin.Mode() == gin.DebugMode {
					msg = fmt.Sprintf("[GIN] %s panic recovered:\n%s\n%v\n%s",
						time.Now().Format(timex.DateTimeSlashCommon),
						ginSecureRequestDump(c.Request),
						rec,
						string(debug.Stack()))
				} else {
					msg = fmt.Sprintf("[GIN] %s panic recovered:\n%v\n%s",
						time.Now().Format(timex.DateTimeSlashCommon),
						rec,
						string(debug.Stack()))
				}
				Errorf(ctx, name, "%s", msg)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// InstallGinWriters 将 gin.DefaultWriter、gin.DefaultErrorWriter 接到 logx（按行、去掉 ANSI）。
// 用于兼容 gin 在 Engine 之外的输出（如 debugPrint 等）。须在 MustLoad 之后调用。
func InstallGinWriters(cfg GinWritersConfig) {
	on := cfg.OutLoggerName
	if on == "" {
		on = NameGinWriter
	}
	en := cfg.ErrLoggerName
	if en == "" {
		en = NameGinWriter
	}
	gin.DefaultWriter = &ginLineWriter{name: on, level: LevelInfo}
	gin.DefaultErrorWriter = &ginLineWriter{name: en, level: LevelError}
}

func ginAccessLevel(code int) Level {
	switch {
	case code >= http.StatusInternalServerError:
		return LevelError
	case code >= http.StatusBadRequest:
		return LevelWarn
	default:
		return LevelInfo
	}
}

// logGinAccessLine 写入访问日志。请求结束后 formatter 的栈上多为 net/http、gin 等，无法稳定对应业务 handler；
// file:line 固定为本包内发起写入的位置，路径使用 runtime.Caller 结果的绝对路径，便于终端/IDE 按文件路径跳转。
func logGinAccessLine(ctx context.Context, loggerName string, statusCode int, msg string) {
	lv := ginAccessLevel(statusCode)
	_, df, dl, ok := runtime.Caller(1)
	if !ok {
		LogFrom(ctx, loggerName, lv, "?", 0, msg)
		return
	}
	LogFrom(ctx, loggerName, lv, absSourcePathForLog(df), dl, msg)
}

func formatGinAccessLine(param gin.LogFormatterParams) string {
	latency := param.Latency
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}
	errMsg := strings.TrimSpace(param.ErrorMessage)
	path := param.Path
	if errMsg != "" {
		errMsg = " | " + errMsg
	}
	return fmt.Sprintf("%3d | %13v | %15s | %-7s %s%s",
		param.StatusCode,
		latency,
		param.ClientIP,
		param.Method,
		path,
		errMsg,
	)
}

func ginSecureRequestDump(r *http.Request) string {
	httpRequest, err := httputil.DumpRequest(r, false)
	if err != nil {
		reportAppendError("gin dump request", err)
		return ""
	}
	lines := strings.Split(string(httpRequest), "\r\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "Authorization:") {
			lines[i] = "Authorization: *"
		}
	}
	return strings.Join(lines, "\r\n")
}

type ginLineWriter struct {
	// name 写入 logx 时使用的 logger 名称。
	name string
	// level 写入时使用的日志级别。
	level Level
	// buf 按行缓冲未完成的内容。
	buf bytes.Buffer
	// mu 保护缓冲并发写入。
	mu sync.Mutex
}

const ginLineWriterMaxBuf = 64 << 10 // 64KiB，防止无换行时缓冲无限增长

func (w *ginLineWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, _ = w.buf.Write(p)
	for {
		b := w.buf.Bytes()
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			if w.buf.Len() > ginLineWriterMaxBuf {
				line := strings.TrimRight(w.buf.String(), "\r")
				w.buf.Reset()
				line = stripANSI(line)
				if line != "" {
					reportAppendError("gin writer", fmt.Errorf("line exceeded %d bytes without newline; flushing", ginLineWriterMaxBuf))
					logGinWriterLine(w.name, w.level, line)
				}
			}
			break
		}
		line := strings.TrimRight(string(b[:i]), "\r")
		w.buf.Next(i + 1)
		line = stripANSI(line)
		if line != "" {
			logGinWriterLine(w.name, w.level, line)
		}
	}
	return len(p), nil
}

func stripANSI(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// logGinWriterLine 将 gin DefaultWriter 的一行写入 logx；file 为绝对路径以便跳转。
func logGinWriterLine(loggerName string, lv Level, msg string) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file, line = "?", 0
	}
	LogFrom(context.Background(), loggerName, lv, absSourcePathForLog(file), line, msg)
}

// absSourcePathForLog 将 runtime.Caller 给出的路径规范为绝对路径（供 Gin 桥接日志在 IDE/终端中可点击跳转）。
func absSourcePathForLog(file string) string {
	if file == "" || file == "?" {
		return file
	}
	p, err := filepath.Abs(file)
	if err != nil {
		return filepath.ToSlash(filepath.Clean(file))
	}
	return filepath.ToSlash(filepath.Clean(p))
}

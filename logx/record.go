package logx

import "time"

// Record 单条日志记录。
type Record struct {
	// Time 日志时间。
	Time time.Time
	// Level 日志级别。
	Level Level
	// Logger 产生该记录的 logger 名称。
	Logger string
	// Msg 日志正文。
	Msg string
	// File 调用方源文件路径。
	File string
	// Line 调用方源文件行号。
	Line int
	// Fields 附加键值字段。
	Fields map[string]string
}

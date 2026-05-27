package logx

import "time"

// Record 单条日志记录。
type Record struct {
	Time   time.Time
	Level  Level
	Logger string
	Msg    string
	File   string
	Line   int
}

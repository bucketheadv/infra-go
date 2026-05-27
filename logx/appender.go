package logx

// Appender 输出端（控制台、滚动文件等）。
type Appender interface {
	Append(r *Record)
	Close() error
}

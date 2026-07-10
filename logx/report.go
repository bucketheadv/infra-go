package logx

import (
	"fmt"
	"os"
)

// reportAppendError 将 appender 写入失败输出到 stderr，避免静默丢日志。
func reportAppendError(op string, err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintf(os.Stderr, "logx: %s: %v\n", op, err)
}

// ensureTrailingNewline 保证每条日志记录对应至少一个物理换行，便于按行滚动计数。
func ensureTrailingNewline(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return append(data, '\n')
	}
	return data
}

package logx

import (
	"io"
	"os"
	"sync"
)

type consoleAppender struct {
	// mu 保护并发写入。
	mu sync.Mutex
	// out 输出目标（通常为 stdout）。
	out io.Writer
	// layout 布局类型（text / json）。
	layout string
	// pattern 文本布局 pattern。
	pattern string
	// colored 是否启用 ANSI 着色。
	colored bool
	// levelColors 级别名到颜色的映射。
	levelColors map[Level]string
	// fieldColors 字段名到颜色的映射。
	fieldColors map[string]string
	// callerFileMax 调用方文件路径截断最大长度。
	callerFileMax int
}

func newConsoleAppender(
	layout string,
	pattern string,
	colored bool,
	levelColors map[Level]string,
	fieldColors map[string]string,
	callerMax int,
) *consoleAppender {
	if layout == "" {
		layout = "text"
	}
	return &consoleAppender{
		out:           os.Stdout,
		layout:        layout,
		pattern:       pattern,
		colored:       colored,
		levelColors:   levelColors,
		fieldColors:   fieldColors,
		callerFileMax: callerMax,
	}
}

func (c *consoleAppender) Append(r *Record) {
	var data []byte
	var err error
	switch c.layout {
	case "json":
		data, err = formatJSONLine(r, c.callerFileMax)
		if err != nil {
			reportAppendError("json format", err)
			data = formatTextLine(r, c.callerFileMax, false, nil)
		}
		data = append(data, '\n')
	case "pattern":
		data = ensureTrailingNewline(formatPatternLine(r, c.callerFileMax, c.colored, c.fieldColors, c.levelColors, c.pattern))
	default:
		data = formatTextLine(r, c.callerFileMax, c.colored, c.levelColors)
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.out.Write(data); err != nil {
		reportAppendError("console write", err)
	}
}

func (c *consoleAppender) Close() error {
	return nil
}

package applog

import (
	"io"
	"os"
)

type consoleAppender struct {
	out           io.Writer
	layout        string
	pattern       string
	colored       bool
	levelColors   map[Level]string
	fieldColors   map[string]string
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
			data = formatTextLine(r, c.callerFileMax, false, nil)
		}
		data = append(data, '\n')
	case "pattern":
		data = formatPatternLine(r, c.callerFileMax, c.colored, c.fieldColors, c.levelColors, c.pattern)
	default:
		data = formatTextLine(r, c.callerFileMax, c.colored, c.levelColors)
	}
	_, _ = c.out.Write(data)
}

func (c *consoleAppender) Close() error {
	return nil
}

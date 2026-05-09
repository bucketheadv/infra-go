package applog

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

func formatFileLine(file string, line int, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 48
	}
	fl := file + ":" + strconv.Itoa(line)
	if n := len(fl); n > maxLen {
		fl = "..." + fl[n-maxLen+3:]
	}
	return fl
}

func levelColorCode(level Level, colored bool, overrides map[Level]string) string {
	if !colored {
		return ""
	}
	if overrides != nil {
		if c, ok := overrides[level]; ok && c != "" {
			return normalizeColorCode(c)
		}
	}
	switch level {
	case LevelTrace:
		return "\x1b[90m"
	case LevelDebug:
		return "\x1b[36m"
	case LevelInfo:
		return "\x1b[32m"
	case LevelWarn:
		return "\x1b[33m"
	case LevelError:
		return "\x1b[31m"
	case LevelFatal:
		return "\x1b[35m"
	default:
		return "\x1b[0m"
	}
}

const ansiReset = "\x1b[0m"

var namedSGR = map[string]string{
	"black":   "30",
	"red":     "31",
	"green":   "32",
	"yellow":  "33",
	"blue":    "34",
	"magenta": "35",
	"cyan":    "36",
	"white":   "37",
	"gray":    "90",
	"grey":    "90",
}

// normalizeColorCode 支持颜色名（green）、纯数字 SGR（32）、样式 faint/bold/dim 或完整序列（\x1b[32m）。
func normalizeColorCode(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if strings.HasPrefix(s, "\x1b[") {
		return s
	}
	switch strings.ToLower(s) {
	case "faint", "dim":
		return "\x1b[2m"
	case "bold":
		return "\x1b[1m"
	}
	if len(s) <= 3 && isDigits(s) {
		return "\x1b[" + s + "m"
	}
	if code, ok := namedSGR[strings.ToLower(s)]; ok {
		return "\x1b[" + code + "m"
	}
	return "\x1b[" + s + "m"
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

func parseLevelColors(m map[string]string) map[Level]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[Level]string)
	for k, v := range m {
		lv := parseLevel(k)
		out[lv] = v
	}
	return out
}

func formatTextLine(r *Record, callerMax int, colored bool, colors map[Level]string) []byte {
	prefix := levelColorCode(r.Level, colored, colors)
	suffix := ""
	if colored && prefix != "" {
		suffix = ansiReset
	}
	fl := formatFileLine(r.File, r.Line, callerMax)
	pid := strconv.Itoa(os.Getpid())
	line := "[" + r.Level.String() + "][" +
		r.Time.Format("2006-01-02T15:04:05.000") + "][" + pid + "][" + fl + "][" + r.Logger + "] " +
		r.Msg + "\n"
	if prefix == "" {
		return []byte(line)
	}
	var b strings.Builder
	b.Grow(len(prefix) + len(line) + len(suffix))
	b.WriteString(prefix)
	b.WriteString(strings.TrimSuffix(line, "\n"))
	b.WriteString(suffix)
	b.WriteByte('\n')
	return []byte(b.String())
}

func formatJSONLine(r *Record, callerMax int) ([]byte, error) {
	fl := formatFileLine(r.File, r.Line, callerMax)
	m := map[string]any{
		"level":    strings.ToLower(r.Level.String()),
		"time":     r.Time.Format("2006-01-02T15:04:05.000"),
		"pid":      os.Getpid(),
		"fileLine": fl,
		"logger":   r.Logger,
		"msg":      r.Msg,
	}
	return json.Marshal(m)
}

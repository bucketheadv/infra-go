package logx

import (
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/bucketheadv/infra-go/timex"
)

// defaultLogPattern 与原先 text 布局信息项接近；时间格式为 Go 参考时间布局。
const defaultLogPattern = "[%level][%d][%pid][%fileLine] [%logger] %msg%n"

// formatPatternLine 按 pattern 输出一行。
// colored 为 true 时：普通占位符使用 fieldColors / levelColors；%clr(子模式){颜色} 使用括号后的颜色（类似 Spring %clr)。
func formatPatternLine(
	r *Record,
	callerMax int,
	colored bool,
	fieldColors map[string]string,
	levelColors map[Level]string,
	pattern string,
) []byte {
	if pattern == "" {
		pattern = defaultLogPattern
	}
	return formatPatternImpl(r, callerMax, colored, fieldColors, levelColors, pattern, true)
}

func formatPatternImpl(
	r *Record,
	callerMax int,
	colored bool,
	fieldColors map[string]string,
	levelColors map[Level]string,
	pattern string,
	applyFieldColor bool,
) []byte {
	var b strings.Builder
	i := 0
	for i < len(pattern) {
		if pattern[i] != '%' {
			start := i
			for i < len(pattern) && pattern[i] != '%' {
				i++
			}
			b.WriteString(pattern[start:i])
			continue
		}
		if i+1 < len(pattern) && pattern[i+1] == '%' {
			b.WriteByte('%')
			i += 2
			continue
		}
		if strings.HasPrefix(pattern[i+1:], "clr(") {
			inner, cspec, clen, ok := parseClrBlock(pattern[i+1:])
			if !ok {
				b.WriteByte('%')
				i++
				continue
			}
			sub := formatPatternImpl(r, callerMax, colored, fieldColors, levelColors, inner, false)
			b.WriteString(wrapWithColor(string(sub), cspec, colored))
			i += 1 + clen
			continue
		}
		pct := i
		i++
		if i >= len(pattern) {
			b.WriteByte('%')
			break
		}
		dStart := i
		for i < len(pattern) && pattern[i] >= '0' && pattern[i] <= '9' {
			i++
		}
		if i >= len(pattern) && dStart < i {
			b.WriteString(pattern[pct:])
			break
		}
		tokStart := i
		for i < len(pattern) && isPatternIdentByte(pattern[i]) {
			i++
		}
		token := pattern[tokStart:i]
		var layoutMod string
		if i < len(pattern) && pattern[i] == '{' {
			end := strings.IndexByte(pattern[i:], '}')
			if end < 0 {
				b.WriteString(pattern[tokStart-1:])
				break
			}
			layoutMod = pattern[i+1 : i+end]
			i += end + 1
		}
		fieldKey, raw := resolvePatternToken(r, callerMax, token, layoutMod)
		if applyFieldColor {
			b.WriteString(wrapFieldColor(raw, fieldKey, colored, r.Level, fieldColors, levelColors))
		} else {
			b.WriteString(raw)
		}
	}
	return []byte(b.String())
}

func parseClrBlock(s string) (inner string, color string, length int, ok bool) {
	if !strings.HasPrefix(s, "clr(") {
		return "", "", 0, false
	}
	depth := 0
	j := 4
	start := j
	for j < len(s) {
		switch s[j] {
		case '(':
			depth++
		case ')':
			if depth == 0 {
				inner = s[start:j]
				j++
				goto afterParen
			}
			depth--
		}
		j++
	}
	return "", "", 0, false
afterParen:
	if j >= len(s) || s[j] != '{' {
		return "", "", 0, false
	}
	end := strings.IndexByte(s[j:], '}')
	if end < 0 {
		return "", "", 0, false
	}
	color = s[j+1 : j+end]
	length = j + end + 1
	return inner, color, length, true
}

func wrapWithColor(val string, colorSpec string, colored bool) string {
	if !colored {
		return val
	}
	code := normalizeColorCode(strings.TrimSpace(colorSpec))
	if code == "" {
		return val
	}
	return code + val + ansiReset
}

func isPatternIdentByte(c byte) bool {
	return unicode.IsLetter(rune(c)) || c == '_'
}

func resolvePatternToken(r *Record, callerMax int, token, layoutMod string) (fieldKey, value string) {
	switch strings.ToLower(token) {
	case "d", "date":
		layout := layoutMod
		if layout == "" {
			layout = timex.DateTimeMillisISO
		}
		return "date", r.Time.Format(layout)
	case "level", "p":
		return "level", strings.ToUpper(r.Level.String())
	case "fileline", "f":
		return "fileLine", formatFileLine(r.File, r.Line, callerMax)
	case "logger", "c":
		return "logger", r.Logger
	case "pid", "processid", "process_id":
		return "pid", strconv.Itoa(os.Getpid())
	case "msg", "m":
		return "msg", r.Msg
	case "n":
		return "", "\n"
	default:
		if token == "" {
			return "", "%"
		}
		lit := "%" + token
		if layoutMod != "" {
			lit += "{" + layoutMod + "}"
		}
		return "", lit
	}
}

func wrapFieldColor(
	val, fieldKey string,
	colored bool,
	lv Level,
	fieldColors map[string]string,
	levelColors map[Level]string,
) string {
	if val == "" && fieldKey == "" {
		return val
	}
	if !colored || fieldKey == "" {
		return val
	}
	var code string
	if fieldKey == "level" {
		if fieldColors != nil {
			if c := strings.TrimSpace(fieldColors["level"]); c != "" {
				code = normalizeColorCode(c)
			}
		}
		if code == "" {
			code = levelColorCode(lv, true, levelColors)
		}
	} else if fieldColors != nil {
		if c := strings.TrimSpace(fieldColors[fieldKey]); c != "" {
			code = normalizeColorCode(c)
		}
	}
	if code == "" {
		return val
	}
	return code + val + ansiReset
}

func normalizeFieldColorKeys(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string)
	for k, v := range m {
		kk := strings.ToLower(strings.TrimSpace(k))
		switch kk {
		case "date", "time":
			out["date"] = v
		case "level", "p":
			out["level"] = v
		case "fileline", "file", "caller", "location":
			out["fileLine"] = v
		case "logger", "category", "name":
			out["logger"] = v
		case "msg", "message":
			out["msg"] = v
		case "pid", "processid", "process":
			out["pid"] = v
		}
	}
	return out
}

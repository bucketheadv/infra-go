package stringx

import (
	"strings"
	"unicode"
)

type stringLike interface {
	~string | *string
}

func asString[T stringLike](v T) (string, bool) {
	switch x := any(v).(type) {
	case string:
		return x, true
	case *string:
		if x == nil {
			return "", false
		}
		return *x, true
	default:
		return "", false
	}
}

// IsEmpty 判断字符串是否为空。
// - string: 长度为 0 返回 true；
// - *string: nil 或指向空串返回 true。
func IsEmpty[T stringLike](v T) bool {
	s, ok := asString(v)
	if !ok {
		return true
	}
	return s == ""
}

// IsBlank 判断字符串是否为空白。
// - string: 空串或仅空白字符返回 true；
// - *string: nil、指向空串或仅空白字符返回 true。
func IsBlank[T stringLike](v T) bool {
	s, ok := asString(v)
	if !ok {
		return true
	}
	return strings.TrimSpace(s) == ""
}

// DefaultIfBlank 当 s 为空白时返回 defaultVal，否则返回 s 本身。
func DefaultIfBlank[T stringLike](s T, defaultVal string) string {
	if IsBlank(s) {
		return defaultVal
	}
	v, _ := asString(s)
	return v
}

// Truncate 将字符串截断到 maxLen 个 rune；maxLen <= 0 时返回空串。
func Truncate[T stringLike](s T, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	v, ok := asString(s)
	if !ok {
		return ""
	}
	runes := []rune(v)
	if len(runes) <= maxLen {
		return v
	}
	return string(runes[:maxLen])
}

// SplitTrim 按 sep 分割字符串，并去除每段首尾空白；空段会被丢弃。
func SplitTrim[T stringLike](s T, sep string) []string {
	v, ok := asString(s)
	if !ok {
		return nil
	}
	parts := strings.Split(v, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// JoinNonEmpty 拼接非空白字符串，自动跳过 blank 段。
func JoinNonEmpty(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	filtered := make([]string, 0, len(parts))
	for _, p := range parts {
		if !IsBlank(p) {
			filtered = append(filtered, p)
		}
	}
	return strings.Join(filtered, sep)
}

// SnakeCase 将字符串转为 snake_case（小写 + 下划线）。
// 例如 "HelloWorld" -> "hello_world"，"HTTPStatus" -> "http_status"。
func SnakeCase[T stringLike](s T) string {
	v, ok := asString(s)
	if !ok {
		return ""
	}
	runes := []rune(v)
	if len(runes) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(runes) + 4)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := runes[i-1]
				if !unicode.IsUpper(prev) && prev != '_' {
					b.WriteByte('_')
				} else if i+1 < len(runes) && unicode.IsUpper(prev) && unicode.IsLower(runes[i+1]) {
					b.WriteByte('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		if r == '-' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// CamelCase 将 snake_case / kebab-case / PascalCase 转为 PascalCase。
// 例如 "hello_world" -> "HelloWorld"，"HelloWorld" -> "HelloWorld"。
func CamelCase[T stringLike](s T) string {
	v, ok := asString(s)
	if !ok || v == "" {
		return ""
	}
	normalized := strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(strings.TrimSpace(v))
	parts := strings.Fields(normalized)
	if len(parts) == 1 {
		parts = splitCamelParts(parts[0])
	}
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}

// splitCamelParts 按大写字母边界拆分 PascalCase/camelCase 字符串。
func splitCamelParts(s string) []string {
	runes := []rune(s)
	if len(runes) == 0 {
		return nil
	}
	parts := make([]string, 0, 4)
	start := 0
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) && unicode.IsLower(runes[i-1]) {
			parts = append(parts, string(runes[start:i]))
			start = i
			continue
		}
		if i+1 < len(runes) && unicode.IsUpper(runes[i-1]) && unicode.IsUpper(runes[i]) && unicode.IsLower(runes[i+1]) {
			parts = append(parts, string(runes[start:i]))
			start = i
		}
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

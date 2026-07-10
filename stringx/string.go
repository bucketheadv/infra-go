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
// 注意：与 collectionx.IsEmpty 不同，本函数仅面向 string / *string。
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
		return []string{}
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

// LowerCamelCase 将 snake_case / kebab-case / PascalCase 转为 lowerCamelCase。
// 例如 "hello_world" -> "helloWorld"，"HelloWorld" -> "helloWorld"。
func LowerCamelCase[T stringLike](s T) string {
	cc := CamelCase(s)
	if cc == "" {
		return ""
	}
	runes := []rune(cc)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
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

// EqualsIgnoreCase 判断两个字符串是否相等（忽略大小写）。
func EqualsIgnoreCase[T, U stringLike](a T, b U) bool {
	sa, oka := asString(a)
	sb, okb := asString(b)
	if !oka || !okb {
		return !oka && !okb
	}
	return strings.EqualFold(sa, sb)
}

// Contains 判断 s 是否包含 substr。
func Contains[T, U stringLike](s T, substr U) bool {
	v, ok := asString(s)
	if !ok {
		return false
	}
	sub, ok := asString(substr)
	if !ok {
		return false
	}
	return strings.Contains(v, sub)
}

// HasPrefix 判断 s 是否以 prefix 开头。
func HasPrefix[T, U stringLike](s T, prefix U) bool {
	v, ok := asString(s)
	if !ok {
		return false
	}
	p, ok := asString(prefix)
	if !ok {
		return false
	}
	return strings.HasPrefix(v, p)
}

// HasSuffix 判断 s 是否以 suffix 结尾。
func HasSuffix[T, U stringLike](s T, suffix U) bool {
	v, ok := asString(s)
	if !ok {
		return false
	}
	suf, ok := asString(suffix)
	if !ok {
		return false
	}
	return strings.HasSuffix(v, suf)
}

// KebabCase 将字符串转为 kebab-case（小写 + 连字符）。
func KebabCase[T stringLike](s T) string {
	return strings.ReplaceAll(SnakeCase(s), "_", "-")
}

// Repeat 将字符串重复 count 次；count <= 0 时返回空串。
func Repeat[T stringLike](s T, count int) string {
	if count <= 0 {
		return ""
	}
	v, ok := asString(s)
	if !ok {
		return ""
	}
	return strings.Repeat(v, count)
}

// Ellipsis 将字符串截断并在末尾追加 "..."；maxLen 为总长度上限（按 rune 计，含省略号）。
func Ellipsis[T stringLike](s T, maxLen int) string {
	const suffix = "..."
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
	suffixRunes := []rune(suffix)
	if maxLen <= len(suffixRunes) {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-len(suffixRunes)]) + suffix
}

// PadLeft 在左侧填充 pad 直至长度达到 length（按 rune 计）；已足够长则原样返回。
func PadLeft[T stringLike](s T, length int, pad rune) string {
	if length <= 0 {
		return ""
	}
	v, ok := asString(s)
	if !ok {
		return ""
	}
	runes := []rune(v)
	if len(runes) >= length {
		return v
	}
	padStr := string(pad)
	var b strings.Builder
	b.Grow(length * len(padStr))
	for i := len(runes); i < length; i++ {
		b.WriteString(padStr)
	}
	b.WriteString(v)
	return b.String()
}

// PadRight 在右侧填充 pad 直至长度达到 length（按 rune 计）；已足够长则原样返回。
func PadRight[T stringLike](s T, length int, pad rune) string {
	if length <= 0 {
		return ""
	}
	v, ok := asString(s)
	if !ok {
		return ""
	}
	runes := []rune(v)
	if len(runes) >= length {
		return v
	}
	padStr := string(pad)
	var b strings.Builder
	b.Grow(length*len(padStr))
	b.WriteString(v)
	for i := len(runes); i < length; i++ {
		b.WriteString(padStr)
	}
	return b.String()
}

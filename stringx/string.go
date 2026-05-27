package stringx

import "strings"

type stringLike interface {
	~string | *string
}

// IsEmpty 判断字符串是否为空。
// - string: 长度为 0 返回 true；
// - *string: nil 或指向空串返回 true。
func IsEmpty[T stringLike](v T) bool {
	switch x := any(v).(type) {
	case string:
		return x == ""
	case *string:
		return x == nil || *x == ""
	default:
		return true
	}
}

// IsBlank 判断字符串是否为空白。
// - string: 空串或仅空白字符返回 true；
// - *string: nil、指向空串或仅空白字符返回 true。
func IsBlank[T stringLike](v T) bool {
	switch x := any(v).(type) {
	case string:
		return strings.TrimSpace(x) == ""
	case *string:
		return x == nil || strings.TrimSpace(*x) == ""
	default:
		return true
	}
}

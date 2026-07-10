package jsonx

import "strings"

// ErrorCode 表示 jsonx 错误码。
type ErrorCode string

const (
	// ErrCodeTargetNil 目标对象为 nil。
	ErrCodeTargetNil ErrorCode = "TARGET_NIL"
	// ErrCodeTargetNotPointer 目标对象不是可写的非空指针。
	ErrCodeTargetNotPointer ErrorCode = "TARGET_NOT_POINTER"
	// ErrCodeConvertNumber 数字转换失败。
	ErrCodeConvertNumber ErrorCode = "CONVERT_NUMBER"
	// ErrCodeTypeMismatch JSON 值与目标字段类型不匹配。
	ErrCodeTypeMismatch ErrorCode = "TYPE_MISMATCH"
	// ErrCodeInvalidJSON JSON 文本非法或含尾部多余内容。
	ErrCodeInvalidJSON ErrorCode = "INVALID_JSON"
)

// Error 是 jsonx 自定义错误类型。
type Error struct {
	// Code 错误码。
	Code ErrorCode
	// Message 可读错误说明。
	Message string
	// Expected 期望的目标类型或格式。
	Expected string
	// Value 触发错误的原始值摘要。
	Value string
	// Cause 底层错误（可选）。
	Cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("jsonx: ")
	if e.Code != "" {
		b.WriteString(string(e.Code))
	} else {
		b.WriteString("ERROR")
	}
	if e.Message != "" {
		b.WriteString(": ")
		b.WriteString(e.Message)
	}
	if e.Expected != "" {
		b.WriteString(" expected=")
		b.WriteString(e.Expected)
	}
	if e.Value != "" {
		b.WriteString(" value=")
		b.WriteString(e.Value)
	}
	if e.Cause != nil {
		b.WriteString(" cause=")
		b.WriteString(e.Cause.Error())
	}
	return b.String()
}

// Unwrap 返回底层错误。
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

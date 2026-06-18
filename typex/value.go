package typex

import "cmp"

// Deref 解引用指针；当 ptr 为 nil 时返回 defaultVal。
func Deref[T any](ptr *T, defaultVal T) T {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// OrZero 解引用指针；当 ptr 为 nil 时返回类型零值。
func OrZero[T any](ptr *T) T {
	var zero T
	if ptr == nil {
		return zero
	}
	return *ptr
}

// Coalesce 返回第一个非零值；若全部为零值则返回零值。
func Coalesce[T cmp.Ordered | bool](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}

// Must 在 err 非 nil 时 panic，否则返回 v。
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// FirstNonNil 返回首个非 nil 指针；全部为 nil 时返回 nil。
func FirstNonNil[T any](ptrs ...*T) *T {
	for _, p := range ptrs {
		if p != nil {
			return p
		}
	}
	return nil
}

// If 根据 cond 返回 a 或 b。
func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

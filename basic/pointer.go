package basic

// Ptr 将任意值转换为对应类型的指针。
//
//go:fix inline
func Ptr[T any](v T) *T {
	return new(v)
}

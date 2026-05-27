package basic

// Ptr 将任意值转换为对应类型的指针。
func Ptr[T any](v T) *T {
	p := new(T)
	*p = v
	return p
}

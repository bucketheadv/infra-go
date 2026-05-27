package collection

// IsSlicePtrEmpty 判断切片指针是否为空。
// 当指针为 nil，或指向切片长度为 0 时返回 true。
func IsSlicePtrEmpty[T any](arr *[]T) bool {
	return arr == nil || len(*arr) == 0
}

// IsMapPtrEmpty 判断 map 指针是否为空。
// 当指针为 nil，或指向 map 长度为 0 时返回 true。
func IsMapPtrEmpty[K comparable, V any](m *map[K]V) bool {
	return m == nil || len(*m) == 0
}

// IsSliceEmpty 判断切片是否为空（非指针）。
// 当切片为 nil，或长度为 0 时返回 true。
func IsSliceEmpty[T any](arr []T) bool {
	return len(arr) == 0
}

// IsMapEmpty 判断 map 是否为空（非指针）。
// 当 map 为 nil，或长度为 0 时返回 true。
func IsMapEmpty[K comparable, V any](m map[K]V) bool {
	return len(m) == 0
}

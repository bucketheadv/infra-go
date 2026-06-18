package collectionx

import (
	"cmp"
	"sort"
)

// Map 将切片元素逐一映射为新类型切片，保留原顺序。
func Map[T any, R any](arr []T, fn func(T) R) []R {
	if len(arr) == 0 {
		return []R{}
	}
	result := make([]R, len(arr))
	for i, v := range arr {
		result[i] = fn(v)
	}
	return result
}

// Filter 保留满足 predicate 的元素，保持原顺序。
func Filter[T any](arr []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Unique 对可比较类型去重，保留首次出现顺序。
func Unique[T comparable](arr []T) []T {
	if len(arr) == 0 {
		return []T{}
	}
	seen := make(map[T]struct{}, len(arr))
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// Contains 判断切片是否包含指定元素。
func Contains[T comparable](arr []T, v T) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}

// Find 返回首个满足 predicate 的元素；未找到时第二个返回值为 false。
func Find[T any](arr []T, predicate func(T) bool) (T, bool) {
	var zero T
	for _, v := range arr {
		if predicate(v) {
			return v, true
		}
	}
	return zero, false
}

// FindIndex 返回首个满足 predicate 的元素下标；未找到时第二个返回值为 false。
func FindIndex[T any](arr []T, predicate func(T) bool) (int, bool) {
	for i, v := range arr {
		if predicate(v) {
			return i, true
		}
	}
	return -1, false
}

// FilterMap 映射切片元素并在 fn 返回 false 时跳过；保留原顺序。
func FilterMap[T any, R any](arr []T, fn func(T) (R, bool)) []R {
	result := make([]R, 0, len(arr))
	for _, v := range arr {
		if r, ok := fn(v); ok {
			result = append(result, r)
		}
	}
	return result
}

// DistinctBy 按 keyFn 提取的键去重，保留首次出现顺序。
func DistinctBy[T any, K comparable](arr []T, keyFn func(T) K) []T {
	if len(arr) == 0 {
		return []T{}
	}
	seen := make(map[K]struct{}, len(arr))
	result := make([]T, 0, len(arr))
	for _, v := range arr {
		k := keyFn(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		result = append(result, v)
	}
	return result
}

// Union 返回 a 与 b 的并集，去重并保留 a 的顺序，再追加 b 中首次出现的元素。
func Union[T comparable](a, b []T) []T {
	if len(a) == 0 {
		return Unique(b)
	}
	if len(b) == 0 {
		return Unique(a)
	}
	seen := make(map[T]struct{}, len(a)+len(b))
	result := make([]T, 0, len(a)+len(b))
	for _, v := range a {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	for _, v := range b {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// Every 当所有元素都满足 predicate 时返回 true；空切片返回 true。
func Every[T any](arr []T, predicate func(T) bool) bool {
	for _, v := range arr {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// Some 当存在任一元素满足 predicate 时返回 true。
func Some[T any](arr []T, predicate func(T) bool) bool {
	for _, v := range arr {
		if predicate(v) {
			return true
		}
	}
	return false
}

// Reverse 原地反转切片元素顺序。
func Reverse[T any](arr []T) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

// Flatten 将二维切片展平为一维切片。
func Flatten[T any](arr [][]T) []T {
	total := 0
	for _, inner := range arr {
		total += len(inner)
	}
	result := make([]T, 0, total)
	for _, inner := range arr {
		result = append(result, inner...)
	}
	return result
}

// Partition 将切片按固定大小分组。
// 例如 [1,2,3,4,5] 且 size=2，返回 [[1,2],[3,4],[5]]。
// 当 arr 为空时返回空二维切片。
// 当 size 小于等于 0 时返回空二维切片。
func Partition[T any](arr []T, size int) [][]T {
	result := make([][]T, 0)
	length := len(arr)
	if length <= 0 || size <= 0 {
		return result
	}
	outSize := length / size
	for i := 0; i <= outSize; i++ {
		innerArr := make([]T, 0)
		for j := i * size; j < min((i+1)*size, length); j++ {
			innerArr = append(innerArr, arr[j])
		}
		if len(innerArr) > 0 {
			result = append(result, innerArr)
		}
	}
	return result
}

// GroupBy 按 function 产生的键对切片元素分组。
// 返回值中每个 key 对应同组元素，且组内顺序与原切片一致。
func GroupBy[T any, R cmp.Ordered](arr []T, function func(T) R) map[R][]T {
	result := map[R][]T{}
	for _, ele := range arr {
		k := function(ele)
		result[k] = append(result[k], ele)
	}
	return result
}

// ArrayToMap 将切片按 keyFunc 生成的键转换为 map。
// 当 coverExists=false 时保留首个同 key 元素；
// 当 coverExists=true 时后出现元素会覆盖之前的值。
func ArrayToMap[T any, R cmp.Ordered](arr []T, coverExists bool, keyFunc func(T) R) map[R]T {
	result := map[R]T{}
	for _, ele := range arr {
		k := keyFunc(ele)
		_, ok := result[k]
		if !ok || coverExists {
			result[k] = ele
		}
	}
	return result
}

// Reduce 将切片归约为单个值。
func Reduce[T any, R any](arr []T, initial R, fn func(R, T) R) R {
	acc := initial
	for _, v := range arr {
		acc = fn(acc, v)
	}
	return acc
}

// Paginate 按 page（从 1 开始）和 size 截取切片页；非法参数返回空切片。
func Paginate[T any](arr []T, page, size int) []T {
	if page < 1 || size <= 0 || len(arr) == 0 {
		return []T{}
	}
	start := (page - 1) * size
	if start >= len(arr) {
		return []T{}
	}
	end := start + size
	if end > len(arr) {
		end = len(arr)
	}
	return arr[start:end]
}

// SortBy 按 keyFn 提取的键升序排序，返回新切片，不修改原切片。
func SortBy[T any, K cmp.Ordered](arr []T, keyFn func(T) K) []T {
	result := append([]T(nil), arr...)
	sort.Slice(result, func(i, j int) bool {
		return keyFn(result[i]) < keyFn(result[j])
	})
	return result
}

// Intersect 返回 a 与 b 的交集，保留 a 中的首次出现顺序。
func Intersect[T comparable](a, b []T) []T {
	if len(a) == 0 || len(b) == 0 {
		return []T{}
	}
	set := make(map[T]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	result := make([]T, 0)
	seen := make(map[T]struct{})
	for _, v := range a {
		if _, ok := set[v]; !ok {
			continue
		}
		if _, dup := seen[v]; dup {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// Difference 返回存在于 a 但不在 b 中的元素，保留 a 的顺序。
func Difference[T comparable](a, b []T) []T {
	if len(a) == 0 {
		return []T{}
	}
	set := make(map[T]struct{}, len(b))
	for _, v := range b {
		set[v] = struct{}{}
	}
	result := make([]T, 0, len(a))
	for _, v := range a {
		if _, ok := set[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// MinOf 返回切片最小值；空切片时第二个返回值为 false。
func MinOf[T cmp.Ordered](arr []T) (T, bool) {
	var zero T
	if len(arr) == 0 {
		return zero, false
	}
	smallest := arr[0]
	for _, v := range arr[1:] {
		if v < smallest {
			smallest = v
		}
	}
	return smallest, true
}

// MaxOf 返回切片最大值；空切片时第二个返回值为 false。
func MaxOf[T cmp.Ordered](arr []T) (T, bool) {
	var zero T
	if len(arr) == 0 {
		return zero, false
	}
	largest := arr[0]
	for _, v := range arr[1:] {
		if v > largest {
			largest = v
		}
	}
	return largest, true
}

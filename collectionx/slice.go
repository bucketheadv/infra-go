package collectionx

import "cmp"

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

package collection

import "cmp"

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

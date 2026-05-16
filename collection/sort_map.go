package collection

import (
	"cmp"
	"sort"
)

// SortedMapTraversal 按 key 排序后遍历 map。
// reverse=false 时按升序遍历，reverse=true 时按降序遍历。
// function 会按排序后的顺序被调用。
func SortedMapTraversal[T cmp.Ordered, R any](m map[T]R, reverse bool, function func(T, R)) {
	keys := make([]T, 0)
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if reverse {
			return keys[i] > keys[j]
		}
		return keys[i] < keys[j]
	})

	for _, v := range keys {
		function(v, m[v])
	}
}

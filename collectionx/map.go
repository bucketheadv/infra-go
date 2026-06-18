package collectionx

import (
	"cmp"
	"sort"
)

// Keys 返回 map 的全部键；遍历顺序不保证稳定。
func Keys[K comparable, V any](m map[K]V) []K {
	if len(m) == 0 {
		return []K{}
	}
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 返回 map 的全部值；遍历顺序不保证稳定。
func Values[K comparable, V any](m map[K]V) []V {
	if len(m) == 0 {
		return []V{}
	}
	vals := make([]V, 0, len(m))
	for _, v := range m {
		vals = append(vals, v)
	}
	return vals
}

// MergeMaps 合并多个 map，后出现 map 的同 key 会覆盖先前的值。
func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	total := 0
	for _, m := range maps {
		total += len(m)
	}
	result := make(map[K]V, total)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// SortedMapTraversal 按 key 排序后遍历 map。
// reverse=false 时按升序遍历，reverse=true 时按降序遍历。
// function 会按排序后的顺序被调用。
func SortedMapTraversal[T cmp.Ordered, R any](m map[T]R, reverse bool, function func(T, R)) {
	keys := make([]T, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		if reverse {
			return keys[i] > keys[j]
		}
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		function(k, m[k])
	}
}

// GetOrDefault 获取 map 中的值；key 不存在时返回 defaultVal。
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultVal V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultVal
}

// Pick 从 map 中挑选指定 key，不存在的 key 会被忽略。
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit 返回排除指定 key 后的 map 副本。
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	skip := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		skip[k] = struct{}{}
	}
	result := make(map[K]V, len(m))
	for k, v := range m {
		if _, excluded := skip[k]; excluded {
			continue
		}
		result[k] = v
	}
	return result
}

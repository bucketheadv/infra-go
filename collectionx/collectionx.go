package collectionx

import "reflect"

// IsEmpty 判断 slice 或 map 是否为空；支持值与指针。
// nil、长度为 0，或 nil 指针均返回 true；非 slice/map 类型返回 false。
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		return rv.Len() == 0
	case reflect.Pointer:
		if rv.IsNil() {
			return true
		}
		elem := rv.Elem()
		switch elem.Kind() {
		case reflect.Slice, reflect.Map:
			return elem.Len() == 0
		default:
			return false
		}
	default:
		return false
	}
}

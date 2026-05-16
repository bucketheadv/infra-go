package basic

import (
	"cmp"
	"fmt"
	"strconv"
)

// StringTo 将字符串解析为目标基础类型。
// 支持 string、整型、浮点型和 bool。
func StringTo[T cmp.Ordered | bool](v string) (T, error) {
	var zero T
	switch any(zero).(type) {
	case string:
		return any(v).(T), nil
	case int:
		i, err := strconv.ParseInt(v, 10, strconv.IntSize)
		if err != nil {
			return zero, err
		}
		return any(int(i)).(T), nil
	case int8:
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return zero, err
		}
		return any(int8(i)).(T), nil
	case int16:
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return zero, err
		}
		return any(int16(i)).(T), nil
	case int32:
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return zero, err
		}
		return any(int32(i)).(T), nil
	case int64:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return zero, err
		}
		return any(i).(T), nil
	case uint:
		u, err := strconv.ParseUint(v, 10, strconv.IntSize)
		if err != nil {
			return zero, err
		}
		return any(uint(u)).(T), nil
	case uint8:
		u, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return zero, err
		}
		return any(uint8(u)).(T), nil
	case uint16:
		u, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return zero, err
		}
		return any(uint16(u)).(T), nil
	case uint32:
		u, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return zero, err
		}
		return any(uint32(u)).(T), nil
	case uint64:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return zero, err
		}
		return any(u).(T), nil
	case uintptr:
		u, err := strconv.ParseUint(v, 10, strconv.IntSize)
		if err != nil {
			return zero, err
		}
		return any(uintptr(u)).(T), nil
	case float32:
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return zero, err
		}
		return any(float32(f)).(T), nil
	case float64:
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return zero, err
		}
		return any(f).(T), nil
	case bool:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return zero, err
		}
		return any(b).(T), nil
	default:
		return zero, fmt.Errorf("不支持的数据类型: %T", zero)
	}
}

// ArrayElemTo 将字符串切片逐个转换为目标基础类型切片。
// 当任一元素转换失败时立即返回已成功转换的部分结果与错误。
func ArrayElemTo[T cmp.Ordered | bool](v []string) ([]T, error) {
	result := make([]T, 0)
	for _, v := range v {
		tmp, err := StringTo[T](v)
		if err != nil {
			return result, err
		}
		result = append(result, tmp)
	}
	return result, nil
}

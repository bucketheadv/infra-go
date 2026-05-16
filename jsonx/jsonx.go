package jsonx

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal 是增强版 JSON 反序列化。
// 当目标字段是数字类型而 JSON 值为字符串时，会自动尝试转换。
func Unmarshal(data []byte, v any) error {
	if v == nil {
		return &Error{
			Code:    ErrCodeTargetNil,
			Message: "target is nil",
		}
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &Error{
			Code:    ErrCodeTargetNotPointer,
			Message: "target must be non-nil pointer",
		}
	}

	var raw any
	// 第一步先解成通用结构（map[string]any / []any 等）。
	// 这里显式启用 UseNumber，避免标准库把所有数字直接转成 float64，
	// 从而在后续能按目标字段位宽（如 int32、uint16）做更精确的转换。
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return err
	}

	converted, err := convertByType(raw, rv.Type().Elem())
	if err != nil {
		return err
	}

	// 第二步直接反射写入目标对象。
	// 这样可避免“先 Marshal 再 Unmarshal”带来的二次编解码开销，
	// 在中大型结构和高频调用场景下能显著减少分配和 CPU 使用。
	return assignValue(rv.Elem(), converted)
}

func convertByType(val any, t reflect.Type) (any, error) {
	if t == nil {
		return val, nil
	}

	for t.Kind() == reflect.Pointer {
		// 类型推导阶段只负责“看到最终目标类型是什么”，
		// 不在这里分配指针实例，分配动作留到 assignValue。
		if val == nil {
			return nil, nil
		}
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		m, ok := val.(map[string]any)
		if !ok || val == nil {
			return val, nil
		}
		// 结构体：按 json tag 建立“字段名 -> 字段类型”映射，
		// 再递归处理每个字段值，提前把数字字符串转换到合适形态。
		return convertStructMap(m, t)
	case reflect.Slice, reflect.Array:
		arr, ok := val.([]any)
		if !ok || val == nil {
			return val, nil
		}
		out := make([]any, len(arr))
		for i := range arr {
			c, err := convertByType(arr[i], t.Elem())
			if err != nil {
				return nil, err
			}
			out[i] = c
		}
		return out, nil
	case reflect.Map:
		m, ok := val.(map[string]any)
		if !ok || val == nil {
			return val, nil
		}
		out := make(map[string]any, len(m))
		for k, v := range m {
			c, err := convertByType(v, t.Elem())
			if err != nil {
				return nil, err
			}
			out[k] = c
		}
		return out, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return toIntValue(val, t.Bits())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return toUintValue(val, t.Bits())
	case reflect.Float32, reflect.Float64:
		return toFloatValue(val, t.Bits())
	default:
		return val, nil
	}
}

func convertStructMap(m map[string]any, t reflect.Type) (map[string]any, error) {
	fields := buildJSONFieldTypeMap(t)
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	for k, v := range m {
		ft, ok := fields[k]
		if !ok {
			continue
		}
		c, err := convertByType(v, ft)
		if err != nil {
			return nil, err
		}
		out[k] = c
	}
	return out, nil
}

func buildJSONFieldTypeMap(t reflect.Type) map[string]reflect.Type {
	out := make(map[string]reflect.Type)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		// 跳过不可导出的普通字段，避免反射写入 panic。
		// 匿名字段仍允许参与（与标准库处理方式一致）。
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}
		name, skip := parseJSONFieldName(f)
		if skip || name == "" {
			continue
		}
		out[name] = f.Type
	}
	return out
}

func parseJSONFieldName(f reflect.StructField) (string, bool) {
	tag := f.Tag.Get("json")
	if tag == "-" {
		// 显式忽略字段。
		return "", true
	}
	if tag == "" {
		// 无 tag 时使用字段名（与 encoding/json 一致）。
		return f.Name, false
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		// json:",omitempty" 这类写法，名字为空时仍回退字段名。
		return f.Name, false
	}
	return name, false
}

func assignValue(dst reflect.Value, src any) error {
	if !dst.IsValid() {
		return nil
	}

	for dst.Kind() == reflect.Pointer {
		// 赋值阶段遇到指针字段时按需分配，保证后续 Elem() 可写。
		// 若源值为 nil，则保留/设置为 nil 指针。
		if src == nil {
			dst.Set(reflect.Zero(dst.Type()))
			return nil
		}
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		dst = dst.Elem()
	}

	if src == nil {
		// 对非指针目标，nil 值统一回落到零值。
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}

	switch dst.Kind() {
	case reflect.Struct:
		m, ok := src.(map[string]any)
		if !ok {
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "source is not object for struct field",
				Expected: dst.Type().String(),
				Value:    reflect.TypeOf(src).String(),
			}
		}
		// 对象 -> struct：按字段映射递归赋值。
		// 未出现的字段保持原值/零值，不额外覆盖。
		return assignStruct(dst, m)
	case reflect.Slice:
		arr, ok := src.([]any)
		if !ok {
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "source is not array for slice field",
				Expected: dst.Type().String(),
				Value:    reflect.TypeOf(src).String(),
			}
		}
		// 切片长度与输入数组保持一致，逐项递归赋值。
		s := reflect.MakeSlice(dst.Type(), len(arr), len(arr))
		for i := range arr {
			if err := assignValue(s.Index(i), arr[i]); err != nil {
				return err
			}
		}
		dst.Set(s)
		return nil
	case reflect.Array:
		arr, ok := src.([]any)
		if !ok {
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "source is not array for array field",
				Expected: dst.Type().String(),
				Value:    reflect.TypeOf(src).String(),
			}
		}
		// 固定数组按“最短长度”写入，超出部分忽略，不足部分保留零值。
		n := dst.Len()
		if len(arr) < n {
			n = len(arr)
		}
		for i := 0; i < n; i++ {
			if err := assignValue(dst.Index(i), arr[i]); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		m, ok := src.(map[string]any)
		if !ok {
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "source is not object for map field",
				Expected: dst.Type().String(),
				Value:    reflect.TypeOf(src).String(),
			}
		}
		if dst.Type().Key().Kind() != reflect.String {
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "json object only supports string map keys",
				Expected: dst.Type().String(),
			}
		}
		// map 值类型继续走 assignValue，确保嵌套数字字符串也能自动转换。
		// 这里仅支持 string key，因为 JSON object 的 key 本质就是 string。
		mv := reflect.MakeMapWithSize(dst.Type(), len(m))
		for k, v := range m {
			elem := reflect.New(dst.Type().Elem()).Elem()
			if err := assignValue(elem, v); err != nil {
				return err
			}
			mv.SetMapIndex(reflect.ValueOf(k), elem)
		}
		dst.Set(mv)
		return nil
	case reflect.Interface:
		// interface{} 直接承接转换后的值。
		dst.Set(reflect.ValueOf(src))
		return nil
	case reflect.String:
		s, ok := src.(string)
		if !ok {
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "source is not string",
				Expected: "string",
				Value:    reflect.TypeOf(src).String(),
			}
		}
		dst.SetString(s)
		return nil
	case reflect.Bool:
		switch b := src.(type) {
		case bool:
			dst.SetBool(b)
			return nil
		case string:
			// 对 bool 增强支持：允许 "true"/"false" 字符串自动转换。
			parsed, err := strconv.ParseBool(strings.TrimSpace(b))
			if err != nil {
				return &Error{
					Code:     ErrCodeTypeMismatch,
					Message:  "cannot convert string to bool",
					Expected: "bool",
					Value:    b,
					Cause:    err,
				}
			}
			dst.SetBool(parsed)
			return nil
		default:
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "source is not bool",
				Expected: "bool",
				Value:    reflect.TypeOf(src).String(),
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 数字族统一先收敛为 int64，再根据目标位宽 SetInt。
		v, err := toIntValue(src, dst.Type().Bits())
		if err != nil {
			return err
		}
		switch n := v.(type) {
		case int64:
			dst.SetInt(n)
			return nil
		default:
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "converted value is not int64",
				Expected: "int64",
				Value:    reflect.TypeOf(v).String(),
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// 无符号数字族统一收敛为 uint64，再根据目标位宽 SetUint。
		v, err := toUintValue(src, dst.Type().Bits())
		if err != nil {
			return err
		}
		switch n := v.(type) {
		case uint64:
			dst.SetUint(n)
			return nil
		default:
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "converted value is not uint64",
				Expected: "uint64",
				Value:    reflect.TypeOf(v).String(),
			}
		}
	case reflect.Float32, reflect.Float64:
		// 浮点族统一收敛为 float64，再按目标类型 SetFloat。
		v, err := toFloatValue(src, dst.Type().Bits())
		if err != nil {
			return err
		}
		switch n := v.(type) {
		case float64:
			dst.SetFloat(n)
			return nil
		default:
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "converted value is not float64",
				Expected: "float64",
				Value:    reflect.TypeOf(v).String(),
			}
		}
	default:
		// 兜底路径：复用标准库处理复杂/自定义类型（如 time.Time、
		// 或实现了 UnmarshalJSON 的业务类型），以减少手写分支复杂度。
		b, err := json.Marshal(src)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, dst.Addr().Interface())
	}
}

func assignStruct(dst reflect.Value, src map[string]any) error {
	t := dst.Type()
	fields := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}
		name, skip := parseJSONFieldName(f)
		if skip || name == "" {
			continue
		}
		fields[name] = i
	}
	// 仅处理 JSON 中存在的键；未知字段保持忽略，与标准库行为一致。
	// 如果希望“未知字段报错”，可在此增加严格模式开关。
	for key, value := range src {
		idx, ok := fields[key]
		if !ok {
			continue
		}
		if err := assignValue(dst.Field(idx), value); err != nil {
			return err
		}
	}
	return nil
}

func toIntValue(val any, bits int) (any, error) {
	if bits == 0 {
		bits = strconv.IntSize
	}
	switch x := val.(type) {
	case string:
		// 核心增强点：数字字符串 -> 整型。
		n, err := strconv.ParseInt(strings.TrimSpace(x), 10, bits)
		if err != nil {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert string to int",
				Expected: "int",
				Value:    x,
				Cause:    err,
			}
		}
		return n, nil
	case json.Number:
		// UseNumber 场景：避免中间态 float64 带来的精度损失。
		n, err := strconv.ParseInt(x.String(), 10, bits)
		if err != nil {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert number to int",
				Expected: "int",
				Value:    x.String(),
				Cause:    err,
			}
		}
		return n, nil
	case float64:
		// 兼容少数非 UseNumber 路径；直接收敛到 int64。
		// 与 Go 原生转换一致，小数会被截断。
		return int64(x), nil
	default:
		return val, nil
	}
}

func toUintValue(val any, bits int) (any, error) {
	if bits == 0 {
		bits = strconv.IntSize
	}
	switch x := val.(type) {
	case string:
		// 核心增强点：数字字符串 -> 无符号整型。
		n, err := strconv.ParseUint(strings.TrimSpace(x), 10, bits)
		if err != nil {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert string to uint",
				Expected: "uint",
				Value:    x,
				Cause:    err,
			}
		}
		return n, nil
	case json.Number:
		// UseNumber 场景下的无符号整型解析。
		n, err := strconv.ParseUint(x.String(), 10, bits)
		if err != nil {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert number to uint",
				Expected: "uint",
				Value:    x.String(),
				Cause:    err,
			}
		}
		return n, nil
	case float64:
		if x < 0 {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert negative float to uint",
				Expected: "uint",
				Value:    strconv.FormatFloat(x, 'f', -1, 64),
			}
		}
		// 兼容少数非 UseNumber 路径；小数部分会被截断。
		return uint64(x), nil
	default:
		return val, nil
	}
}

func toFloatValue(val any, bits int) (any, error) {
	switch x := val.(type) {
	case string:
		// 核心增强点：数字字符串 -> 浮点型。
		n, err := strconv.ParseFloat(strings.TrimSpace(x), bits)
		if err != nil {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert string to float",
				Expected: "float",
				Value:    x,
				Cause:    err,
			}
		}
		return n, nil
	case json.Number:
		// UseNumber 场景下直接按目标精度解析。
		n, err := strconv.ParseFloat(x.String(), bits)
		if err != nil {
			return nil, &Error{
				Code:     ErrCodeConvertNumber,
				Message:  "cannot convert number to float",
				Expected: "float",
				Value:    x.String(),
				Cause:    err,
			}
		}
		return n, nil
	case float64:
		return x, nil
	default:
		return val, nil
	}
}

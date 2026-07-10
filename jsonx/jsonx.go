package jsonx

import (
	"bytes"
	"encoding/json"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal 是增强版 JSON 反序列化。
// 支持常见宽松转换：字符串↔数字、数字/布尔→字符串、布尔接受 0/1 等。
// 浮点赋给整型时若含小数部分会报错，不会静默截断。
// 解到 any / map[string]any 时数字为 float64（与 encoding/json 默认一致）。
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
	// 拒绝尾部多余内容，避免 {"a":1}{"b":2} 这类输入被静默截断。
	if rest := bytes.TrimSpace(data[dec.InputOffset():]); len(rest) > 0 {
		return &Error{
			Code:    ErrCodeInvalidJSON,
			Message: "trailing data after JSON value",
		}
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
	case reflect.Interface:
		// any / interface{}：将 json.Number 归一为 float64，与 encoding/json 默认行为对齐。
		return normalizeJSONValue(val), nil
	default:
		return val, nil
	}
}

// normalizeJSONValue 将 UseNumber 产生的 json.Number 转为 float64（递归处理 map/slice），
// 使解到 any / map[string]any 时与 encoding/json 默认行为一致。
func normalizeJSONValue(v any) any {
	switch x := v.(type) {
	case json.Number:
		f, err := x.Float64()
		if err != nil {
			return x.String()
		}
		return f
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, vv := range x {
			out[k] = normalizeJSONValue(vv)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, vv := range x {
			out[i] = normalizeJSONValue(vv)
		}
		return out
	default:
		return v
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
	collectJSONFieldTypes(t, out)
	return out
}

// collectJSONFieldTypes 收集 JSON 字段类型；匿名嵌入且无显式 json 名时展开内层字段。
// 显式字段优先于嵌入字段（与 encoding/json 一致）。
func collectJSONFieldTypes(t reflect.Type, out map[string]reflect.Type) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}
		if f.Anonymous && shouldInlineAnonymous(f) {
			continue
		}
		name, skip := parseJSONFieldName(f)
		if skip || name == "" {
			continue
		}
		out[name] = f.Type
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !(f.Anonymous && shouldInlineAnonymous(f)) {
			continue
		}
		ft := f.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() != reflect.Struct {
			continue
		}
		inner := make(map[string]reflect.Type)
		collectJSONFieldTypes(ft, inner)
		for name, typ := range inner {
			if _, exists := out[name]; !exists {
				out[name] = typ
			}
		}
	}
}

func shouldInlineAnonymous(f reflect.StructField) bool {
	tag := f.Tag.Get("json")
	if tag == "-" {
		return false
	}
	if tag == "" {
		return true
	}
	name := strings.Split(tag, ",")[0]
	return name == ""
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
		if m, ok := src.(map[string]any); ok {
			// 对象 -> struct：按字段映射递归赋值。
			// 未出现的字段保持原值/零值，不额外覆盖。
			return assignStruct(dst, m)
		}
		// time.Time、自定义 Unmarshaler 等：回退标准库解码。
		return assignViaJSON(dst, src)
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
		// interface{} 承接归一化后的值（数字为 float64，与 encoding/json 一致）。
		dst.Set(reflect.ValueOf(normalizeJSONValue(src)))
		return nil
	case reflect.String:
		switch s := src.(type) {
		case string:
			dst.SetString(s)
			return nil
		case json.Number:
			dst.SetString(s.String())
			return nil
		case float64:
			dst.SetString(strconv.FormatFloat(s, 'f', -1, 64))
			return nil
		case bool:
			dst.SetString(strconv.FormatBool(s))
			return nil
		default:
			return &Error{
				Code:     ErrCodeTypeMismatch,
				Message:  "cannot convert value to string",
				Expected: "string",
				Value:    reflect.TypeOf(src).String(),
			}
		}
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
		case json.Number:
			n, err := b.Int64()
			if err != nil || (n != 0 && n != 1) {
				return &Error{
					Code:     ErrCodeTypeMismatch,
					Message:  "cannot convert number to bool",
					Expected: "bool",
					Value:    b.String(),
					Cause:    err,
				}
			}
			dst.SetBool(n == 1)
			return nil
		case float64:
			if b != 0 && b != 1 {
				return &Error{
					Code:     ErrCodeTypeMismatch,
					Message:  "cannot convert float to bool",
					Expected: "bool",
					Value:    strconv.FormatFloat(b, 'f', -1, 64),
				}
			}
			dst.SetBool(b == 1)
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
	fields := make(map[string][]int)
	collectJSONFieldIndexes(dst.Type(), fields, nil)
	// 仅处理 JSON 中存在的键；未知字段保持忽略，与标准库行为一致。
	for key, value := range src {
		idx, ok := fields[key]
		if !ok {
			continue
		}
		if err := assignValue(dst.FieldByIndex(idx), value); err != nil {
			return err
		}
	}
	return nil
}

func collectJSONFieldIndexes(t reflect.Type, out map[string][]int, index []int) {
	// 先注册本层显式字段，再展开匿名嵌入，保证外层显式字段优先（与 encoding/json 一致）。
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}
		if f.Anonymous && shouldInlineAnonymous(f) {
			continue
		}
		name, skip := parseJSONFieldName(f)
		if skip || name == "" {
			continue
		}
		out[name] = append(append([]int(nil), index...), i)
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !(f.Anonymous && shouldInlineAnonymous(f)) {
			continue
		}
		ft := f.Type
		for ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() != reflect.Struct {
			continue
		}
		idx := append(append([]int(nil), index...), i)
		inner := make(map[string][]int)
		collectJSONFieldIndexes(ft, inner, idx)
		for name, fieldIdx := range inner {
			if _, exists := out[name]; !exists {
				out[name] = fieldIdx
			}
		}
	}
}

func assignViaJSON(dst reflect.Value, src any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return &Error{
			Code:     ErrCodeTypeMismatch,
			Message:  "cannot marshal value for struct field",
			Expected: dst.Type().String(),
			Value:    reflect.TypeOf(src).String(),
			Cause:    err,
		}
	}
	if !dst.CanAddr() {
		return &Error{
			Code:     ErrCodeTypeMismatch,
			Message:  "struct field is not addressable",
			Expected: dst.Type().String(),
		}
	}
	if err := json.Unmarshal(b, dst.Addr().Interface()); err != nil {
		return &Error{
			Code:     ErrCodeTypeMismatch,
			Message:  "cannot unmarshal value into struct field",
			Expected: dst.Type().String(),
			Value:    string(b),
			Cause:    err,
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
		return floatToInt64(x, bits)
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
		return floatToUint64(x, bits)
	default:
		return val, nil
	}
}

func floatToInt64(x float64, bits int) (int64, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert NaN/Inf to int",
			Expected: "int",
			Value:    strconv.FormatFloat(x, 'f', -1, 64),
		}
	}
	if math.Trunc(x) != x {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert non-integer number to int",
			Expected: "int",
			Value:    strconv.FormatFloat(x, 'f', -1, 64),
		}
	}
	s := strconv.FormatFloat(x, 'f', -1, 64)
	n, err := strconv.ParseInt(s, 10, bits)
	if err != nil {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert float to int",
			Expected: "int",
			Value:    s,
			Cause:    err,
		}
	}
	return n, nil
}

func floatToUint64(x float64, bits int) (uint64, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert NaN/Inf to uint",
			Expected: "uint",
			Value:    strconv.FormatFloat(x, 'f', -1, 64),
		}
	}
	if x < 0 {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert negative float to uint",
			Expected: "uint",
			Value:    strconv.FormatFloat(x, 'f', -1, 64),
		}
	}
	if math.Trunc(x) != x {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert non-integer number to uint",
			Expected: "uint",
			Value:    strconv.FormatFloat(x, 'f', -1, 64),
		}
	}
	s := strconv.FormatFloat(x, 'f', -1, 64)
	n, err := strconv.ParseUint(s, 10, bits)
	if err != nil {
		return 0, &Error{
			Code:     ErrCodeConvertNumber,
			Message:  "cannot convert float to uint",
			Expected: "uint",
			Value:    s,
			Cause:    err,
		}
	}
	return n, nil
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

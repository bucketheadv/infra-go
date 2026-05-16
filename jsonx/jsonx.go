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
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil {
		return err
	}

	converted, err := convertByType(raw, rv.Type().Elem())
	if err != nil {
		return err
	}

	return assignValue(rv.Elem(), converted)
}

func convertByType(val any, t reflect.Type) (any, error) {
	if t == nil {
		return val, nil
	}

	for t.Kind() == reflect.Pointer {
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
		return "", true
	}
	if tag == "" {
		return f.Name, false
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return f.Name, false
	}
	return name, false
}

func assignValue(dst reflect.Value, src any) error {
	if !dst.IsValid() {
		return nil
	}

	for dst.Kind() == reflect.Pointer {
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
		return uint64(x), nil
	default:
		return val, nil
	}
}

func toFloatValue(val any, bits int) (any, error) {
	switch x := val.(type) {
	case string:
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

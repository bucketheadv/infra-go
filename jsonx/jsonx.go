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

	b, err := json.Marshal(converted)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
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
	default:
		return val, nil
	}
}

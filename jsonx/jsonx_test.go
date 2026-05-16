package jsonx

import (
	"errors"
	"testing"
)

type scoreInfo struct {
	PassRate float64 `json:"passRate"`
}

type userProfile struct {
	ID      int               `json:"id"`
	Age     int32             `json:"age"`
	Score   float64           `json:"score"`
	Tags    []int             `json:"tags"`
	Ratios  map[string]uint64 `json:"ratios"`
	Nested  scoreInfo         `json:"nested"`
	Balance *float32          `json:"balance"`
	Name    string            `json:"name"`
}

// TestUnmarshalAutoConvertNumbers 验证数字字符串自动转换。
func TestUnmarshalAutoConvertNumbers(t *testing.T) {
	input := []byte(`{
		"id":"1001",
		"age":"20",
		"score":"98.5",
		"tags":["1","2","3"],
		"ratios":{"a":"10","b":"20"},
		"nested":{"passRate":"99.9"},
		"balance":"88.8",
		"name":"alice"
	}`)

	var got userProfile
	if err := Unmarshal(input, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if got.ID != 1001 || got.Age != 20 {
		t.Fatalf("unexpected id/age: %+v", got)
	}
	if got.Score != 98.5 {
		t.Fatalf("unexpected score: %v", got.Score)
	}
	if len(got.Tags) != 3 || got.Tags[0] != 1 || got.Tags[2] != 3 {
		t.Fatalf("unexpected tags: %#v", got.Tags)
	}
	if got.Ratios["a"] != 10 || got.Ratios["b"] != 20 {
		t.Fatalf("unexpected ratios: %#v", got.Ratios)
	}
	if got.Nested.PassRate != 99.9 {
		t.Fatalf("unexpected nested pass rate: %v", got.Nested.PassRate)
	}
	if got.Balance == nil || *got.Balance != float32(88.8) {
		t.Fatalf("unexpected balance: %v", got.Balance)
	}
	if got.Name != "alice" {
		t.Fatalf("unexpected name: %s", got.Name)
	}
}

// TestUnmarshalKeepsNormalNumber 验证原本就是数字时仍可正常解析。
func TestUnmarshalKeepsNormalNumber(t *testing.T) {
	input := []byte(`{"id":1,"age":2,"score":3.5}`)
	var got userProfile
	if err := Unmarshal(input, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.ID != 1 || got.Age != 2 || got.Score != 3.5 {
		t.Fatalf("unexpected value: %+v", got)
	}
}

// TestUnmarshalInvalidNumberString 验证数字字段遇到非法字符串时返回错误。
func TestUnmarshalInvalidNumberString(t *testing.T) {
	input := []byte(`{"id":"abc"}`)
	var got userProfile
	err := Unmarshal(input, &got)
	if err == nil {
		t.Fatalf("expected error")
	}
	var je *Error
	if !errors.As(err, &je) {
		t.Fatalf("expected jsonx.Error, got=%T", err)
	}
	if je.Code != ErrCodeConvertNumber {
		t.Fatalf("unexpected error code: %s", je.Code)
	}
}

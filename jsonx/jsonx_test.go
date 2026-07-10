package jsonx

import (
	"errors"
	"testing"
	"time"
)

type scoreInfo struct {
	// PassRate 通过率。
	PassRate float64 `json:"passRate"`
}

type userProfile struct {
	// ID 用户 ID。
	ID int `json:"id"`
	// Age 年龄。
	Age int32 `json:"age"`
	// Score 分数。
	Score float64 `json:"score"`
	// Tags 标签列表。
	Tags []int `json:"tags"`
	// Ratios 比率映射。
	Ratios map[string]uint64 `json:"ratios"`
	// Nested 嵌套分数信息。
	Nested scoreInfo `json:"nested"`
	// Balance 余额（可空）。
	Balance *float32 `json:"balance"`
	// Name 姓名。
	Name string `json:"name"`
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

func TestUnmarshalBoolNumber(t *testing.T) {
	type flag struct {
		OK bool `json:"ok"`
	}
	var got flag
	if err := Unmarshal([]byte(`{"ok":1}`), &got); err != nil || !got.OK {
		t.Fatalf("bool from 1 failed: %+v err=%v", got, err)
	}
	if err := Unmarshal([]byte(`{"ok":0}`), &got); err != nil || got.OK {
		t.Fatalf("bool from 0 failed: %+v err=%v", got, err)
	}
}

func TestUnmarshalNumberToString(t *testing.T) {
	type row struct {
		ID   string `json:"id"`
		OK   string `json:"ok"`
		Rate string `json:"rate"`
	}
	var got row
	if err := Unmarshal([]byte(`{"id":123,"ok":true,"rate":1.5}`), &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.ID != "123" || got.OK != "true" || got.Rate != "1.5" {
		t.Fatalf("got %+v", got)
	}
}

func TestUnmarshalRejectsFractionalInt(t *testing.T) {
	type row struct {
		ID int `json:"id"`
	}
	var got row
	err := Unmarshal([]byte(`{"id":1.9}`), &got)
	if err == nil {
		t.Fatal("expected error for fractional int")
	}
	var je *Error
	if !errors.As(err, &je) || je.Code != ErrCodeConvertNumber {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalRejectsTrailingData(t *testing.T) {
	var got map[string]any
	err := Unmarshal([]byte(`{"a":1}{"b":2}`), &got)
	if err == nil {
		t.Fatal("expected trailing data error")
	}
	var je *Error
	if !errors.As(err, &je) || je.Code != ErrCodeInvalidJSON {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUnmarshalMapAnyUsesFloat64(t *testing.T) {
	var got map[string]any
	if err := Unmarshal([]byte(`{"n":1,"x":1.5}`), &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if _, ok := got["n"].(float64); !ok {
		t.Fatalf("n type = %T, want float64", got["n"])
	}
	if got["x"].(float64) != 1.5 {
		t.Fatalf("x = %v", got["x"])
	}
}

func TestUnmarshalTimeAndEmbedded(t *testing.T) {
	type Inner struct {
		Name string `json:"name"`
	}
	type Outer struct {
		Inner
		At time.Time `json:"at"`
	}

	input := []byte(`{"name":"alice","at":"2026-05-16T10:00:00Z"}`)
	var got Outer
	if err := Unmarshal(input, &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.Name != "alice" {
		t.Fatalf("embedded name = %q", got.Name)
	}
	want := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	if !got.At.Equal(want) {
		t.Fatalf("at = %v, want %v", got.At, want)
	}
}

func TestUnmarshalExplicitFieldOverridesEmbedded(t *testing.T) {
	type Inner struct {
		V int `json:"v"`
	}
	type Outer struct {
		Inner
		V int `json:"v"`
	}
	var got Outer
	if err := Unmarshal([]byte(`{"v":9}`), &got); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if got.V != 9 {
		t.Fatalf("outer V = %d, want 9", got.V)
	}
	if got.Inner.V != 0 {
		t.Fatalf("embedded V should stay 0, got %d", got.Inner.V)
	}
}

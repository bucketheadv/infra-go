package basic

import (
	"encoding/json"
	"testing"
)

// TestPairFields 验证 Pair 字段可正确保存左右值。
func TestPairFields(t *testing.T) {
	p := Pair[int, string]{
		Left:  7,
		Right: "seven",
	}
	if p.Left != 7 || p.Right != "seven" {
		t.Fatalf("unexpected pair value: %+v", p)
	}
}

// TestPairJSONTags 验证 Pair 的 JSON 标签为 left/right。
func TestPairJSONTags(t *testing.T) {
	p := Pair[int, string]{
		Left:  1,
		Right: "ok",
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(data) != `{"left":1,"right":"ok"}` {
		t.Fatalf("unexpected json: %s", string(data))
	}
}

// TestTripleFields 验证 Triple 字段可正确保存左中右值。
func TestTripleFields(t *testing.T) {
	tri := Triple[int, string, bool]{
		Left:   1,
		Middle: "mid",
		Right:  true,
	}
	if tri.Left != 1 || tri.Middle != "mid" || !tri.Right {
		t.Fatalf("unexpected triple value: %+v", tri)
	}
}

// TestTripleJSONTags 验证 Triple 的 JSON 标签为 left/middle/right。
func TestTripleJSONTags(t *testing.T) {
	tri := Triple[int, string, bool]{
		Left:   1,
		Middle: "m",
		Right:  true,
	}
	data, err := json.Marshal(tri)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(data) != `{"left":1,"middle":"m","right":true}` {
		t.Fatalf("unexpected json: %s", string(data))
	}
}

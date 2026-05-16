package collection

import (
	"testing"
)

// TestSortedMapTraversal 验证 map 在升序/降序遍历下的回调顺序。
func TestSortedMapTraversal(t *testing.T) {
	m := map[int]string{
		3: "three",
		1: "one",
		2: "two",
	}

	// reverse=false：按 key 升序输出。
	var ascKeys []int
	var ascValues []string
	SortedMapTraversal(m, false, func(k int, v string) {
		ascKeys = append(ascKeys, k)
		ascValues = append(ascValues, v)
	})
	assertIntSliceEqual(t, ascKeys, []int{1, 2, 3})
	assertStringSliceEqual(t, ascValues, []string{"one", "two", "three"})

	// reverse=true：按 key 降序输出。
	var descKeys []int
	SortedMapTraversal(m, true, func(k int, v string) {
		descKeys = append(descKeys, k)
	})
	assertIntSliceEqual(t, descKeys, []int{3, 2, 1})

	// 空 map：回调不应被调用。
	m = map[int]string{}
	called := false
	SortedMapTraversal(m, false, func(k int, v string) {
		called = true
	})
	if called {
		t.Fatalf("callback should not be called for empty map")
	}
}

// assertIntSliceEqual 校验 int 切片完全一致。
func assertIntSliceEqual(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice mismatch at %d: got=%d want=%d", i, got[i], want[i])
		}
	}
}

// assertStringSliceEqual 校验 string 切片完全一致。
func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice length mismatch: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice mismatch at %d: got=%s want=%s", i, got[i], want[i])
		}
	}
}

package collectionx

import (
	"reflect"
	"testing"
)

func TestKeysValuesMergeMaps(t *testing.T) {
	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"b": 3, "c": 4}

	keys := Keys(m1)
	if len(keys) != 2 {
		t.Fatalf("Keys() len = %d", len(keys))
	}

	vals := Values(m1)
	if len(vals) != 2 {
		t.Fatalf("Values() len = %d", len(vals))
	}

	merged := MergeMaps(m1, m2)
	want := map[string]int{"a": 1, "b": 3, "c": 4}
	if !reflect.DeepEqual(merged, want) {
		t.Fatalf("MergeMaps() = %v, want %v", merged, want)
	}
}

func TestSortedMapTraversal(t *testing.T) {
	m := map[int]string{
		3: "three",
		1: "one",
		2: "two",
	}

	var ascKeys []int
	var ascValues []string
	SortedMapTraversal(m, false, func(k int, v string) {
		ascKeys = append(ascKeys, k)
		ascValues = append(ascValues, v)
	})
	assertIntSliceEqual(t, ascKeys, []int{1, 2, 3})
	assertStringSliceEqual(t, ascValues, []string{"one", "two", "three"})

	var descKeys []int
	SortedMapTraversal(m, true, func(k int, v string) {
		descKeys = append(descKeys, k)
	})
	assertIntSliceEqual(t, descKeys, []int{3, 2, 1})

	m = map[int]string{}
	called := false
	SortedMapTraversal(m, false, func(k int, v string) {
		called = true
	})
	if called {
		t.Fatalf("callback should not be called for empty map")
	}
}

func TestGetOrDefaultPickOmit(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	if got := GetOrDefault(m, "b", 0); got != 2 {
		t.Fatalf("GetOrDefault() = %d", got)
	}
	if got := GetOrDefault(m, "x", 9); got != 9 {
		t.Fatalf("GetOrDefault(default) = %d", got)
	}
	if !reflect.DeepEqual(Pick(m, "a", "c", "x"), map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("Pick() failed")
	}
	if !reflect.DeepEqual(Omit(m, "b"), map[string]int{"a": 1, "c": 3}) {
		t.Fatalf("Omit() failed")
	}
}

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

package collection

import (
	"testing"
)

// TestPartition 覆盖分片的常见输入场景与边界行为。
func TestPartition(t *testing.T) {
	// 表驱动：验证空输入、常规分片以及 size 特殊值。
	tests := []struct {
		name string
		arr  []int
		size int
		want [][]int
	}{
		{
			name: "empty array returns empty slice",
			arr:  nil,
			size: 2,
			want: [][]int{},
		},
		{
			name: "array shorter than chunk size",
			arr:  []int{1},
			size: 2,
			want: [][]int{{1}},
		},
		{
			name: "normal chunking",
			arr:  []int{1, 2, 3, 4, 5},
			size: 2,
			want: [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name: "size one",
			arr:  []int{1, 2, 3, 4, 5},
			size: 1,
			want: [][]int{{1}, {2}, {3}, {4}, {5}},
		},
		{
			name: "size equals length",
			arr:  []int{1, 2, 3},
			size: 3,
			want: [][]int{{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 核对二维切片结构与值是否完全一致。
			got := Partition(tt.arr, tt.size)
			assert2DIntEqual(t, got, tt.want)
		})
	}
}

// TestPartitionPanicWhenSizeNonPositive 验证 size 非法时的当前行为。
func TestPartitionPanicWhenSizeNonPositive(t *testing.T) {
	t.Run("size=0 returns empty", func(t *testing.T) {
		got := Partition([]int{1, 2, 3}, 0)
		if len(got) != 0 {
			t.Fatalf("expected empty result when size=0, got=%v", got)
		}
	})

	t.Run("size<0 returns empty", func(t *testing.T) {
		got := Partition([]int{1, 2, 3}, -1)
		if len(got) != 0 {
			t.Fatalf("expected empty result when size<0, got=%v", got)
		}
	})
}

// TestGroupBy 验证不同类型元素按 key 分组后的结果与顺序。
func TestGroupBy(t *testing.T) {
	// 数字按奇偶分组。
	arr := []int{1, 2, 3, 4, 5}
	got := GroupBy(arr, func(v int) int { return v % 2 })
	want := map[int][]int{
		0: {2, 4},
		1: {1, 3, 5},
	}
	assertMapSlicesEqual(t, got, want)

	// 字符串按自身值分组（含重复值）。
	arr2 := []string{"apple", "banana", "cherry", "apple", "banana"}
	got2 := GroupBy(arr2, func(v string) string { return v })
	want2 := map[string][]string{
		"apple":  {"apple", "apple"},
		"banana": {"banana", "banana"},
		"cherry": {"cherry"},
	}
	assertMapSlicesEqual(t, got2, want2)

	// 结构体按字段分组。
	arr3 := []struct {
		id   int
		name string
	}{
		{1, "Alice"},
		{2, "Bob"},
		{3, "Charlie"},
		{1, "David"},
	}
	got3 := GroupBy(arr3, func(v struct {
		id   int
		name string
	}) int {
		return v.id
	})
	want3 := map[int][]struct {
		id   int
		name string
	}{
		1: {
			{1, "Alice"},
			{1, "David"},
		},
		2: {
			{2, "Bob"},
		},
		3: {
			{3, "Charlie"},
		},
	}
	assertStructGroupEqual(t, got3, want3)
}

// TestArrayToMap 验证同 key 保留首值/覆盖末值两种策略。
func TestArrayToMap(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}

	users := []user{
		{ID: 1, Name: "first"},
		{ID: 2, Name: "alpha"},
		{ID: 2, Name: "beta"},
		{ID: 3, Name: "x"},
	}

	// coverExists=false：保留首次出现的元素。
	gotKeepFirst := ArrayToMap(users, false, func(u user) int { return u.ID })
	wantKeepFirst := map[int]user{
		1: {ID: 1, Name: "first"},
		2: {ID: 2, Name: "alpha"},
		3: {ID: 3, Name: "x"},
	}
	assertUserMapEqual(t, gotKeepFirst, wantKeepFirst)

	// coverExists=true：后出现元素覆盖之前的值。
	gotCover := ArrayToMap(users, true, func(u user) int { return u.ID })
	wantCover := map[int]user{
		1: {ID: 1, Name: "first"},
		2: {ID: 2, Name: "beta"},
		3: {ID: 3, Name: "x"},
	}
	assertUserMapEqual(t, gotCover, wantCover)

	// 空输入返回空 map。
	empty := ArrayToMap([]user{}, false, func(u user) int { return u.ID })
	if len(empty) != 0 {
		t.Fatalf("expected empty map, got len=%d", len(empty))
	}
}

// assert2DIntEqual 校验二维 int 切片的结构和值完全一致。
func assert2DIntEqual(t *testing.T, got, want [][]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("inner length mismatch at %d: got=%d want=%d", i, len(got[i]), len(want[i]))
		}
		for j := range want[i] {
			if got[i][j] != want[i][j] {
				t.Fatalf("value mismatch at [%d][%d]: got=%d want=%d", i, j, got[i][j], want[i][j])
			}
		}
	}
}

// assertMapSlicesEqual 校验 map[key][]value 的键集合与各组顺序内容。
func assertMapSlicesEqual[T comparable](t *testing.T, got, want map[T][]T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("map length mismatch: got=%d want=%d", len(got), len(want))
	}
	for k, wantSlice := range want {
		gotSlice, ok := got[k]
		if !ok {
			t.Fatalf("missing key: %v", k)
		}
		if len(gotSlice) != len(wantSlice) {
			t.Fatalf("slice length mismatch for key=%v: got=%d want=%d", k, len(gotSlice), len(wantSlice))
		}
		for i := range wantSlice {
			if gotSlice[i] != wantSlice[i] {
				t.Fatalf("slice value mismatch for key=%v at index=%d: got=%v want=%v", k, i, gotSlice[i], wantSlice[i])
			}
		}
	}
}

// assertStructGroupEqual 校验结构体分组结果。
func assertStructGroupEqual(t *testing.T, got, want map[int][]struct {
	id   int
	name string
}) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("map length mismatch: got=%d want=%d", len(got), len(want))
	}
	for k, wantSlice := range want {
		gotSlice, ok := got[k]
		if !ok {
			t.Fatalf("missing key: %d", k)
		}
		if len(gotSlice) != len(wantSlice) {
			t.Fatalf("slice length mismatch for key=%d: got=%d want=%d", k, len(gotSlice), len(wantSlice))
		}
		for i := range wantSlice {
			if gotSlice[i] != wantSlice[i] {
				t.Fatalf("slice value mismatch for key=%d at index=%d", k, i)
			}
		}
	}
}

// assertUserMapEqual 校验 map[int]V 的键值一致性。
func assertUserMapEqual[V comparable](t *testing.T, got, want map[int]V) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("map length mismatch: got=%d want=%d", len(got), len(want))
	}
	for k, wantV := range want {
		gotV, ok := got[k]
		if !ok {
			t.Fatalf("missing key: %d", k)
		}
		if gotV != wantV {
			t.Fatalf("value mismatch for key=%d: got=%v want=%v", k, gotV, wantV)
		}
	}
}


package collectionx

import (
	"reflect"
	"testing"
)

func TestIsEmpty(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		var nilSlice []int
		if !IsEmpty(nilSlice) {
			t.Fatalf("expected true for nil slice")
		}
		if !IsEmpty([]int{}) {
			t.Fatalf("expected true for empty slice")
		}
		if IsEmpty([]int{1}) {
			t.Fatalf("expected false for non-empty slice")
		}
	})

	t.Run("slice pointer", func(t *testing.T) {
		var nilPtr *[]int
		if !IsEmpty(nilPtr) {
			t.Fatalf("expected true for nil pointer")
		}
		var data []int
		if !IsEmpty(&data) {
			t.Fatalf("expected true for pointer to nil slice")
		}
		data = []int{}
		if !IsEmpty(&data) {
			t.Fatalf("expected true for pointer to empty slice")
		}
		data = []int{1}
		if IsEmpty(&data) {
			t.Fatalf("expected false for pointer to non-empty slice")
		}
	})

	t.Run("map", func(t *testing.T) {
		var nilMap map[string]int
		if !IsEmpty(nilMap) {
			t.Fatalf("expected true for nil map")
		}
		if !IsEmpty(map[string]int{}) {
			t.Fatalf("expected true for empty map")
		}
		if IsEmpty(map[string]int{"a": 1}) {
			t.Fatalf("expected false for non-empty map")
		}
	})

	t.Run("map pointer", func(t *testing.T) {
		var nilPtr *map[string]int
		if !IsEmpty(nilPtr) {
			t.Fatalf("expected true for nil pointer")
		}
		var data map[string]int
		if !IsEmpty(&data) {
			t.Fatalf("expected true for pointer to nil map")
		}
		data = map[string]int{}
		if !IsEmpty(&data) {
			t.Fatalf("expected true for pointer to empty map")
		}
		data = map[string]int{"a": 1}
		if IsEmpty(&data) {
			t.Fatalf("expected false for pointer to non-empty map")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		if IsEmpty("hello") {
			t.Fatalf("expected false for string")
		}
		if IsEmpty(0) {
			t.Fatalf("expected false for int")
		}
	})
}

func TestMapFilterUnique(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	got := Map(arr, func(v int) int { return v * 2 })
	want := []int{2, 4, 6, 8, 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Map() = %v, want %v", got, want)
	}

	filtered := Filter(arr, func(v int) bool { return v%2 == 1 })
	if !reflect.DeepEqual(filtered, []int{1, 3, 5}) {
		t.Fatalf("Filter() = %v", filtered)
	}

	unique := Unique([]int{1, 2, 2, 3, 1})
	if !reflect.DeepEqual(unique, []int{1, 2, 3}) {
		t.Fatalf("Unique() = %v", unique)
	}
}

func TestContainsFindEverySome(t *testing.T) {
	arr := []int{1, 2, 3}
	if !Contains(arr, 2) || Contains(arr, 9) {
		t.Fatalf("Contains() mismatch")
	}

	v, ok := Find(arr, func(x int) bool { return x > 2 })
	if !ok || v != 3 {
		t.Fatalf("Find() = (%v, %v)", v, ok)
	}

	if !Every(arr, func(x int) bool { return x > 0 }) {
		t.Fatalf("Every() should be true")
	}
	if Every(arr, func(x int) bool { return x%2 == 0 }) {
		t.Fatalf("Every() should be false")
	}
	if !Some(arr, func(x int) bool { return x == 2 }) {
		t.Fatalf("Some() should be true")
	}
}

func TestReverseFlatten(t *testing.T) {
	arr := []int{1, 2, 3}
	Reverse(arr)
	if !reflect.DeepEqual(arr, []int{3, 2, 1}) {
		t.Fatalf("Reverse() = %v", arr)
	}

	flat := Flatten([][]int{{1, 2}, {}, {3}})
	if !reflect.DeepEqual(flat, []int{1, 2, 3}) {
		t.Fatalf("Flatten() = %v", flat)
	}
}

func TestPartition(t *testing.T) {
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
			got := Partition(tt.arr, tt.size)
			assert2DIntEqual(t, got, tt.want)
		})
	}
}

func TestPartitionSizeNonPositive(t *testing.T) {
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

func TestGroupBy(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}
	got := GroupBy(arr, func(v int) int { return v % 2 })
	want := map[int][]int{
		0: {2, 4},
		1: {1, 3, 5},
	}
	assertMapSlicesEqual(t, got, want)

	arr2 := []string{"apple", "banana", "cherry", "apple", "banana"}
	got2 := GroupBy(arr2, func(v string) string { return v })
	want2 := map[string][]string{
		"apple":  {"apple", "apple"},
		"banana": {"banana", "banana"},
		"cherry": {"cherry"},
	}
	assertMapSlicesEqual(t, got2, want2)

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

	gotKeepFirst := ArrayToMap(users, false, func(u user) int { return u.ID })
	wantKeepFirst := map[int]user{
		1: {ID: 1, Name: "first"},
		2: {ID: 2, Name: "alpha"},
		3: {ID: 3, Name: "x"},
	}
	assertUserMapEqual(t, gotKeepFirst, wantKeepFirst)

	gotCover := ArrayToMap(users, true, func(u user) int { return u.ID })
	wantCover := map[int]user{
		1: {ID: 1, Name: "first"},
		2: {ID: 2, Name: "beta"},
		3: {ID: 3, Name: "x"},
	}
	assertUserMapEqual(t, gotCover, wantCover)

	empty := ArrayToMap([]user{}, false, func(u user) int { return u.ID })
	if len(empty) != 0 {
		t.Fatalf("expected empty map, got len=%d", len(empty))
	}
}

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

func TestReducePaginateSortBy(t *testing.T) {
	sum := Reduce([]int{1, 2, 3}, 0, func(acc, v int) int { return acc + v })
	if sum != 6 {
		t.Fatalf("Reduce() = %d", sum)
	}

	page := Paginate([]int{1, 2, 3, 4, 5}, 2, 2)
	if !reflect.DeepEqual(page, []int{3, 4}) {
		t.Fatalf("Paginate() = %v", page)
	}

	sorted := SortBy([]string{"b", "a", "c"}, func(s string) string { return s })
	if !reflect.DeepEqual(sorted, []string{"a", "b", "c"}) {
		t.Fatalf("SortBy() = %v", sorted)
	}
}

func TestIntersectDifferenceMinMax(t *testing.T) {
	if !reflect.DeepEqual(Intersect([]int{1, 2, 2, 3}, []int{2, 3, 4}), []int{2, 3}) {
		t.Fatalf("Intersect() failed")
	}
	if !reflect.DeepEqual(Difference([]int{1, 2, 3}, []int{2}), []int{1, 3}) {
		t.Fatalf("Difference() failed")
	}
	if minVal, ok := MinOf([]int{3, 1, 2}); !ok || minVal != 1 {
		t.Fatalf("MinOf() = %d, %v", minVal, ok)
	}
	if maxVal, ok := MaxOf([]int{3, 1, 2}); !ok || maxVal != 3 {
		t.Fatalf("MaxOf() = %d, %v", maxVal, ok)
	}
}

func TestFindIndexDistinctByFilterMapUnion(t *testing.T) {
	arr := []int{10, 20, 30, 20}
	if idx, ok := FindIndex(arr, func(v int) bool { return v == 30 }); !ok || idx != 2 {
		t.Fatalf("FindIndex() = %d, %v", idx, ok)
	}
	if _, ok := FindIndex(arr, func(v int) bool { return v == 99 }); ok {
		t.Fatalf("FindIndex() should not find 99")
	}

	type item struct {
		ID   int
		Name string
	}
	items := []item{
		{ID: 1, Name: "a"},
		{ID: 2, Name: "b"},
		{ID: 1, Name: "c"},
		{ID: 3, Name: "d"},
	}
	distinct := DistinctBy(items, func(it item) int { return it.ID })
	if len(distinct) != 3 || distinct[0].Name != "a" || distinct[2].ID != 3 {
		t.Fatalf("DistinctBy() = %+v", distinct)
	}

	mapped := FilterMap([]int{1, 2, 3, 4}, func(v int) (int, bool) {
		if v%2 == 0 {
			return v * 10, true
		}
		return 0, false
	})
	if !reflect.DeepEqual(mapped, []int{20, 40}) {
		t.Fatalf("FilterMap() = %v", mapped)
	}

	if !reflect.DeepEqual(Union([]int{1, 2, 2}, []int{2, 3}), []int{1, 2, 3}) {
		t.Fatalf("Union() failed")
	}
}

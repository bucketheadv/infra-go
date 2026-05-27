package collection

import "testing"

func TestIsSlicePtrEmpty(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		var arr *[]int
		if !IsSlicePtrEmpty(arr) {
			t.Fatalf("expected true for nil pointer")
		}
	})

	t.Run("pointer to nil slice", func(t *testing.T) {
		var data []int
		if !IsSlicePtrEmpty(&data) {
			t.Fatalf("expected true for pointer to nil slice")
		}
	})

	t.Run("pointer to empty slice", func(t *testing.T) {
		data := []int{}
		if !IsSlicePtrEmpty(&data) {
			t.Fatalf("expected true for pointer to empty slice")
		}
	})

	t.Run("pointer to non-empty slice", func(t *testing.T) {
		data := []int{1}
		if IsSlicePtrEmpty(&data) {
			t.Fatalf("expected false for pointer to non-empty slice")
		}
	})
}

func TestIsMapPtrEmpty(t *testing.T) {
	t.Run("nil pointer", func(t *testing.T) {
		var m *map[string]int
		if !IsMapPtrEmpty(m) {
			t.Fatalf("expected true for nil pointer")
		}
	})

	t.Run("pointer to nil map", func(t *testing.T) {
		var data map[string]int
		if !IsMapPtrEmpty(&data) {
			t.Fatalf("expected true for pointer to nil map")
		}
	})

	t.Run("pointer to empty map", func(t *testing.T) {
		data := map[string]int{}
		if !IsMapPtrEmpty(&data) {
			t.Fatalf("expected true for pointer to empty map")
		}
	})

	t.Run("pointer to non-empty map", func(t *testing.T) {
		data := map[string]int{"a": 1}
		if IsMapPtrEmpty(&data) {
			t.Fatalf("expected false for pointer to non-empty map")
		}
	})
}

func TestIsSliceEmpty(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var data []int
		if !IsSliceEmpty(data) {
			t.Fatalf("expected true for nil slice")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		data := []int{}
		if !IsSliceEmpty(data) {
			t.Fatalf("expected true for empty slice")
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		data := []int{1}
		if IsSliceEmpty(data) {
			t.Fatalf("expected false for non-empty slice")
		}
	})
}

func TestIsMapEmpty(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		var data map[string]int
		if !IsMapEmpty(data) {
			t.Fatalf("expected true for nil map")
		}
	})

	t.Run("empty map", func(t *testing.T) {
		data := map[string]int{}
		if !IsMapEmpty(data) {
			t.Fatalf("expected true for empty map")
		}
	})

	t.Run("non-empty map", func(t *testing.T) {
		data := map[string]int{"a": 1}
		if IsMapEmpty(data) {
			t.Fatalf("expected false for non-empty map")
		}
	})
}

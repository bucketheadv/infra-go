package typex

import (
	"testing"
)

// TestStringToSuccess 验证各支持基础类型的成功转换。
func TestStringToSuccess(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		got, err := StringTo[string]("test")
		assertNoErr(t, err)
		assertEqual(t, got, "test")
	})

	t.Run("int family", func(t *testing.T) {
		v1, err := StringTo[int]("123")
		assertNoErr(t, err)
		assertEqual(t, v1, 123)

		v2, err := StringTo[int32]("123")
		assertNoErr(t, err)
		assertEqual(t, v2, int32(123))

		v3, err := StringTo[int64]("123")
		assertNoErr(t, err)
		assertEqual(t, v3, int64(123))
	})

	t.Run("uint family", func(t *testing.T) {
		v1, err := StringTo[uint]("123")
		assertNoErr(t, err)
		assertEqual(t, v1, uint(123))

		v2, err := StringTo[uint8]("123")
		assertNoErr(t, err)
		assertEqual(t, v2, uint8(123))

		v3, err := StringTo[uint16]("123")
		assertNoErr(t, err)
		assertEqual(t, v3, uint16(123))

		v4, err := StringTo[uint32]("123")
		assertNoErr(t, err)
		assertEqual(t, v4, uint32(123))

		v5, err := StringTo[uint64]("123")
		assertNoErr(t, err)
		assertEqual(t, v5, uint64(123))
	})

	t.Run("float family", func(t *testing.T) {
		v1, err := StringTo[float32]("1.23")
		assertNoErr(t, err)
		assertEqual(t, v1, float32(1.23))

		v2, err := StringTo[float64]("1.23")
		assertNoErr(t, err)
		assertEqual(t, v2, 1.23)
	})

	t.Run("bool", func(t *testing.T) {
		got, err := StringTo[bool]("true")
		assertNoErr(t, err)
		assertEqual(t, got, true)
	})
}

// TestStringToError 验证非法输入时会返回错误。
func TestStringToError(t *testing.T) {
	_, err := StringTo[int]("abc")
	if err == nil {
		t.Fatalf("expected int parse error")
	}

	_, err = StringTo[bool]("not-bool")
	if err == nil {
		t.Fatalf("expected bool parse error")
	}
}

// TestArrayElemToSuccess 验证切片逐项转换成功场景。
func TestArrayElemToSuccess(t *testing.T) {
	ints, err := ArrayElemTo[int]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, ints, []int{1, 2, 3})

	int8s, err := ArrayElemTo[int8]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, int8s, []int8{1, 2, 3})

	int16s, err := ArrayElemTo[int16]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, int16s, []int16{1, 2, 3})

	int32s, err := ArrayElemTo[int32]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, int32s, []int32{1, 2, 3})

	int64s, err := ArrayElemTo[int64]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, int64s, []int64{1, 2, 3})

	uints, err := ArrayElemTo[uint]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, uints, []uint{1, 2, 3})

	strs, err := ArrayElemTo[string]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, strs, []string{"1", "2", "3"})

	f32s, err := ArrayElemTo[float32]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, f32s, []float32{1, 2, 3})

	f64s, err := ArrayElemTo[float64]([]string{"1", "2", "3"})
	assertNoErr(t, err)
	assertSliceEqual(t, f64s, []float64{1, 2, 3})

	bools, err := ArrayElemTo[bool]([]string{"true", "false"})
	assertNoErr(t, err)
	assertSliceEqual(t, bools, []bool{true, false})
}

// TestArrayElemToErrorAndPartial 验证中途解析失败时会返回部分结果。
func TestArrayElemToErrorAndPartial(t *testing.T) {
	got, err := ArrayElemTo[int]([]string{"1", "bad", "3"})
	if err == nil {
		t.Fatalf("expected parse error")
	}
	assertSliceEqual(t, got, []int{1})
}

func assertNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Fatalf("value mismatch: got=%v want=%v", got, want)
	}
}

func assertSliceEqual[T comparable](t *testing.T, got, want []T) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("length mismatch: got=%d want=%d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("value mismatch at %d: got=%v want=%v", i, got[i], want[i])
		}
	}
}

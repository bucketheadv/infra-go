package typex

import "testing"

func TestDerefOrZero(t *testing.T) {
	v := 10
	if got := Deref(&v, 0); got != 10 {
		t.Fatalf("Deref() = %d", got)
	}
	if got := Deref[int](nil, 7); got != 7 {
		t.Fatalf("Deref(nil) = %d", got)
	}
	if got := OrZero(&v); got != 10 {
		t.Fatalf("OrZero() = %d", got)
	}
	if got := OrZero[int](nil); got != 0 {
		t.Fatalf("OrZero(nil) = %d", got)
	}
}

func TestCoalesce(t *testing.T) {
	if got := Coalesce(0, 0, 3, 4); got != 3 {
		t.Fatalf("Coalesce() = %d", got)
	}
	if got := Coalesce("", "hello"); got != "hello" {
		t.Fatalf("Coalesce() = %q", got)
	}
	v := 0
	if got := CoalescePtr(9, nil, &v); got != 0 {
		t.Fatalf("CoalescePtr() should keep zero value, got %d", got)
	}
	if got := CoalescePtr(9, nil, nil); got != 9 {
		t.Fatalf("CoalescePtr() default = %d", got)
	}
}

func TestMust(t *testing.T) {
	if got := Must(42, nil); got != 42 {
		t.Fatalf("Must() = %d", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatalf("Must() should panic")
		}
	}()
	Must(0, assertErr("boom"))
}

func TestFirstNonNilAndIf(t *testing.T) {
	var p1, p2 *int
	v := 1
	p2 = &v
	if got := FirstNonNil(p1, p2); got != p2 {
		t.Fatalf("FirstNonNil() unexpected")
	}
	if FirstNonNil[int](nil) != nil {
		t.Fatalf("FirstNonNil(all nil) should be nil")
	}
	if If(true, "a", "b") != "a" || If(false, "a", "b") != "b" {
		t.Fatalf("If() failed")
	}
}

func assertErr(msg string) error {
	return &testError{msg}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

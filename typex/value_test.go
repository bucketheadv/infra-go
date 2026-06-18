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

func assertErr(msg string) error {
	return &testError{msg}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

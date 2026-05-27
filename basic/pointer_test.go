package basic

import "testing"

func TestPtr(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		p := Ptr(123)
		if p == nil || *p != 123 {
			t.Fatalf("Ptr(int) failed: got=%v", p)
		}
	})

	t.Run("string", func(t *testing.T) {
		p := Ptr("hello")
		if p == nil || *p != "hello" {
			t.Fatalf("Ptr(string) failed: got=%v", p)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type user struct {
			ID   int
			Name string
		}
		p := Ptr(user{ID: 1, Name: "alice"})
		if p == nil || p.ID != 1 || p.Name != "alice" {
			t.Fatalf("Ptr(struct) failed: got=%v", p)
		}
	})
}

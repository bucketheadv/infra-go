package versionx

import "testing"

// TestCompareRules 验证版本比较核心规则。
func TestCompareRules(t *testing.T) {
	assertCompare(t, "1.2", "1.3.0", -1)
	assertCompare(t, "1.2", "1.1.99", 1)
	assertCompare(t, "1.2.30-beta", "1.2.30", 1)
	assertCompare(t, "1.2", "1.2.0", 0)
	assertCompare(t, "1.2.3", "1.2.3.0", 0)
	assertCompare(t, "1.2.3.40", "1.2.3.5", 1)
	assertCompare(t, "1.2.3.40-beta", "1.2.3.40", 1)
	assertCompare(t, "1.2.30-alpha", "1.2.30-beta", -1)
}

func TestLessGreaterEqual(t *testing.T) {
	ok, err := Less("1.2", "1.3")
	if err != nil || !ok {
		t.Fatalf("Less() = %v, %v", ok, err)
	}
	ok, err = Greater("1.3", "1.2")
	if err != nil || !ok {
		t.Fatalf("Greater() = %v, %v", ok, err)
	}
	ok, err = Equal("1.2", "1.2.0")
	if err != nil || !ok {
		t.Fatalf("Equal() = %v, %v", ok, err)
	}
	if ok, _ := Equal("1.2", "1.3"); ok {
		t.Fatalf("Equal() should be false")
	}
}

// TestParseInvalid 验证非法格式会返回错误。
func TestParseInvalid(t *testing.T) {
	cases := []string{
		"",
		"1",
		"1.2.3.4.5",
		"1..2",
		"1.2-",
		"a.b",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			if _, err := Parse(c); err == nil {
				t.Fatalf("expected parse error for %q", c)
			}
		})
	}
}

// TestMustParse 验证 MustParse 行为。
func TestMustParse(t *testing.T) {
	v := MustParse("1.2.3.40-beta")
	if v.String() != "1.2.3.40-beta" {
		t.Fatalf("unexpected version string: %s", v.String())
	}

	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic for invalid version")
		}
	}()
	_ = MustParse("bad")
}

func assertCompare(t *testing.T, left, right string, want int) {
	t.Helper()
	got, err := Compare(left, right)
	if err != nil {
		t.Fatalf("compare error: %v", err)
	}
	if got != want {
		t.Fatalf("compare mismatch: %s ? %s, got=%d want=%d", left, right, got, want)
	}
}

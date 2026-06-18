package stringx

import "testing"

func TestIsEmpty(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		cases := []struct {
			name string
			in   string
			want bool
		}{
			{name: "empty", in: "", want: true},
			{name: "space", in: " ", want: false},
			{name: "text", in: "hello", want: false},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := IsEmpty(tc.in)
				if got != tc.want {
					t.Fatalf("IsEmpty(%q) = %v, want %v", tc.in, got, tc.want)
				}
			})
		}
	})

	t.Run("string pointer", func(t *testing.T) {
		var nilPtr *string
		empty := ""
		text := "hello"
		cases := []struct {
			name string
			in   *string
			want bool
		}{
			{name: "nil", in: nilPtr, want: true},
			{name: "empty", in: &empty, want: true},
			{name: "text", in: &text, want: false},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := IsEmpty(tc.in)
				if got != tc.want {
					t.Fatalf("IsEmpty(%v) = %v, want %v", tc.in, got, tc.want)
				}
			})
		}
	})
}

func TestIsBlank(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		cases := []struct {
			name string
			in   string
			want bool
		}{
			{name: "empty", in: "", want: true},
			{name: "space", in: " ", want: true},
			{name: "tab and newline", in: "\t\n", want: true},
			{name: "text", in: "hello", want: false},
			{name: "text with spaces", in: " hello ", want: false},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := IsBlank(tc.in)
				if got != tc.want {
					t.Fatalf("IsBlank(%q) = %v, want %v", tc.in, got, tc.want)
				}
			})
		}
	})

	t.Run("string pointer", func(t *testing.T) {
		var nilPtr *string
		empty := ""
		blank := " \t"
		text := "hello"
		cases := []struct {
			name string
			in   *string
			want bool
		}{
			{name: "nil", in: nilPtr, want: true},
			{name: "empty", in: &empty, want: true},
			{name: "blank", in: &blank, want: true},
			{name: "text", in: &text, want: false},
		}

		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := IsBlank(tc.in)
				if got != tc.want {
					t.Fatalf("IsBlank(%v) = %v, want %v", tc.in, got, tc.want)
				}
			})
		}
	})
}

func TestDefaultIfBlank(t *testing.T) {
	if got := DefaultIfBlank("", "x"); got != "x" {
		t.Fatalf("got %q", got)
	}
	if got := DefaultIfBlank("  ", "x"); got != "x" {
		t.Fatalf("got %q", got)
	}
	if got := DefaultIfBlank("ok", "x"); got != "ok" {
		t.Fatalf("got %q", got)
	}

	var nilPtr *string
	if got := DefaultIfBlank(nilPtr, "x"); got != "x" {
		t.Fatalf("nil ptr got %q", got)
	}
	text := "ok"
	if got := DefaultIfBlank(&text, "x"); got != "ok" {
		t.Fatalf("ptr got %q", got)
	}
}

func TestTruncate(t *testing.T) {
	if Truncate("hello", 3) != "hel" {
		t.Fatalf("truncate ascii failed")
	}
	if Truncate("你好世界", 2) != "你好" {
		t.Fatalf("truncate rune failed")
	}
}

func TestSplitTrimJoinNonEmpty(t *testing.T) {
	got := SplitTrim(" a, b , ,c ", ",")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("SplitTrim() = %v", got)
	}

	joined := JoinNonEmpty([]string{"a", " ", "", "b"}, ",")
	if joined != "a,b" {
		t.Fatalf("JoinNonEmpty() = %q", joined)
	}
}

func TestSnakeCaseAndCamelCase(t *testing.T) {
	cases := []struct {
		in    string
		snake string
		camel string
	}{
		{"HelloWorld", "hello_world", "HelloWorld"},
		{"hello_world", "hello_world", "HelloWorld"},
		{"user-id", "user_id", "UserId"},
		{"HTTPStatus", "http_status", "HttpStatus"},
	}
	for _, tc := range cases {
		if got := SnakeCase(tc.in); got != tc.snake {
			t.Fatalf("SnakeCase(%q) = %q, want %q", tc.in, got, tc.snake)
		}
		if got := CamelCase(tc.in); got != tc.camel {
			t.Fatalf("CamelCase(%q) = %q, want %q", tc.in, got, tc.camel)
		}
	}
}

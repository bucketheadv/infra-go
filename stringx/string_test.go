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

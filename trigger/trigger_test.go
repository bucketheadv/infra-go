package trigger

import (
	"strings"
	"testing"
	"time"
)

var testLoc = time.FixedZone("UTC+08:00", 8*3600)

func TestNextTriggerTimes(t *testing.T) {
	start := time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc)
	got := mustTimes(t, "0 * * * * *", start, testLoc, 3)
	assertTimes(t, got,
		time.Date(2026, 5, 16, 10, 1, 0, 0, testLoc),
		time.Date(2026, 5, 16, 10, 2, 0, 0, testLoc),
		time.Date(2026, 5, 16, 10, 3, 0, 0, testLoc),
	)
}

func TestNextTriggerTimesInvalidSpec(t *testing.T) {
	got, err := NextTriggerTimes("invalid-spec", time.Now(), time.UTC, 2)
	t.Logf("NextTriggerTimes invalid spec result: got=%v err=%v", got, err)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err := ValidateSpec("invalid-spec"); err == nil {
		t.Fatalf("ValidateSpec() should fail for invalid spec")
	}
	if err := ValidateSpec("0 * * * * *"); err != nil {
		t.Fatalf("ValidateSpec() err=%v", err)
	}
}

func TestNextTriggerTime(t *testing.T) {
	start := time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc)
	got, err := NextTriggerTime("0 * * * * *", start, testLoc)
	if err != nil {
		t.Fatalf("NextTriggerTime() err=%v", err)
	}
	want := time.Date(2026, 5, 16, 10, 1, 0, 0, testLoc)
	if !got.Equal(want) {
		t.Fatalf("NextTriggerTime() = %v, want %v", got, want)
	}
}

func TestNextTriggerTimesNonPositiveN(t *testing.T) {
	got, err := NextTriggerTimes("0 * * * * *", time.Now(), time.UTC, 0)
	t.Logf("NextTriggerTimes non-positive n result: %s err=%v", formatTimes(got), err)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result for n<=0")
	}
}

func TestNextTriggerTimesWithYear(t *testing.T) {
	start := time.Date(2026, 12, 31, 23, 59, 50, 0, testLoc)
	got := mustTimes(t, "0 0 0 1 1 * 2027", start, testLoc, 1)
	assertTimes(t, got, time.Date(2027, 1, 1, 0, 0, 0, 0, testLoc))
}

func TestNextTriggerTimesWithInvalidYearExpr(t *testing.T) {
	got, err := NextTriggerTimes("0 * * * * * 20x7", time.Now(), time.UTC, 1)
	t.Logf("NextTriggerTimesWithInvalidYearExpr result: %s err=%v", formatTimes(got), err)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestNextTriggerTimesYearWildcard(t *testing.T) {
	start := time.Date(2026, 5, 16, 10, 0, 30, 0, testLoc)
	got := mustTimes(t, "0 * * * * * *", start, testLoc, 3)
	assertTimes(t, got,
		time.Date(2026, 5, 16, 10, 1, 0, 0, testLoc),
		time.Date(2026, 5, 16, 10, 2, 0, 0, testLoc),
		time.Date(2026, 5, 16, 10, 3, 0, 0, testLoc),
	)
}

func TestNextTriggerTimesSixFieldsSpec(t *testing.T) {
	start := time.Date(2026, 5, 16, 10, 0, 30, 0, testLoc)
	got := mustTimes(t, "0 * * * * *", start, testLoc, 3)
	// 6 位表达式默认 year="*"，这里确保行为与显式 "*" 一致。
	gotWithYearWildcard := mustTimes(t, "0 * * * * * *", start, testLoc, 3)
	assertTimes(t, got, gotWithYearWildcard...)
}

// 覆盖月/日/周维度表达式。
func TestNextTriggerTimesMonthDayWeekCoverage(t *testing.T) {
	start := time.Date(2026, 5, 16, 9, 59, 50, 0, testLoc)
	cases := []struct {
		name string
		spec string
		want time.Time
	}{
		{
			name: "day-of-month",
			spec: "0 0 10 20 * *",
			want: time.Date(2026, 5, 20, 10, 0, 0, 0, testLoc),
		},
		{
			name: "month",
			spec: "0 0 10 1 6 *",
			want: time.Date(2026, 6, 1, 10, 0, 0, 0, testLoc),
		},
		{
			name: "day-of-week",
			spec: "0 0 10 * * 6",
			want: time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc),
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := mustTimes(t, tt.spec, start, testLoc, 1)
			assertTimes(t, got, tt.want)
		})
	}
}

// 覆盖步长格式：0/1。
func TestNextTriggerTimesStepFormat(t *testing.T) {
	start := time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc)
	got := mustTimes(t, "0/1 * * * * *", start, testLoc, 3)
	assertTimes(t, got,
		time.Date(2026, 5, 16, 10, 0, 1, 0, testLoc),
		time.Date(2026, 5, 16, 10, 0, 2, 0, testLoc),
		time.Date(2026, 5, 16, 10, 0, 3, 0, testLoc),
	)
}

// 覆盖列表/区间等常见格式（秒字段）。
func TestNextTriggerTimesRangeAndListFormat(t *testing.T) {
	start := time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc)

	t.Run("range 0-1", func(t *testing.T) {
		got := mustTimes(t, "0-1 * * * * *", start, testLoc, 4)
		assertTimes(t, got,
			time.Date(2026, 5, 16, 10, 0, 1, 0, testLoc),
			time.Date(2026, 5, 16, 10, 1, 0, 0, testLoc),
			time.Date(2026, 5, 16, 10, 1, 1, 0, testLoc),
			time.Date(2026, 5, 16, 10, 2, 0, 0, testLoc),
		)
	})

	t.Run("list 0,1", func(t *testing.T) {
		got := mustTimes(t, "0,1 * * * * *", start, testLoc, 4)
		assertTimes(t, got,
			time.Date(2026, 5, 16, 10, 0, 1, 0, testLoc),
			time.Date(2026, 5, 16, 10, 1, 0, 0, testLoc),
			time.Date(2026, 5, 16, 10, 1, 1, 0, testLoc),
			time.Date(2026, 5, 16, 10, 2, 0, 0, testLoc),
		)
	})
}

// 覆盖更多常见格式：通配步长、区间步长、年份列表/区间/步长。
func TestNextTriggerTimesMoreFormats(t *testing.T) {
	t.Run("wildcard step */2 seconds", func(t *testing.T) {
		start := time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc)
		got := mustTimes(t, "*/2 * * * * *", start, testLoc, 3)
		assertTimes(t, got,
			time.Date(2026, 5, 16, 10, 0, 2, 0, testLoc),
			time.Date(2026, 5, 16, 10, 0, 4, 0, testLoc),
			time.Date(2026, 5, 16, 10, 0, 6, 0, testLoc),
		)
	})

	t.Run("range step 1-5/2 seconds", func(t *testing.T) {
		start := time.Date(2026, 5, 16, 10, 0, 0, 0, testLoc)
		got := mustTimes(t, "1-5/2 * * * * *", start, testLoc, 4)
		assertTimes(t, got,
			time.Date(2026, 5, 16, 10, 0, 1, 0, testLoc),
			time.Date(2026, 5, 16, 10, 0, 3, 0, testLoc),
			time.Date(2026, 5, 16, 10, 0, 5, 0, testLoc),
			time.Date(2026, 5, 16, 10, 1, 1, 0, testLoc),
		)
	})

	t.Run("year list 2027,2029", func(t *testing.T) {
		start := time.Date(2026, 12, 31, 23, 59, 50, 0, testLoc)
		got := mustTimes(t, "0 0 0 1 1 * 2027,2029", start, testLoc, 2)
		assertTimes(t, got,
			time.Date(2027, 1, 1, 0, 0, 0, 0, testLoc),
			time.Date(2029, 1, 1, 0, 0, 0, 0, testLoc),
		)
	})

	t.Run("year range 2027-2028", func(t *testing.T) {
		start := time.Date(2026, 12, 31, 23, 59, 50, 0, testLoc)
		got := mustTimes(t, "0 0 0 1 1 * 2027-2028", start, testLoc, 2)
		assertTimes(t, got,
			time.Date(2027, 1, 1, 0, 0, 0, 0, testLoc),
			time.Date(2028, 1, 1, 0, 0, 0, 0, testLoc),
		)
	})

	t.Run("year wildcard step */2", func(t *testing.T) {
		start := time.Date(2026, 12, 31, 23, 59, 50, 0, testLoc)
		got := mustTimes(t, "0 0 0 1 1 * */2", start, testLoc, 2)
		assertTimes(t, got,
			time.Date(2028, 1, 1, 0, 0, 0, 0, testLoc),
			time.Date(2030, 1, 1, 0, 0, 0, 0, testLoc),
		)
	})
}

func formatTimes(times []time.Time) string {
	if len(times) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(times))
	for _, t := range times {
		items = append(items, t.Format(time.RFC3339))
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func mustTimes(t *testing.T, spec string, start time.Time, loc *time.Location, n int) []time.Time {
	t.Helper()
	got, err := NextTriggerTimes(spec, start, loc, n)
	t.Logf("spec=%q start=%s result=%s err=%v", spec, start.Format(time.RFC3339), formatTimes(got), err)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	return got
}

func assertTimes(t *testing.T, got []time.Time, want ...time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("size mismatch: got=%d want=%d, got=%s", len(got), len(want), formatTimes(got))
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Fatalf("mismatch at %d: got=%s want=%s", i, got[i].Format(time.RFC3339), want[i].Format(time.RFC3339))
		}
	}
}

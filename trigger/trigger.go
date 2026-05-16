package trigger

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bucketheadv/infra-go/timezone"
	"github.com/robfig/cron/v3"
)

var defaultParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
)

// NextTriggerTimes 计算cron下次执行时间
// spec 支持两种格式：
//   - 6 位：秒 分 时 日 月 周
//   - 7 位：秒 分 时 日 月 周 年（最后一位为年份表达式）
// 年份表达式支持：*、单值（如 2027）、列表（如 2027,2029）、区间（如 2027-2030）、步长（如 */2、2027-2030/2）。
// 其他字段遵循 cron 常见写法：*、列表（0,1）、区间（0-1）、步长（0/1、*/2）等。
// startTime 表示从哪个时间点开始向后计算。
// loc 表示按哪个时区计算触发时间。
// n 返回最近的多少条时间。
func NextTriggerTimes(spec string, startTime time.Time, loc *time.Location, n int) ([]time.Time, error) {
	if n <= 0 {
		return []time.Time{}, nil
	}
	cronSpec, yearExpr, err := splitSpecAndYear(spec)
	if err != nil {
		return nil, err
	}
	s, err := defaultParser.Parse(cronSpec)
	if err != nil {
		return nil, fmt.Errorf("parse cron spec failed: %w", err)
	}
	yearMatcher, err := newYearMatcher(yearExpr)
	if err != nil {
		return nil, err
	}
	dateTime := timezone.WithZone(startTime, loc)
	result := make([]time.Time, 0, n)
	for len(result) < n {
		dateTime = s.Next(dateTime)
		if dateTime.Year() > 9999 {
			return nil, fmt.Errorf("no trigger time matched year expression %q", yearExpr)
		}
		if yearMatcher.Match(dateTime.Year()) {
			result = append(result, dateTime)
		}
	}
	return result, nil
}

func splitSpecAndYear(spec string) (cronSpec string, yearExpr string, err error) {
	fields := strings.Fields(spec)
	switch len(fields) {
	case 6:
		return strings.Join(fields, " "), "*", nil
	case 7:
		return strings.Join(fields[:6], " "), fields[6], nil
	default:
		return "", "", fmt.Errorf("parse cron spec failed: expected 6 or 7 fields, found %d", len(fields))
	}
}

type yearMatcher struct {
	matchers []yearSegmentMatcher
}

func newYearMatcher(expr string) (*yearMatcher, error) {
	e := strings.TrimSpace(expr)
	if e == "" {
		return nil, fmt.Errorf("invalid year expression: empty")
	}
	parts := strings.Split(e, ",")
	segments := make([]yearSegmentMatcher, 0, len(parts))
	for _, p := range parts {
		seg, err := parseYearSegment(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}
	return &yearMatcher{matchers: segments}, nil
}

func (m *yearMatcher) Match(year int) bool {
	for _, seg := range m.matchers {
		if seg.Match(year) {
			return true
		}
	}
	return false
}

type yearSegmentMatcher struct {
	start int
	end   int
	step  int
}

func (m yearSegmentMatcher) Match(year int) bool {
	if year < m.start || year > m.end {
		return false
	}
	return (year-m.start)%m.step == 0
}

func parseYearSegment(expr string) (yearSegmentMatcher, error) {
	if expr == "" {
		return yearSegmentMatcher{}, fmt.Errorf("invalid year expression segment: empty")
	}
	base := expr
	step := 1
	if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return yearSegmentMatcher{}, fmt.Errorf("invalid year expression segment: %q", expr)
		}
		base = parts[0]
		s, err := strconv.Atoi(parts[1])
		if err != nil || s <= 0 {
			return yearSegmentMatcher{}, fmt.Errorf("invalid year step in %q", expr)
		}
		step = s
	}

	var start, end int
	switch {
	case base == "*":
		start, end = 0, 9999
	case strings.Contains(base, "-"):
		parts := strings.Split(base, "-")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return yearSegmentMatcher{}, fmt.Errorf("invalid year range in %q", expr)
		}
		s, err := strconv.Atoi(parts[0])
		if err != nil {
			return yearSegmentMatcher{}, fmt.Errorf("invalid year value in %q", expr)
		}
		e, err := strconv.Atoi(parts[1])
		if err != nil {
			return yearSegmentMatcher{}, fmt.Errorf("invalid year value in %q", expr)
		}
		start, end = s, e
	default:
		y, err := strconv.Atoi(base)
		if err != nil {
			return yearSegmentMatcher{}, fmt.Errorf("invalid year value in %q", expr)
		}
		start, end = y, y
	}

	if start < 0 || end < 0 || start > 9999 || end > 9999 || start > end {
		return yearSegmentMatcher{}, fmt.Errorf("invalid year bounds in %q", expr)
	}
	return yearSegmentMatcher{
		start: start,
		end:   end,
		step:  step,
	}, nil
}

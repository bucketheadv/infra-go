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
	schedule, err := defaultParser.Parse(cronSpec)
	if err != nil {
		return nil, fmt.Errorf("parse cron spec failed: %w", err)
	}
	matcher, err := newYearExprMatcher(yearExpr)
	if err != nil {
		return nil, err
	}
	dateTime := timezone.WithZone(startTime, loc)
	result := make([]time.Time, 0, n)
	const maxIterations = 366 * 10000 // 防止年份过滤长期不命中时空转
	for iter := 0; len(result) < n; iter++ {
		if iter >= maxIterations {
			return nil, fmt.Errorf("no trigger time matched year expression %q within search limit", yearExpr)
		}
		dateTime = schedule.Next(dateTime)
		if dateTime.Year() > 9999 {
			return nil, fmt.Errorf("no trigger time matched year expression %q", yearExpr)
		}
		if matcher.matches(dateTime.Year()) {
			result = append(result, dateTime)
		}
	}
	return result, nil
}

// NextTriggerTime 返回 cron 下一次触发时间。
func NextTriggerTime(spec string, startTime time.Time, loc *time.Location) (time.Time, error) {
	times, err := NextTriggerTimes(spec, startTime, loc, 1)
	if err != nil {
		return time.Time{}, err
	}
	if len(times) == 0 {
		return time.Time{}, fmt.Errorf("no trigger time found")
	}
	return times[0], nil
}

// ValidateSpec 校验 cron 表达式是否合法（含可选年份字段）。
func ValidateSpec(spec string) error {
	cronSpec, yearExpr, err := splitSpecAndYear(spec)
	if err != nil {
		return err
	}
	if _, err := defaultParser.Parse(cronSpec); err != nil {
		return fmt.Errorf("parse cron spec failed: %w", err)
	}
	if _, err := newYearExprMatcher(yearExpr); err != nil {
		return err
	}
	return nil
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

type yearExprMatcher struct {
	// segments 年份表达式解析后的匹配段。
	segments []yearSegment
}

func newYearExprMatcher(expr string) (*yearExprMatcher, error) {
	trimmed := strings.TrimSpace(expr)
	if trimmed == "" {
		return nil, fmt.Errorf("invalid year expression: empty")
	}
	parts := strings.Split(trimmed, ",")
	segments := make([]yearSegment, 0, len(parts))
	for _, part := range parts {
		segment, err := parseYearSegment(strings.TrimSpace(part))
		if err != nil {
			return nil, err
		}
		segments = append(segments, segment)
	}
	return &yearExprMatcher{segments: segments}, nil
}

func (m *yearExprMatcher) matches(year int) bool {
	for _, segment := range m.segments {
		if segment.matches(year) {
			return true
		}
	}
	return false
}

type yearSegment struct {
	// start 起始年份（含）。
	start int
	// end 结束年份（含）。
	end int
	// step 步长；1 表示连续。
	step int
}

func (segment yearSegment) matches(year int) bool {
	if year < segment.start || year > segment.end {
		return false
	}
	return (year-segment.start)%segment.step == 0
}

func parseYearSegment(expr string) (yearSegment, error) {
	if expr == "" {
		return yearSegment{}, fmt.Errorf("invalid year expression segment: empty")
	}
	base := expr
	step := 1
	if strings.Contains(expr, "/") {
		parts := strings.Split(expr, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return yearSegment{}, fmt.Errorf("invalid year expression segment: %q", expr)
		}
		base = parts[0]
		stepVal, err := strconv.Atoi(parts[1])
		if err != nil || stepVal <= 0 {
			return yearSegment{}, fmt.Errorf("invalid year step in %q", expr)
		}
		step = stepVal
	}

	var start, end int
	switch {
	case base == "*":
		start, end = 0, 9999
	case strings.Contains(base, "-"):
		parts := strings.Split(base, "-")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return yearSegment{}, fmt.Errorf("invalid year range in %q", expr)
		}
		startVal, err := strconv.Atoi(parts[0])
		if err != nil {
			return yearSegment{}, fmt.Errorf("invalid year value in %q", expr)
		}
		endVal, err := strconv.Atoi(parts[1])
		if err != nil {
			return yearSegment{}, fmt.Errorf("invalid year value in %q", expr)
		}
		start, end = startVal, endVal
	default:
		yearVal, err := strconv.Atoi(base)
		if err != nil {
			return yearSegment{}, fmt.Errorf("invalid year value in %q", expr)
		}
		start, end = yearVal, yearVal
	}

	if start < 0 || end < 0 || start > 9999 || end > 9999 || start > end {
		return yearSegment{}, fmt.Errorf("invalid year bounds in %q", expr)
	}
	return yearSegment{
		start: start,
		end:   end,
		step:  step,
	}, nil
}

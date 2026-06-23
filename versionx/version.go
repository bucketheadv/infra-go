package versionx

import (
	"fmt"
	"strconv"
	"strings"
)

// Version 表示解析后的版本号。
type Version struct {
	major  int
	minor  int
	patch  int
	build  int
	suffix string
	raw    string
}

// Parse 解析版本号字符串。
func Parse(s string) (Version, error) {
	raw := strings.TrimSpace(s)
	if raw == "" {
		return Version{}, fmt.Errorf("version is empty")
	}

	mainPart := raw
	suffix := ""
	if i := strings.IndexByte(raw, '-'); i >= 0 {
		mainPart = raw[:i]
		suffix = strings.TrimSpace(raw[i+1:])
		if suffix == "" {
			return Version{}, fmt.Errorf("invalid version suffix: %q", raw)
		}
	}

	parts := strings.Split(mainPart, ".")
	if len(parts) < 2 || len(parts) > 4 {
		return Version{}, fmt.Errorf("invalid version format: %q", raw)
	}

	nums := [4]int{}
	for i, p := range parts {
		if p == "" {
			return Version{}, fmt.Errorf("invalid version segment: %q", raw)
		}
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("invalid version segment: %q", raw)
		}
		nums[i] = n
	}

	return Version{
		major:  nums[0],
		minor:  nums[1],
		patch:  nums[2],
		build:  nums[3],
		suffix: suffix,
		raw:    raw,
	}, nil
}

// MustParse 解析失败时 panic。
func MustParse(s string) Version {
	v, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}

// Less 判断 left 是否小于 right。
func Less(left, right string) (bool, error) {
	cmp, err := Compare(left, right)
	return cmp < 0, err
}

// Greater 判断 left 是否大于 right。
func Greater(left, right string) (bool, error) {
	cmp, err := Compare(left, right)
	return cmp > 0, err
}

// Equal 判断两个版本是否相等。
func Equal(left, right string) (bool, error) {
	cmp, err := Compare(left, right)
	return cmp == 0, err
}

// Compare 比较两个版本字符串。
// 返回值：-1 表示 left < right，0 表示相等，1 表示 left > right。
func Compare(left, right string) (int, error) {
	lv, err := Parse(left)
	if err != nil {
		return 0, err
	}
	rv, err := Parse(right)
	if err != nil {
		return 0, err
	}
	return lv.Compare(rv), nil
}

// Compare 比较两个 Version。
func (v Version) Compare(other Version) int {
	if c := compareInt(v.major, other.major); c != 0 {
		return c
	}
	if c := compareInt(v.minor, other.minor); c != 0 {
		return c
	}
	if c := compareInt(v.patch, other.patch); c != 0 {
		return c
	}
	if c := compareInt(v.build, other.build); c != 0 {
		return c
	}

	if v.suffix == "" && other.suffix == "" {
		return 0
	}
	if v.suffix != "" && other.suffix == "" {
		return 1
	}
	if v.suffix == "" && other.suffix != "" {
		return -1
	}
	return compareString(v.suffix, other.suffix)
}

// String 返回原始版本字符串。
func (v Version) String() string {
	return v.raw
}

func compareInt(a, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func compareString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

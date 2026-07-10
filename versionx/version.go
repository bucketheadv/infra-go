package versionx

import (
	"fmt"
	"strconv"
	"strings"
)

// Version 表示解析后的版本号。
// 比较规则接近 SemVer：数字段相同时，预发布后缀版本小于正式版。
type Version struct {
	// major 主版本号。
	major int
	// minor 次版本号。
	minor int
	// patch 修订号。
	patch int
	// build 构建号（第四段，可选）。
	build int
	// suffix 预发布后缀（如 alpha、rc.1）。
	suffix string
	// raw 原始输入字符串。
	raw string
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

// Major 返回主版本号。
func (v Version) Major() int { return v.major }

// Minor 返回次版本号。
func (v Version) Minor() int { return v.minor }

// Patch 返回补丁号。
func (v Version) Patch() int { return v.patch }

// Build 返回构建号（第四段，缺失为 0）。
func (v Version) Build() int { return v.build }

// Suffix 返回预发布后缀（不含前导 '-'）。
func (v Version) Suffix() string { return v.suffix }

// Less 判断 left 是否小于 right。
func Less(left, right string) (bool, error) {
	order, err := Compare(left, right)
	return order < 0, err
}

// Greater 判断 left 是否大于 right。
func Greater(left, right string) (bool, error) {
	order, err := Compare(left, right)
	return order > 0, err
}

// Equal 判断两个版本是否相等。
func Equal(left, right string) (bool, error) {
	order, err := Compare(left, right)
	return order == 0, err
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
// 数字段相同时：无后缀（正式版）> 有后缀（预发布）；
// 双方都有后缀时按 SemVer 预发布规则比较（点分标识，纯数字按数值，否则按字典序）。
func (v Version) Compare(other Version) int {
	if order := compareInt(v.major, other.major); order != 0 {
		return order
	}
	if order := compareInt(v.minor, other.minor); order != 0 {
		return order
	}
	if order := compareInt(v.patch, other.patch); order != 0 {
		return order
	}
	if order := compareInt(v.build, other.build); order != 0 {
		return order
	}

	if v.suffix == "" && other.suffix == "" {
		return 0
	}
	if v.suffix != "" && other.suffix == "" {
		return -1
	}
	if v.suffix == "" && other.suffix != "" {
		return 1
	}
	return comparePrerelease(v.suffix, other.suffix)
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

func comparePrerelease(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		if order := comparePrereleaseIdent(as[i], bs[i]); order != 0 {
			return order
		}
	}
	return compareInt(len(as), len(bs))
}

func comparePrereleaseIdent(a, b string) int {
	an, aNum := parsePrereleaseNum(a)
	bn, bNum := parsePrereleaseNum(b)
	if aNum && bNum {
		return compareInt(an, bn)
	}
	if aNum {
		return -1 // 纯数字标识小于非数字
	}
	if bNum {
		return 1
	}
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func parsePrereleaseNum(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

package numx

import (
	"cmp"
	"math"
)

// Clamp 将 v 限制在 [lo, hi] 闭区间内。
// 浮点 NaN 按 IEEE 比较语义原样返回（NaN 与任何值比较均为 false）。
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	if lo > hi {
		lo, hi = hi, lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// InRange 判断 v 是否在 [lo, hi] 数值闭区间内。
// 与 timex.InRange（时间区间）语义不同。浮点 NaN 恒为 false。
func InRange[T cmp.Ordered](v, lo, hi T) bool {
	if lo > hi {
		lo, hi = hi, lo
	}
	return v >= lo && v <= hi
}

// Round 将浮点数四舍五入到 places 位小数；places < 0 时原样返回。
func Round(v float64, places int) float64 {
	if places < 0 {
		return v
	}
	pow := math.Pow(10, float64(places))
	return math.Round(v*pow) / pow
}

// Percent 计算 part 占 total 的百分比（0~100）；total 为 0 时返回 0。
// part/total 含 Inf/NaN 时结果遵循浮点除法。
func Percent(part, total float64) float64 {
	if total == 0 {
		return 0
	}
	return part / total * 100
}

// ApproximatelyEqual 判断两个浮点数是否在 epsilon 误差范围内相等。
// NaN 与任何值（含自身）均不相等，符合 IEEE 754。
func ApproximatelyEqual(a, b, epsilon float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) <= epsilon
}

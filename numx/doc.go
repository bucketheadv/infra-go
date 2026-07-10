// Package numx 提供常用数值计算工具。
//
// 模块：github.com/bucketheadv/infra-go/numx
//
// 能力概览：
//   - 区间：Clamp / InRange；
//   - 浮点：Round / ApproximatelyEqual；
//   - 比例：Percent。
//
// 浮点特殊值（NaN / ±Inf）按 IEEE 754 比较语义处理，不做额外归一化：
//   - Clamp(NaN, ...) 返回 NaN；InRange(NaN, ...) 为 false；
//   - ApproximatelyEqual(NaN, NaN, ε) 为 false；
//   - Percent 在 total 为 0 时返回 0，其余遵循浮点除法结果（含 Inf/NaN）。
package numx

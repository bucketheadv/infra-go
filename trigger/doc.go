// Package trigger 提供 cron 触发时间计算能力。
//
// 模块：github.com/bucketheadv/infra-go/trigger
//
// 能力概览：
//   - 按 cron 表达式计算未来 N 次触发时间；
//   - 支持传入时区进行时间基准转换；
//   - 支持可选年份字段（第 7 段）；
//   - 统一通过 NextTriggerTimes 返回结果与错误。
//
// 说明：
//   - cron 解析启用秒字段，基础表达式为 6 段（秒 分 时 日 月 周）；
//   - 可选第 7 段作为年份约束，支持 *, 单值, 区间, 列表, 步长（如 */2、2025-2030/2）；
//   - n<=0 时返回空结果；
//   - 无效表达式会返回错误。
package trigger

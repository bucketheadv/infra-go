// Package tabular 提供 Excel/CSV 与结构体数组互转能力。
//
// 模块：github.com/bucketheadv/infra-go/tabular
//
// 能力概览：
//   - 按 sheet 名称或下标读取 Excel；
//   - 按结构体字段自动生成表头并写入 Excel/CSV；
//   - 按首行表头自动识别列并读取到结构体数组；
//   - 支持 []map[string]any / []map[string]string 读写 Excel/CSV；
//   - 支持字段标签：title / header / excel / csv / json（按优先级匹配）。
package tabular

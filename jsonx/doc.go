// Package jsonx 提供增强版 JSON 解析能力。
//
// 模块：github.com/bucketheadv/infra-go/jsonx
//
// 能力概览：
//   - 按目标 struct 字段类型自动转换数字字符串；
//   - 支持 int/uint/float 等数值类型；
//   - 支持嵌套 struct、切片、map、指针字段；
//   - 保持与 encoding/json 的使用习惯（Unmarshal）。
package jsonx

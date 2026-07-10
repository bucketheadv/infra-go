// Package versionx 提供版本号解析与比较能力。
//
// 模块：github.com/bucketheadv/infra-go/versionx
//
// 支持格式：
//   - 1.2
//   - 1.2.30
//   - 1.2.3.40
//   - 1.2.30-beta
//
// 比较规则（接近 SemVer）：
//   - 按主版本、次版本、补丁号、构建号依次比较（缺失段按 0 处理）；
//   - 当数字部分相同时，预发布版本小于正式版（如 1.2.30-beta < 1.2.30）；
//   - 双方都带后缀时按 SemVer 预发布规则比较（点分标识；纯数字按数值，否则字典序）。
package versionx

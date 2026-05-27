// Package logx 提供可复用的应用日志（YAML 配置、pattern、GORM 桥接）。
//
// 模块：github.com/bucketheadv/infra-go/logx
//
// 能力概览：
//   - 多命名 logger（loggers），未命中名称时回退 root；
//   - appender：console、rollingFile；layout 支持 text | pattern | json；
//   - pattern：占位符 %d/%date、%level/%p、%fileLine/%F、%logger/%c、%pid、%msg/%m、%n，%% 转义；
//     %clr(子模式){颜色}；fieldColors / levelColors；
//   - 调用位置：跳过本包栈帧；GORM：跳过 logx 与 gorm.io 后的业务帧；
//   - NewGormLogger：SQL 单行写入命名 logger（默认 NameGorm）；
//   - GinLogger / GinRecovery / InstallGinWriters：与 gin 默认 Logger、Recovery 及 DefaultWriter 行为兼容，输出走 logx。
package logx

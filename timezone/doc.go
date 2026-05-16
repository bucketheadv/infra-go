// Package timezone 提供应用层时区处理能力。
//
// 模块：github.com/bucketheadv/infra-go/timezone
//
// 能力概览：
//   - joda 风格命名：ForID / MustForID / UTC；
//   - 默认时区：GetDefault / SetDefault；
//   - 偏移时区：ForOffsetHours / ForOffsetHoursMinutes；
//   - joda 语义转换：WithZone / WithZoneRetainFields；
//   - 常用时区常量：IDAsiaShanghai / IDAsiaTokyo / IDEuropeLondon / IDAmericaNewYork；
//   - 时间转换：ConvertUTCToLocal / ConvertLocalToUTC；
//   - 时间解析与格式化：ParseDateTime / FormatDateTime。
package timezone

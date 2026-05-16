// Package timezone 提供应用层时区处理能力。
//
// 模块：github.com/bucketheadv/infra-go/timezone
//
// 能力概览：
//   - 内置常用 UTC 偏移时区列表：TimeZones（UTC-12:00 ~ UTC+13:00，含半小时/45 分钟时区）；
//   - 时区查询：GetTimeZone(name)；
//   - 绝对时间转换：WithZone；
//   - 保留年月日时分秒字段的时区重解释：WithZoneRetainFields。
package timezone

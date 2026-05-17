package timefmt

const (
	// DateTimeMillisISO 日期时间到毫秒，ISO 风格但不带时区。
	DateTimeMillisISO = "2006-01-02T15:04:05.000"
	// DateTimeISO 日期时间，ISO 风格但不带时区。
	DateTimeISO = "2006-01-02T15:04:05"
	// DateTimeCommon 常见日期时间格式（空格分隔）。
	DateTimeCommon = "2006-01-02 15:04:05"
	// DateTimeMillisCommon 常见日期时间格式（空格分隔，毫秒精度）。
	DateTimeMillisCommon = "2006-01-02 15:04:05.000"
	// DateOnly 仅日期格式。
	DateOnly = "2006-01-02"
	// DateOnlySlash 仅日期格式（斜杠分隔）。
	DateOnlySlash = "2006/01/02"
	// DateCompact 紧凑日期格式。
	DateCompact = "20060102"
	// DateTimeSlashCommon 常见斜杠日期时间格式。
	DateTimeSlashCommon = "2006/01/02 - 15:04:05"
	// DateTimeCompact 紧凑日期时间格式。
	DateTimeCompact = "20060102150405"
	// TimeOnly 仅时间格式。
	TimeOnly = "15:04:05"
	// TimeOnlyMillis 仅时间格式（毫秒精度）。
	TimeOnlyMillis = "15:04:05.000"
	// TimeOnlyNano 仅时间格式（纳秒精度）。
	TimeOnlyNano = "15:04:05.000000000"
)

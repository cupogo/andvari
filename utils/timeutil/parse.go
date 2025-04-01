package timeutil

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 支持的时间格式列表，按照可能性顺序排列
var formats = []string{
	// ISO 8601 格式
	"2006-01-02T15:04:05Z07:00", // ISO 8601 带时区
	"2006-01-02T15:04:05Z",      // ISO 8601 UTC
	"2006-01-02T15:04:05",       // ISO 8601 无时区
	"2006-01-02 15:04:05Z07:00", // ISO 8601 变体，空格替代T
	"2006-01-02 15:04:05",       // ISO 8601 变体，无时区

	// 日期格式
	"2006-01-02", // ISO 8601 日期
	"2006/01/02", // 亚洲常用日期格式
	"02/01/2006", // 欧洲日期格式 (DD/MM/YYYY)
	// "01/02/2006", // 美式日期格式 (MM/DD/YYYY) 部分值和欧式不容易区分
	"02.01.2006", // 欧洲日期格式 (DD.MM.YYYY)
	"20060102",   // 紧凑数字格式 (YYYYMMDD)

	// 带时间的日期格式
	"2006-01-02 15:04",    // 日期+时间(时分)
	"2006/01/02 15:04:05", // 亚洲日期+时间
	"02/01/2006 15:04:05", // 欧洲日期+时间
	// "01/02/2006 15:04:05", // 美式日期+时间 部分值和欧式不容易区分
	"02.01.2006 15:04:05", // 欧洲日期+时间

	// 12小时制格式
	"2006-01-02 3:04:05 PM", // ISO日期 + 12小时制
	"01/02/2006 3:04:05 PM", // 美式日期 + 12小时制

	// RFC 格式
	// time.RFC3339,     // RFC 3339 = ISO 8601
	time.RFC3339Nano, // RFC 3339 带纳秒
	time.RFC1123,     // RFC 1123
	time.RFC1123Z,    // RFC 1123 带时区
	time.RFC822,      // RFC 822
	time.RFC822Z,     // RFC 822 带时区
	time.RFC850,      // RFC 850

	// 其他常见格式
	"2006-01-02 15:04:05.000",       // 带毫秒
	"2006-01-02 15:04:05.000000",    // 带微秒
	"2006-01-02 15:04:05.000000000", // 带纳秒
	"20060102150405",                // 紧凑格式 (YYYYMMDDHHMMSS)
}

// ParseTime 尝试解析多种常见的时间日期格式，返回 time.Time 对象
func ParseTime(timeStr string) (time.Time, error) {
	timeStr = strings.TrimSpace(timeStr)
	if timeStr == "" {
		return time.Time{}, errors.New("empty time string")
	}

	// 尝试所有支持的格式
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			// slog.Info("match", "fmt", format, "str", timeStr)
			return t, nil
		}
	}

	// 尝试Unix时间戳格式（秒级）
	if timestamp, err := parseUnixTimestamp(timeStr); err == nil {
		return timestamp, nil
	}

	// 尝试Unix时间戳格式（毫秒级）
	if timestamp, err := parseUnixMillisTimestamp(timeStr); err == nil {
		return timestamp, nil
	}

	// 如果所有格式都失败，返回错误
	return time.Time{}, errors.New("unable to parse time string: " + timeStr)
}

// 尝试解析Unix时间戳（秒级）
func parseUnixTimestamp(timeStr string) (t time.Time, err error) {
	var sec int64
	if sec, err = strconv.ParseInt(timeStr, 10, 64); err != nil {
		return
	}

	// 验证时间戳在合理范围内（1970-2100年之间）
	if sec < 0 || sec > 4102444800 {
		return time.Time{}, errors.New("unix timestamp out of reasonable range")
	}
	// slog.Info("match unix timestamp", "sec", sec, "str", timeStr)

	return time.Unix(sec, 0), nil
}

// 尝试解析Unix时间戳（毫秒级）
func parseUnixMillisTimestamp(timeStr string) (time.Time, error) {
	var msec int64
	if _, err := fmt.Sscanf(timeStr, "%d", &msec); err != nil {
		return time.Time{}, err
	}

	// 验证是否为毫秒级时间戳（长度通常为13位）
	// 简单判断：如果数字太大，可能是毫秒级时间戳
	if msec > 4102444800000 || msec < 0 {
		return time.Time{}, errors.New("unix millisecond timestamp out of reasonable range")
	}

	if msec > 10000000000 { // 可能是毫秒级时间戳
		return time.Unix(msec/1000, (msec%1000)*1000000), nil
	}

	return time.Time{}, errors.New("not a millisecond timestamp")
}

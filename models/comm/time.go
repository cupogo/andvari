package comm

import (
	"fmt"
	"strconv"
	"strings"
)

// Time 表示一天中的某个时间点，最小单位是毫秒
type Time int32

const (
	// 一天的毫秒数
	MillisecondsPerDay = 24 * 60 * 60 * 1000
)

// String 返回一个 Time 值的字符串表示形式,如果秒值为0(不管毫秒),则不输出秒值
func (t Time) String() string {
	hour := int32(t) / 3600000
	minute := (int32(t) % 3600000) / 60000
	second := (int32(t) % 60000) / 1000
	millisecond := int32(t) % 1000

	if second == 0 {
		return fmt.Sprintf("%02d:%02d", hour, minute)
	}
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hour, minute, second, millisecond)
}

// Format 根据给定的布局格式返回一个 Time 值的字符串表示形式
func (t Time) Format(layout string) string {
	hour := int32(t) / 3600000
	minute := (int32(t) % 3600000) / 60000
	second := (int32(t) % 60000) / 1000
	millisecond := int32(t) % 1000
	return strings.NewReplacer(
		"hh", fmt.Sprintf("%02d", hour),
		"mm", fmt.Sprintf("%02d", minute),
		"ss", fmt.Sprintf("%02d", second),
		"mmm", fmt.Sprintf("%03d", millisecond),
	).Replace(layout)
}

// MarshalText 实现了 encoding.TextMarshaler 接口
func (t Time) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// UnmarshalText 实现了 encoding.TextUnmarshaler 接口
func (t *Time) UnmarshalText(data []byte) error {
	parsed, err := ParseTime(string(data))
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}

// IsZero 返回一个指示 Time 是否表示零值的布尔值
func (t Time) IsZero() bool {
	return int32(t) == 0
}

// ParseTime 解析一个表示时间的字符串并返回一个 Time 值
func ParseTime(s string) (Time, error) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, ErrInvalidTime
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, ErrInvalidHour
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, ErrInvalidMinute
	}

	var second, millisecond int
	if len(parts) == 3 {
		secondParts := strings.Split(parts[2], ".")
		second, err = strconv.Atoi(secondParts[0])
		if err != nil || second < 0 || second > 59 {
			return 0, ErrInvalidSecond
		}

		if len(secondParts) > 1 {
			millisecond, err = strconv.Atoi(secondParts[1])
			if err != nil || millisecond < 0 || millisecond >= 1000 {
				return 0, ErrInvalidMillisecond
			}
		}
	}

	totalMilliseconds := (hour*3600+minute*60+second)*1000 + millisecond
	return Time(totalMilliseconds), nil
}

// AddMilliseconds 将一个时间量(以毫秒为单位)加到 Time 值上
func (t Time) AddMilliseconds(durationMillis int64) Time {
	totalMillis := int64(t) + durationMillis
	totalMillis = totalMillis % int64(MillisecondsPerDay)
	if totalMillis < 0 {
		totalMillis += int64(MillisecondsPerDay)
	}
	return Time(totalMillis)
}

// AddSeconds 将秒数添加到 Time 值上
func (t Time) AddSeconds(seconds int) Time {
	return t.AddMilliseconds(int64(seconds) * 1000)
}

package utils

import (
	"time"
)

// TimeFormat 默认时间格式
const TimeFormat = "2006-01-02 15:04:05"

// FormatTime 格式化时间为默认格式
func FormatTime(t time.Time) string {
	return t.Format(TimeFormat)
}

// ParseTime 解析时间字符串
func ParseTime(timeStr string) (time.Time, error) {
	return time.Parse(TimeFormat, timeStr)
}

// Now 返回当前时间（默认格式）
func Now() string {
	return time.Now().Format(TimeFormat)
}

// Timestamp 返回当前时间戳
func Timestamp() int64 {
	return time.Now().Unix()
}

// FormatTimestamp 格式化时间戳为默认格式
func FormatTimestamp(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(TimeFormat)
}

// IsZeroTime 检查是否是零值时间
func IsZeroTime(t time.Time) bool {
	return t.IsZero() || t.UnixNano() == 0
}

// SafeTime 安全获取时间，如果是零值则返回当前时间
func SafeTime(t time.Time) time.Time {
	if IsZeroTime(t) {
		return time.Now()
	}
	return t
}

// FormatToDefault 将任意时间格式化为默认格式
func FormatToDefault(t time.Time) string {
	if IsZeroTime(t) {
		return ""
	}
	return FormatTime(t)
}

package utils

import "time"

// NowTime 获取当前时间字符串
func NowTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// NowTimeUnix 获取当前时间戳
func NowTimeUnix() int64 {
	return time.Now().Unix()
}

// NowTimeUnixMilli 获取当前时间戳毫秒
func NowTimeUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// NowTimeUnixNano 获取当前时间戳纳秒
func NowTimeUnixNano() int64 {
	return time.Now().UnixNano()
}

func AddTimeToString(duration time.Duration) string {
	return time.Now().Add(duration).Format("2006-01-02 15:04:05")
}

func SubTimeToString(duration time.Duration) string {
	return time.Now().Add(-duration).Format("2006-01-02 15:04:05")
}

func AddTimeToUnix(duration time.Duration) int64 {
	return time.Now().Add(duration).Unix()
}

func SubTimeToUnix(duration time.Duration) int64 {
	return time.Now().Add(-duration).Unix()
}

func AddTimeToUnixMilli(duration time.Duration) int64 {
	return time.Now().Add(duration).UnixMilli()
}

func SubTimeToUnixMilli(duration time.Duration) int64 {
	return time.Now().Add(-duration).UnixMilli()
}

func AddTimeToUnixNano(duration time.Duration) int64 {
	return time.Now().Add(duration).UnixNano()
}

func SubTimeToUnixNano(duration time.Duration) int64 {
	return time.Now().Add(-duration).UnixNano()
}

package core

import (
	"time"
)

// TimeConfig 时间配置 - 定义时间字段名称和格式
type TimeConfig struct {
	CreatedAt string `json:"created_at" yaml:"created_at"` // 创建时间字段名
	UpdatedAt string `json:"updated_at" yaml:"updated_at"` // 更新时间字段名
	DeletedAt string `json:"deleted_at" yaml:"deleted_at"` // 删除时间字段名
	Format    string `json:"format" yaml:"format"`         // 时间格式，默认 "2006-01-02 15:04:05"
}

// DefaultTimeConfig 获取默认时间配置
func DefaultTimeConfig() *TimeConfig {
	return &TimeConfig{
		CreatedAt: "created_at",
		UpdatedAt: "updated_at",
		DeletedAt: "deleted_at",
		Format:    "2006-01-02 15:04:05", // Go 的参考时间格式
	}
}

// Validate 验证时间配置
func (tc *TimeConfig) Validate() {
	if tc.CreatedAt == "" {
		tc.CreatedAt = "created_at"
	}
	if tc.UpdatedAt == "" {
		tc.UpdatedAt = "updated_at"
	}
	if tc.DeletedAt == "" {
		tc.DeletedAt = "deleted_at"
	}
	if tc.Format == "" {
		tc.Format = "2006-01-02 15:04:05"
	}
}

// GetCreatedAt 获取创建时间字段名
func (tc *TimeConfig) GetCreatedAt() string {
	tc.Validate()
	return tc.CreatedAt
}

// GetUpdatedAt 获取更新时间字段名
func (tc *TimeConfig) GetUpdatedAt() string {
	tc.Validate()
	return tc.UpdatedAt
}

// GetDeletedAt 获取删除时间字段名
func (tc *TimeConfig) GetDeletedAt() string {
	tc.Validate()
	return tc.DeletedAt
}

// GetFormat 获取时间格式
func (tc *TimeConfig) GetFormat() string {
	tc.Validate()
	return tc.Format
}

// FormatTime 格式化时间为配置的格式
func (tc *TimeConfig) FormatTime(t time.Time) string {
	tc.Validate()
	return t.Format(tc.Format)
}

// ParseTime 解析时间字符串
func (tc *TimeConfig) ParseTime(timeStr string) (time.Time, error) {
	tc.Validate()
	return time.Parse(tc.Format, timeStr)
}

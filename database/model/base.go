package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// DateTime 自定义时间类型
type DateTime struct {
	time.Time
}

// MarshalJSON 实现 JSON 序列化
func (t DateTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format("2006-01-02 15:04:05"))
}

// UnmarshalJSON 实现 JSON 反序列化
func (t *DateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" || s == "null" {
		t.Time = time.Time{}
		return nil
	}

	parsed, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

// Value 实现 driver.Valuer
func (t DateTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan 实现 sql.Scanner
func (t *DateTime) Scan(v interface{}) error {
	if v == nil {
		t.Time = time.Time{}
		return nil
	}

	switch value := v.(type) {
	case time.Time:
		t.Time = value
		return nil
	case string:
		parsed, err := time.Parse("2006-01-02 15:04:05", value)
		if err != nil {
			return err
		}
		t.Time = parsed
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// 基础模型
type BaseModel struct {
	ID         uint     `gorm:"column:id;primaryKey;autoIncrement;common:主键ID" json:"id"`
	CreateTime DateTime `gorm:"autoCreateTime;column:create_time;type:datetime" json:"create_time"`
	UpdateTime DateTime `gorm:"autoUpdateTime;column:update_time;type:datetime" json:"update_time"`
}

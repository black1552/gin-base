package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime 自定义时间类型，支持 null 值
type LocalTime struct {
	time.Time
}

func (t *LocalTime) MarshalJSON() ([]byte, error) {
	if t == nil || t.IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf("\"%s\"", t.Format("2006-01-02 15:04:05"))
	return []byte(formatted), nil
}

func (t *LocalTime) Value() (driver.Value, error) {
	if t == nil || t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

func (t *LocalTime) Scan(v interface{}) error {
	if v == nil {
		t.Time = time.Time{}
		return nil
	}

	switch value := v.(type) {
	case time.Time:
		t.Time = value
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

// BaseModel 基础模型，包含自定义的创建时间和更新时间
type BaseModel struct {
	Id         int        `json:"id" gorm:"column:id;primaryKey;autoIncrement;common:主键ID"` // 主键 ID
	CreateTime *LocalTime `json:"create_time" gorm:"column:create_time;type:datetime;autoCreateTime;common:创建时间"`
	UpdateTime *LocalTime `json:"update_time" gorm:"column:update_time;type:datetime;autoUpdateTime;common:更新时间"`
}

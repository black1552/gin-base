package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型，包含自定义的创建时间和更新时间
type BaseModel struct {
	Id         int       `json:"id" gorm:"column:id;primaryKey;autoIncrement;common:主键ID"` // 主键 ID
	CreateTime time.Time `json:"create_time" gorm:"column:create_time;type:datetime;autoCreateTime;common:创建时间"`
	UpdateTime time.Time `json:"update_time" gorm:"column:update_time;type:datetime;autoUpdateTime;common:更新时间"`
}

// BeforeCreate 创建前自动设置时间
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	b.CreateTime = now
	b.UpdateTime = now
	return nil
}

// BeforeUpdate 更新前自动更新时间
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdateTime = time.Now()
	return nil
}

func (b *BaseModel) AfterFind(tx *gorm.DB) error {
	b.UpdateTime = time.Now()
	return nil
}

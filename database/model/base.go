package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	Id         int    `gorm:"column:id;primaryKey;autoIncrement;common:主键ID" json:"id"`
	CreateTime string `gorm:"autoCreateTime;column:create_time;type:datetime" json:"create_time"`
	UpdateTime string `gorm:"autoUpdateTime;column:update_time;type:datetime" json:"update_time"`
}

func (b BaseModel) BeforeCreate(tx *gorm.DB) error {
	b.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	b.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

func (b BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	return nil
}

package base

import (
	"github.com/gogf/gf/v2/os/gtime"
	"gorm.io/gorm"
)

type IdModel struct {
	Id int `json:"id" gorm:"column:id;type:int(11);common:id"`
}
type TimeModel struct {
	CreateTime string `json:"create_time" gorm:"column:create_time;type:varchar(255);common:创建时间"`
	UpdateTime string `json:"update_time" gorm:"column:update_time;type:varchar(255);common:更新时间"`
}

func (tm *TimeModel) BeforeCreate(scope *gorm.DB) error {
	scope.Set("create_time", gtime.Datetime())
	scope.Set("update_time", gtime.Datetime())
	return nil
}

func (tm *TimeModel) BeforeUpdate(scope *gorm.DB) error {
	scope.Set("update_time", gtime.Datetime())
	return nil
}

//func (tm *TimeModel) AfterFind(scope *gorm.DB) error {
//	tm.CreateTime = gtime.New(tm.CreateTime).String()
//	tm.UpdateTime = gtime.New(tm.UpdateTime).String()
//	return nil
//}

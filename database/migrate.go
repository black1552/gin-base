package database

import (
	"git.magicany.cc/black1552/gin-base/log"
	"github.com/gogf/gf/v2/frame/g"
)

func SetAutoMigrate(models ...interface{}) {
	if g.IsNil(Db) {
		log.Error("数据库连接失败")
		return
	}
	Db = Db.Set("gorm:table_options", "ENGINE=InnoDB")
	err := Db.AutoMigrate(models...)
	if err != nil {
		log.Error("数据库迁移失败", err)
	}
}
func RenameColumn(dst interface{}, name, newName string) {
	if Db.Migrator().HasColumn(dst, name) {
		err := Db.Migrator().RenameColumn(dst, name, newName)
		if err != nil {
			log.Error("数据库修改字段失败", err)
			return
		}
	} else {
		log.Info("数据库字段不存在", name)
	}
}

// DropColumn
// 删除字段
// 例：DropColumn(&User{}, "Sex")
func DropColumn(dst interface{}, name string) {
	if Db.Migrator().HasColumn(dst, name) {
		err := Db.Migrator().DropColumn(dst, name)
		if err != nil {
			log.Error("数据库删除字段失败", err)
			return
		}
	} else {
		log.Info("数据库字段不存在", name)
	}
}

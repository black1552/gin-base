package database

import (
	"database/sql"

	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/log"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	Type  gorm.Dialector
	Db    *gorm.DB
	err   error
	sqlDb *sql.DB
	dns   = config.GetConfigValue("database.dns", gfile.Join(gfile.Pwd(), "db", "database.db"))
)

func init() {
	if g.IsEmpty(dns) {
		log.Error("gormDns未配置", "请检查配置文件")
		return
	}
	switch config.GetConfigValue("database.type", "sqlite").String() {
	case "mysql":
		log.Info("使用mysql数据库")
		mysqlInit()
	case "sqlite":
		log.Info("使用sqlite数据库")
		sqliteInit()
	}
	Db, err = gorm.Open(Type, &gorm.Config{})
	if err != nil {
		log.Error("数据库连接失败: ", err)
		return
	}
	sqlDb, err = Db.DB()
	if err != nil {
		log.Error("获取sqlDb失败", err)
		return
	}
	if err = sqlDb.Ping(); err != nil {
		log.Error("数据库未正常连接", err)
		return
	}
}

func mysqlInit() {
	Type = mysql.New(mysql.Config{
		DSN:                       dns.String(),
		DefaultStringSize:         255,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	})
}

func sqliteInit() {
	Type = sqlite.Open(dns.String())
}

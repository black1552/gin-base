package main

import (
	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/database"
	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/server"
)

// TIP <p>To run your response, right-click the response and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	// 初始化数据库连接
	initDatabase()

	g := server.New()
	cf, err := config.Unmarshal[config.BaseConfig]()
	if err != nil {
		log.Error("转换配置失败", err)
	}
	log.Info("启动服务:", cf.Server.Addr)
	server.Run(g)
}

// initDatabase 初始化数据库连接
func initDatabase() {
	// 检查是否配置了数据库
	dbConfig := config.GetConfigValue("database")
	if dbConfig.IsEmpty() {
		log.Warn("未配置数据库，跳过数据库初始化")
		return
	}

	log.Info("正在初始化数据库连接...")
	defer func() {
		if err := recover(); err != nil {
			log.Error("数据库初始化失败:", err)
			return
		}
	}()

	// 调用 database 包的 Database 函数进行初始化
	// 这会自动读取配置、创建连接池、设置日志器等
	db := database.Database()
	if db == nil {
		log.Error("数据库实例创建失败")
		return
	}

	log.Info("数据库连接初始化成功")
}

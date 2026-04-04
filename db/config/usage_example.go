package config

import (
	"fmt"
	"git.magicany.cc/black1552/gin-base/db/core"
	"git.magicany.cc/black1552/gin-base/db/driver"
)

// 示例：在应用程序中使用纯自研驱动管理
func ExampleUsage() {
	// 1. 首先导入你选择的第三方数据库驱动
	// 注意：这些驱动需要在 main 包或启动包中导入
	/*
		import (
			_ "github.com/mattn/go-sqlite3"      // SQLite 驱动
			_ "github.com/go-sql-driver/mysql"   // MySQL 驱动
			_ "github.com/lib/pq"               // PostgreSQL 驱动
			_ "github.com/denisenkom/go-mssqldb" // SQL Server 驱动
		)
	*/

	// 2. 获取驱动管理器并注册你选择的驱动
	manager := driver.GetDefaultManager()

	// 注册 SQLite 驱动
	sqliteDriver := driver.NewGenericDriver("sqlite3")
	manager.Register("sqlite3", sqliteDriver)
	manager.Register("sqlite", sqliteDriver) // 别名

	// 注册 MySQL 驱动
	mysqlDriver := driver.NewGenericDriver("mysql")
	manager.Register("mysql", mysqlDriver)

	// 注册 PostgreSQL 驱动
	postgresDriver := driver.NewGenericDriver("postgres")
	manager.Register("postgres", postgresDriver)

	// 3. 加载配置文件
	configFile, err := LoadFromFile("config.yaml")
	if err != nil {
		fmt.Printf("加载配置失败：%v\n", err)
		return
	}

	// 4. 验证驱动是否已注册
	err = manager.RegisterDriverByConfig(configFile.Database.Type)
	if err != nil {
		fmt.Printf("驱动未注册：%v\n", err)
		return
	}

	// 5. 使用配置创建数据库连接
	dbConfig := &core.Config{
		DriverName:      configFile.Database.GetDriverName(),
		DataSource:      configFile.Database.BuildDSN(),
		Debug:           true,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600000000000, // 1小时
	}

	// 6. 创建数据库实例
	db, err := core.NewDatabase(dbConfig)
	if err != nil {
		fmt.Printf("创建数据库连接失败：%v\n", err)
		return
	}

	fmt.Printf("成功连接到 %s 数据库\n", configFile.Database.Type)

	// 现在可以使用 db 进行数据库操作
	_ = db
}

// AdvancedExample 高级使用示例
func AdvancedExample() {
	manager := driver.GetDefaultManager()

	// 根据环境变量或配置动态注册驱动
	databaseType := "sqlite3" // 从配置获取

	// 注册对应驱动
	genericDriver := driver.NewGenericDriver(databaseType)
	manager.Register(databaseType, genericDriver)

	// 验证驱动
	err := manager.RegisterDriverByConfig(databaseType)
	if err != nil {
		fmt.Printf("驱动注册问题：%v\n", err)
		return
	}

	// 现在可以安全地打开数据库连接
	db, err := manager.Open(databaseType, "./example.db")
	if err != nil {
		fmt.Printf("打开数据库失败：%v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("数据库连接成功")
}

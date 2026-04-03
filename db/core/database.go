package core

import (
	"fmt"
	"os"
	"path/filepath"

	"git.magicany.cc/black1552/gin-base/db/driver"
)

// defaultDatabase 全局默认数据库连接实例
var defaultDatabase *Database

// NewDatabase 创建数据库连接 - 初始化数据库连接和相关组件
func NewDatabase(config *Config) (*Database, error) {
	db := &Database{
		config:     config,
		debug:      config.Debug,
		driverName: config.DriverName,
	}

	// 初始化时间配置
	if config.TimeConfig == nil {
		db.timeConfig = DefaultTimeConfig()
	} else {
		db.timeConfig = config.TimeConfig
		db.timeConfig.Validate()
	}

	// 获取驱动管理器
	dm := driver.GetDefaultManager()

	// 打开数据库连接
	sqlDB, err := dm.Open(config.DriverName, config.DataSource)
	if err != nil {
		return nil, fmt.Errorf("打开数据库失败：%w", err)
	}

	db.db = sqlDB

	// 配置连接池参数
	if config.MaxIdleConns > 0 {
		db.db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.MaxOpenConns > 0 {
		db.db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}

	// 测试数据库连接
	if err := db.db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败：%w", err)
	}

	// 初始化组件
	db.mapper = NewFieldMapper()
	db.migrator = NewMigrator(db)

	if config.Debug {
		fmt.Println("[Magic-ORM] 数据库连接成功")
	}

	// 设置为全局默认实例
	defaultDatabase = db

	return db, nil
}

// AutoConnect 自动查找配置文件并创建数据库连接
// 会在当前目录及上级目录中查找 config.yaml, config.toml, config.ini, config.json 等文件
func AutoConnect(debug bool) (*Database, error) {
	// 自动查找配置文件
	configPath, err := findConfigFile("")
	if err != nil {
		return nil, fmt.Errorf("查找配置文件失败：%w", err)
	}

	// 从文件加载配置（使用 config 包）
	return loadAndConnect(configPath, debug)
}

// Connect 从配置文件创建数据库连接（向后兼容）
// Deprecated: 使用 AutoConnect 代替
func Connect(configPath string, debug bool) (*Database, error) {
	return loadAndConnect(configPath, debug)
}

// findConfigFile 在项目目录下自动查找配置文件
// 支持 yaml, yml, toml, ini, json 等格式
// 只在当前目录查找，不越级查找
func findConfigFile(searchDir string) (string, error) {
	// 配置文件名优先级列表
	configNames := []string{
		"config.yaml", "config.yml",
		"config.toml",
		"config.ini",
		"config.json",
		".config.yaml", ".config.yml",
		".config.toml",
		".config.ini",
		".config.json",
	}

	// 如果未指定搜索目录，使用当前目录
	if searchDir == "" {
		var err error
		searchDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("获取当前目录失败：%w", err)
		}
	}

	// 只在当前目录下查找，不向上查找
	for _, name := range configNames {
		filePath := filepath.Join(searchDir, name)
		if _, err := os.Stat(filePath); err == nil {
			return filePath, nil
		}
	}

	return "", fmt.Errorf("未找到配置文件（支持 yaml, yml, toml, ini, json 格式）")
}

// loadAndConnect 从配置文件加载并创建数据库连接
func loadAndConnect(configPath string, debug bool) (*Database, error) {
	// 这里需要调用 config 包的 LoadFromFile
	// 为了避免循环依赖，我们直接在 core 包中实现简单的 YAML 解析
	// 或者通过接口传递配置

	// 简单方案：返回错误，提示使用 config 包
	return nil, fmt.Errorf("请使用 config.AutoConnect() 方法")
}

// GetDefaultDatabase 获取全局默认数据库实例
func GetDefaultDatabase() *Database {
	return defaultDatabase
}

// SetDefaultDatabase 设置全局默认数据库实例
func SetDefaultDatabase(db *Database) {
	defaultDatabase = db
}

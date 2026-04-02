package config

import (
	"fmt"

	"git.magicany.cc/black1552/gin-base/db/core"
	"gopkg.in/yaml.v3"
)

// NewDatabaseFromConfig 从配置文件创建数据库连接（已废弃，请使用 AutoConnect）
// Deprecated: 使用 AutoConnect 代替
func NewDatabaseFromConfig(configPath string, debug bool) (*core.Database, error) {
	return autoConnectWithConfig(configPath, debug)
}

// AutoConnect 自动查找配置文件并创建数据库连接
// 会在当前目录及上级目录中查找 config.yaml, config.toml, config.ini, config.json 等文件
func AutoConnect(debug bool) (*core.Database, error) {
	// 自动查找配置文件
	configPath, err := FindConfigFile("")
	if err != nil {
		return nil, fmt.Errorf("查找配置文件失败：%w", err)
	}

	return autoConnectWithConfig(configPath, debug)
}

// AutoConnectWithDir 在指定目录自动查找配置文件并创建数据库连接
func AutoConnectWithDir(dir string, debug bool) (*core.Database, error) {
	configPath, err := FindConfigFile(dir)
	if err != nil {
		return nil, fmt.Errorf("查找配置文件失败：%w", err)
	}

	return autoConnectWithConfig(configPath, debug)
}

// autoConnectWithConfig 根据配置文件创建数据库连接（内部使用）
func autoConnectWithConfig(configPath string, debug bool) (*core.Database, error) {
	// 从文件加载配置
	configFile, err := LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置失败：%w", err)
	}

	// 构建核心数据库配置
	dbConfig := &core.Config{
		DriverName:      configFile.Database.GetDriverName(),
		DataSource:      configFile.Database.BuildDSN(),
		Debug:           debug,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600000000000,            // 1 小时
		TimeConfig:      core.DefaultTimeConfig(), // 使用默认时间配置
	}

	// 创建数据库连接
	db, err := core.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败：%w", err)
	}

	return db, nil
}

// NewDatabaseFromYAML 从 YAML 内容创建数据库连接
func NewDatabaseFromYAML(yamlContent []byte, debug bool) (*core.Database, error) {
	var configFile Config
	if err := yaml.Unmarshal(yamlContent, &configFile); err != nil {
		return nil, fmt.Errorf("解析 YAML 失败：%w", err)
	}

	if err := configFile.Validate(); err != nil {
		return nil, fmt.Errorf("验证配置失败：%w", err)
	}

	// 构建核心数据库配置
	dbConfig := &core.Config{
		DriverName:      configFile.Database.GetDriverName(),
		DataSource:      configFile.Database.BuildDSN(),
		Debug:           debug,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 3600000000000, // 1 小时
		TimeConfig:      core.DefaultTimeConfig(),
	}

	// 创建数据库连接
	db, err := core.NewDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接失败：%w", err)
	}

	return db, nil
}

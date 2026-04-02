package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DatabaseConfig 数据库配置结构 - 对应配置文件中的 database 部分
type DatabaseConfig struct {
	Host     string `yaml:"host"` // 数据库地址
	Port     string `yaml:"port"` // 数据库端口
	User     string `yaml:"user"` // 用户名
	Password string `yaml:"pass"` // 密码
	Name     string `yaml:"name"` // 数据库名称
	Type     string `yaml:"type"` // 数据库类型（mysql, sqlite, postgres 等）
}

// Config 完整配置文件结构
type Config struct {
	Database DatabaseConfig `yaml:"database"` // 数据库配置
}

// LoadFromFile 从 YAML 文件加载配置
func LoadFromFile(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败：%w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%w", err)
	}

	// 验证必填字段
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Database.Type == "" {
		return fmt.Errorf("数据库类型不能为空")
	}

	if c.Database.Type == "sqlite" {
		// SQLite 只需要 Name（作为文件路径）
		if c.Database.Name == "" {
			return fmt.Errorf("SQLite 数据库名称不能为空")
		}
	} else {
		// 其他数据库需要所有字段
		if c.Database.Host == "" {
			return fmt.Errorf("数据库地址不能为空")
		}
		if c.Database.Port == "" {
			return fmt.Errorf("数据库端口不能为空")
		}
		if c.Database.User == "" {
			return fmt.Errorf("数据库用户名不能为空")
		}
		if c.Database.Password == "" {
			return fmt.Errorf("数据库密码不能为空")
		}
		if c.Database.Name == "" {
			return fmt.Errorf("数据库名称不能为空")
		}
	}

	return nil
}

// BuildDSN 根据配置构建数据源连接字符串（DSN）
func (c *DatabaseConfig) BuildDSN() string {
	switch c.Type {
	case "mysql":
		return c.buildMySQLDSN()
	case "postgres":
		return c.buildPostgresDSN()
	case "sqlite":
		return c.buildSQLiteDSN()
	default:
		// 默认返回原始配置
		return ""
	}
}

// buildMySQLDSN 构建 MySQL DSN
func (c *DatabaseConfig) buildMySQLDSN() string {
	// 格式：user:pass@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
	)
	return dsn
}

// buildPostgresDSN 构建 PostgreSQL DSN
func (c *DatabaseConfig) buildPostgresDSN() string {
	// 格式：host=localhost port=5432 user=user password=password dbname=db sslmode=disable
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.Name,
	)
	return dsn
}

// buildSQLiteDSN 构建 SQLite DSN
func (c *DatabaseConfig) buildSQLiteDSN() string {
	// SQLite 直接使用文件名作为 DSN
	return c.Name
}

// GetDriverName 获取驱动名称
func (c *DatabaseConfig) GetDriverName() string {
	return c.Type
}

// FindConfigFile 在项目目录下自动查找配置文件
// 支持 yaml, yml, toml, ini, json 等格式
// 只在当前目录查找，不越级查找
func FindConfigFile(searchDir string) (string, error) {
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

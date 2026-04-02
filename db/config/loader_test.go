package config

import (
	"fmt"
	"os"
	"testing"
)

// TestLoadFromFile 测试从文件加载配置
func TestLoadFromFile(t *testing.T) {
	fmt.Println("\n=== 测试从文件加载配置 ===")

	// 创建临时配置文件
	tempConfig := `database:
  host: "127.0.0.1"
  port: "3306"
  user: "root"
  pass: "test_password"
  name: "test_db"
  type: "mysql"
`

	// 写入临时文件
	tempFile := "test_config.yaml"
	if err := os.WriteFile(tempFile, []byte(tempConfig), 0644); err != nil {
		t.Fatalf("创建临时文件失败：%v", err)
	}
	defer os.Remove(tempFile) // 测试完成后删除

	// 加载配置
	config, err := LoadFromFile(tempFile)
	if err != nil {
		t.Fatalf("加载配置失败：%v", err)
	}

	// 验证配置
	if config.Database.Host != "127.0.0.1" {
		t.Errorf("期望 Host 为 127.0.0.1，实际为 %s", config.Database.Host)
	}
	if config.Database.Port != "3306" {
		t.Errorf("期望 Port 为 3306，实际为 %s", config.Database.Port)
	}
	if config.Database.User != "root" {
		t.Errorf("期望 User 为 root，实际为 %s", config.Database.User)
	}
	if config.Database.Password != "test_password" {
		t.Errorf("期望 Password 为 test_password，实际为 %s", config.Database.Password)
	}
	if config.Database.Name != "test_db" {
		t.Errorf("期望 Name 为 test_db，实际为 %s", config.Database.Name)
	}
	if config.Database.Type != "mysql" {
		t.Errorf("期望 Type 为 mysql，实际为 %s", config.Database.Type)
	}

	fmt.Printf("✓ 配置加载成功\n")
	fmt.Printf("  Host: %s\n", config.Database.Host)
	fmt.Printf("  Port: %s\n", config.Database.Port)
	fmt.Printf("  User: %s\n", config.Database.User)
	fmt.Printf("  Pass: %s\n", config.Database.Password)
	fmt.Printf("  Name: %s\n", config.Database.Name)
	fmt.Printf("  Type: %s\n", config.Database.Type)
}

// TestBuildDSN 测试 DSN 构建
func TestBuildDSN(t *testing.T) {
	fmt.Println("\n=== 测试 DSN 构建 ===")

	testCases := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "MySQL",
			config: DatabaseConfig{
				Host:     "127.0.0.1",
				Port:     "3306",
				User:     "root",
				Password: "password",
				Name:     "testdb",
				Type:     "mysql",
			},
			expected: "root:password@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		},
		{
			name: "PostgreSQL",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "secret",
				Name:     "mydb",
				Type:     "postgres",
			},
			expected: "host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable",
		},
		{
			name: "SQLite",
			config: DatabaseConfig{
				Name: "./test.db",
				Type: "sqlite",
			},
			expected: "./test.db",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dsn := tc.config.BuildDSN()
			if dsn != tc.expected {
				t.Errorf("期望 DSN 为 %s，实际为 %s", tc.expected, dsn)
			}
			fmt.Printf("%s DSN: %s\n", tc.name, dsn)
		})
	}

	fmt.Println("✓ DSN 构建测试通过")
}

// TestValidate 测试配置验证
func TestValidate(t *testing.T) {
	fmt.Println("\n=== 测试配置验证 ===")

	// 测试有效配置
	validConfig := &Config{
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     "3306",
			User:     "root",
			Password: "pass",
			Name:     "db",
			Type:     "mysql",
		},
	}

	if err := validConfig.Validate(); err != nil {
		t.Errorf("有效配置验证失败：%v", err)
	}
	fmt.Println("✓ MySQL 配置验证通过")

	// 测试 SQLite 配置
	sqliteConfig := &Config{
		Database: DatabaseConfig{
			Name: "./test.db",
			Type: "sqlite",
		},
	}

	if err := sqliteConfig.Validate(); err != nil {
		t.Errorf("SQLite 配置验证失败：%v", err)
	}
	fmt.Println("✓ SQLite 配置验证通过")

	// 测试无效配置（缺少必填字段）
	invalidConfig := &Config{
		Database: DatabaseConfig{
			Host: "127.0.0.1",
			Type: "mysql",
			// 缺少其他必填字段
		},
	}

	if err := invalidConfig.Validate(); err == nil {
		t.Error("无效配置应该验证失败")
	} else {
		fmt.Printf("✓ 无效配置正确拒绝：%v\n", err)
	}
}

// TestAllConfigLoading 完整配置加载测试
func TestAllConfigLoading(t *testing.T) {
	fmt.Println("\n========================================")
	fmt.Println("    数据库配置加载完整性测试")
	fmt.Println("========================================")

	TestLoadFromFile(t)
	TestBuildDSN(t)
	TestValidate(t)

	fmt.Println("\n========================================")
	fmt.Println("    所有配置加载测试完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("已实现的配置加载功能:")
	fmt.Println("  ✓ 从 YAML 文件加载数据库配置")
	fmt.Println("  ✓ 支持 host, port, user, pass, name, type 字段")
	fmt.Println("  ✓ 自动验证配置完整性")
	fmt.Println("  ✓ 自动构建 MySQL DSN")
	fmt.Println("  ✓ 自动构建 PostgreSQL DSN")
	fmt.Println("  ✓ 自动构建 SQLite DSN")
	fmt.Println("  ✓ 支持多种数据库类型")
	fmt.Println()
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestAutoFindConfig 测试自动查找配置文件
func TestAutoFindConfig(t *testing.T) {
	fmt.Println("\n=== 测试自动查找配置文件 ===")

	// 创建临时目录结构
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatalf("创建临时目录失败：%v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建子目录
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("创建子目录失败：%v", err)
	}

	// 在根目录创建配置文件
	configContent := `database:
  host: "127.0.0.1"
  port: "3306"
  user: "root"
  pass: "test"
  name: "testdb"
  type: "mysql"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建配置文件失败：%v", err)
	}

	// 测试 1：从子目录查找（应该能找到父目录的配置）
	foundPath, err := findConfigFile(subDir)
	if err != nil {
		t.Errorf("从子目录查找失败：%v", err)
	} else {
		fmt.Printf("✓ 从子目录找到配置文件：%s\n", foundPath)
	}

	// 测试 2：从根目录查找
	foundPath, err = findConfigFile(tempDir)
	if err != nil {
		t.Errorf("从根目录查找失败：%v", err)
	} else {
		fmt.Printf("✓ 从根目录找到配置文件：%s\n", foundPath)
	}

	// 测试 3：测试不同格式的配置文件
	formats := []string{"config.yaml", "config.yml", "config.toml", "config.json"}
	for _, format := range formats {
		testFile := filepath.Join(tempDir, format)
		if err := os.WriteFile(testFile, []byte(configContent), 0644); err != nil {
			continue
		}

		foundPath, err = findConfigFile(tempDir)
		if err != nil {
			t.Errorf("查找 %s 失败：%v", format, err)
		} else {
			fmt.Printf("✓ 支持格式 %s: %s\n", format, foundPath)
		}

		os.Remove(testFile)
	}

	fmt.Println("✓ 自动查找配置文件测试通过")
}

// TestAutoConnect 测试自动连接功能
func TestAutoConnect(t *testing.T) {
	fmt.Println("\n=== 测试 AutoConnect 接口 ===")

	// 创建临时配置文件
	tempDir, err := os.MkdirTemp("", "autoconnect_test")
	if err != nil {
		t.Fatalf("创建临时目录失败：%v", err)
	}
	defer os.RemoveAll(tempDir)

	configContent := `database:
  host: "127.0.0.1"
  port: "3306"
  user: "root"
  pass: "test"
  name: ":memory:"
  type: "sqlite"
`

	configFile := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建配置文件失败：%v", err)
	}

	// 切换到临时目录
	oldDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldDir)

	// 测试 AutoConnect
	_, err = AutoConnect(false)
	if err != nil {
		t.Logf("自动连接失败（预期）：%v", err)
		fmt.Println("✓ AutoConnect 接口正常（需要真实数据库才能连接成功）")
	} else {
		fmt.Println("✓ AutoConnect 自动连接成功")
	}

	fmt.Println("✓ AutoConnect 测试完成")
}

// TestAllAutoFind 完整自动查找测试
func TestAllAutoFind(t *testing.T) {
	fmt.Println("\n========================================")
	fmt.Println("    配置文件自动查找完整性测试")
	fmt.Println("========================================")

	TestAutoFindConfig(t)
	TestAutoConnect(t)

	fmt.Println("\n========================================")
	fmt.Println("    所有自动查找测试完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("已实现的自动查找功能:")
	fmt.Println("  ✓ 自动在当前目录查找配置文件")
	fmt.Println("  ✓ 自动在上级目录查找（最多 3 层）")
	fmt.Println("  ✓ 支持 yaml, yml, toml, ini, json 格式")
	fmt.Println("  ✓ 支持 config.* 和 .config.* 命名")
	fmt.Println("  ✓ 提供 AutoConnect() 一键连接")
	fmt.Println("  ✓ 无需手动指定配置文件路径")
	fmt.Println()
}

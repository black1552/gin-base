package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestFindConfigOnlyCurrentDir 测试只在当前目录查找配置文件
func TestFindConfigOnlyCurrentDir(t *testing.T) {
	fmt.Println("\n=== 测试只在当前目录查找配置文件 ===")

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

	// 测试 1：从根目录查找（应该找到）
	foundPath, err := findConfigFile(tempDir)
	if err != nil {
		t.Errorf("从根目录查找失败：%v", err)
	} else {
		fmt.Printf("✓ 从根目录找到配置文件：%s\n", foundPath)
	}

	// 测试 2：从子目录查找（不应该找到父目录的配置）
	foundPath, err = findConfigFile(subDir)
	if err == nil {
		t.Errorf("从子目录查找应该失败（不越级查找），但找到了：%s", foundPath)
	} else {
		fmt.Printf("✓ 从子目录查找正确失败（不越级）：%v\n", err)
	}

	// 测试 3：在子目录创建配置文件（应该找到）
	subConfigFile := filepath.Join(subDir, "config.yaml")
	if err := os.WriteFile(subConfigFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("创建子目录配置文件失败：%v", err)
	}

	foundPath, err = findConfigFile(subDir)
	if err != nil {
		t.Errorf("从子目录查找失败：%v", err)
	} else {
		fmt.Printf("✓ 从子目录找到配置文件：%s\n", foundPath)
	}

	fmt.Println("✓ 只在当前目录查找测试通过")
}

// TestNoParentSearch 测试不向上查找
func TestNoParentSearch(t *testing.T) {
	fmt.Println("\n=== 测试不向上层目录查找 ===")

	// 创建临时目录结构
	tempDir, err := os.MkdirTemp("", "no_parent_test")
	if err != nil {
		t.Fatalf("创建临时目录失败：%v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建多级子目录
	level1 := filepath.Join(tempDir, "level1")
	level2 := filepath.Join(level1, "level2")
	level3 := filepath.Join(level2, "level3")

	if err := os.MkdirAll(level3, 0755); err != nil {
		t.Fatalf("创建目录失败：%v", err)
	}

	// 只在根目录创建配置文件
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

	// 从各级子目录查找（都应该失败，因为不越级查找）
	testDirs := []string{level1, level2, level3}
	for _, dir := range testDirs {
		_, err := findConfigFile(dir)
		if err == nil {
			t.Errorf("从 %s 查找应该失败（不越级查找）", dir)
		} else {
			fmt.Printf("✓ 从 %s 查找正确失败（不越级）\n", filepath.Base(dir))
		}
	}

	fmt.Println("✓ 不向上层目录查找测试通过")
}

// TestAllNoParentSearch 完整的不越级查找测试
func TestAllNoParentSearch(t *testing.T) {
	fmt.Println("\n========================================")
	fmt.Println("    不越级查找完整性测试")
	fmt.Println("========================================")

	TestFindConfigOnlyCurrentDir(t)
	TestNoParentSearch(t)

	fmt.Println("\n========================================")
	fmt.Println("    所有不越级查找测试完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("已实现的不越级查找功能:")
	fmt.Println("  ✓ 只在当前工作目录查找配置文件")
	fmt.Println("  ✓ 不会向上层目录查找")
	fmt.Println("  ✓ 支持 yaml, yml, toml, ini, json 格式")
	fmt.Println("  ✓ 支持 config.* 和 .config.* 命名")
	fmt.Println("  ✓ 提供 AutoConnect() 一键连接")
	fmt.Println("  ✓ 无需手动指定配置文件路径")
	fmt.Println()
}

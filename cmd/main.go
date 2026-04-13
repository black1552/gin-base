package main

import (
	"context"
	"fmt"
	"os"

	"git.magicany.cc/black1552/gin-base/cmd/gendao"
	"git.magicany.cc/black1552/gin-base/config"
	_ "git.magicany.cc/black1552/gin-base/database/drivers"
)

func main() {
	ctx := context.Background()

	// 加载配置
	cfg := config.GetAllConfig()
	if cfg == nil {
		fmt.Println("❌ 错误: 配置为空")
		os.Exit(1)
	}

	// 检查数据库配置
	dbConfigMap, ok := cfg["database"].(map[string]any)
	if !ok || len(dbConfigMap) == 0 {
		fmt.Println("❌ 错误: 未找到数据库配置")
		os.Exit(1)
	}

	// 获取默认数据库配置
	defaultDbConfig, ok := dbConfigMap["default"].(map[string]any)
	if !ok {
		fmt.Println("❌ 错误: 未找到 default 数据库配置")
		os.Exit(1)
	}

	// 提取配置值
	host := getStringValue(defaultDbConfig, "host", "127.0.0.1")
	port := getStringValue(defaultDbConfig, "port", "3306")
	name := getStringValue(defaultDbConfig, "name", "test")
	user := getStringValue(defaultDbConfig, "user", "root")
	pass := getStringValue(defaultDbConfig, "pass", "")
	dbType := getStringValue(defaultDbConfig, "type", "mysql")
	link := getStringValue(defaultDbConfig, "link", "")

	fmt.Println("=== Gin-Base DAO 代码生成工具 ===")
	fmt.Printf("🔧 类型: %s\n", dbType)

	// 构建数据库连接字符串
	var connectionInfo string
	if link == "" {
		// 如果没有配置 link，则根据数据库类型构建
		switch dbType {
		case "sqlite":
			// SQLite 使用配置文件中的 link 或默认路径
			connectionInfo = fmt.Sprintf("📁 数据库文件: %s", link)
		case "mysql", "mariadb":
			link = fmt.Sprintf("mysql:%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
				user, pass, host, port, name,
			)
			connectionInfo = fmt.Sprintf("📊 数据库: %s\n🌐 主机: %s:%s", name, host, port)
		case "pgsql", "postgresql":
			link = fmt.Sprintf("pgsql:%s:%s@tcp(%s:%s)/%s?sslmode=disable",
				user, pass, host, port, name,
			)
			connectionInfo = fmt.Sprintf("📊 数据库: %s\n🌐 主机: %s:%s", name, host, port)
		case "mssql":
			link = fmt.Sprintf("mssql:%s:%s@tcp(%s:%s)/%s",
				user, pass, host, port, name,
			)
			connectionInfo = fmt.Sprintf("📊 数据库: %s\n🌐 主机: %s:%s", name, host, port)
		case "oracle":
			link = fmt.Sprintf("oracle:%s:%s@%s:%s/%s",
				user, pass, host, port, name,
			)
			connectionInfo = fmt.Sprintf("📊 数据库: %s\n🌐 主机: %s:%s", name, host, port)
		case "clickhouse":
			link = fmt.Sprintf("clickhouse:%s:%s@tcp(%s:%s)/%s",
				user, pass, host, port, name,
			)
			connectionInfo = fmt.Sprintf("📊 数据库: %s\n🌐 主机: %s:%s", name, host, port)
		default:
			fmt.Printf("⚠️  警告: 未知的数据库类型 %s，尝试使用配置的 link\n", dbType)
			if link == "" {
				fmt.Println("❌ 错误: 未配置数据库连接信息")
				os.Exit(1)
			}
			connectionInfo = "🔗 使用自定义连接"
		}
	} else {
		// 使用配置文件中直接提供的 link
		connectionInfo = fmt.Sprintf("🔗 连接: %s", link)
	}

	fmt.Println(connectionInfo)
	fmt.Println()

	// 准备表名参数
	tablesArg := ""
	if len(os.Args) > 1 {
		tablesArg = joinStrings(os.Args[1:], ",")
		fmt.Printf("📋 指定生成表: %s\n\n", tablesArg)
	} else {
		fmt.Println("💡 提示: 可以通过命令行参数指定要生成的表")
		fmt.Println("   例如: gin-dao-gen users orders\n")
	}

	// 调用 gendao 的 Dao 函数生成代码
	genDao := gendao.CGenDao{}
	input := gendao.CGenDaoInput{
		Link:           link,
		Tables:         tablesArg,
		Path:           "./",
		DaoPath:        "dao",
		DoPath:         "model/do",
		EntityPath:     "model/entity",
		TablePath:      "model/table",
		Group:          "default",
		JsonCase:       "CamelLower",
		DescriptionTag: true,
		GenTable:       true,
	}

	fmt.Println("🔨 开始生成代码...")
	_, err := genDao.Dao(ctx, input)
	if err != nil {
		fmt.Printf("\n❌ 代码生成失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n🎉 代码生成完成！")
	fmt.Println("📁 生成的文件位于:")
	fmt.Println("   - ./dao/")
	fmt.Println("   - ./model/do/")
	fmt.Println("   - ./model/entity/")
	fmt.Println("   - ./model/table/")
}

// 辅助函数：从 map 中获取字符串值
func getStringValue(m map[string]any, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// 辅助函数：连接字符串数组
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

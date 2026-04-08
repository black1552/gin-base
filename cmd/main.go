package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/database"
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
	dbType := getStringValue(defaultDbConfig, "type", "mysql")

	fmt.Println("=== Gin-Base DAO 代码生成工具 ===")
	fmt.Printf("📊 数据库: %s\n", name)
	fmt.Printf("🔧 类型: %s\n", dbType)
	fmt.Printf("🌐 主机: %s:%s\n\n", host, port)

	// 初始化数据库连接
	err := initDatabaseFromMap(dbConfigMap)
	if err != nil {
		fmt.Printf("❌ 数据库初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 获取数据库实例（使用 default 组）
	db := database.Database("default")

	// 调试信息：打印当前使用的数据库名称
	config := db.GetConfig()
	if config != nil {
		fmt.Printf("🔍 调试: 当前数据库名 = %s, 类型 = %s\n", config.Name, config.Type)
	}

	// 获取所有表
	tables, err := db.Tables(ctx)
	if err != nil {
		fmt.Printf("❌ 获取表列表失败: %v\n", err)
		os.Exit(1)
	}

	// 如果没找到表，尝试直接查询验证
	if len(tables) == 0 {
		fmt.Println("⚠️  警告: 未找到任何表，尝试直接查询...")
		result, err := db.Query(ctx, "SHOW TABLES")
		if err != nil {
			fmt.Printf("❌ 直接查询失败: %v\n", err)
		} else {
			fmt.Printf("🔍 直接查询结果: %d 行\n", len(result))
			for _, row := range result {
				fmt.Printf("   - %v\n", row)
			}
		}
	}

	fmt.Printf("📋 找到 %d 个表:\n", len(tables))
	for i, table := range tables {
		fmt.Printf("   %d. %s\n", i+1, table)
	}
	fmt.Println()

	// 询问用户要生成的表
	var selectedTables []string
	if len(os.Args) > 1 {
		// 从命令行参数获取表名
		selectedTables = os.Args[1:]
	} else {
		// 默认生成所有表
		selectedTables = tables
		fmt.Println("💡 提示: 可以通过命令行参数指定要生成的表")
		fmt.Println("   例如: gin-dao-gen users orders")
	}

	// 创建输出目录
	dirs := []string{
		"./internal/dao",
		"./internal/model/do",
		"./internal/model/entity",
		"./internal/model/table",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("❌ 创建目录失败 %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// 为每个表生成代码
	for _, tableName := range selectedTables {
		fmt.Printf("\n🔨 正在生成表 [%s] 的代码...\n", tableName)

		// 获取表字段信息
		fields, err := db.TableFields(ctx, tableName)
		if err != nil {
			fmt.Printf("   ⚠️  获取表字段失败: %v\n", err)
			return // 使用 return 而不是 continue
		}

		// 调试：打印字段信息
		fmt.Printf("   🔍 调试: 找到 %d 个字段\n", len(fields))
		for name, field := range fields {
			fmt.Printf("      - %s: %s (%s)\n", name, field.Type, field.Comment)
		}

		// 生成 Entity
		entityName := tableNameToStructName(tableName)
		generateEntity(tableName, entityName, fields)

		// 生成 DO
		generateDO(tableName, entityName, fields)

		// 生成 DAO
		generateDAO(tableName, entityName)

		// 生成 Table
		generateTable(tableName, entityName, fields)

		fmt.Printf("   ✅ 完成\n")
	}

	fmt.Println("\n🎉 代码生成完成！")
	fmt.Println("📁 生成的文件位于:")
	fmt.Println("   - ./internal/dao/")
	fmt.Println("   - ./internal/model/do/")
	fmt.Println("   - ./internal/model/entity/")
	fmt.Println("   - ./internal/model/table/")
}

// 从 Map 初始化数据库
func initDatabaseFromMap(dbConfigMap map[string]any) error {
	for name, nodeConfig := range dbConfigMap {
		nodeMap, ok := nodeConfig.(map[string]any)
		if !ok {
			continue
		}

		configNode := database.ConfigNode{
			Host:    getStringValue(nodeMap, "host", "127.0.0.1"),
			Port:    getStringValue(nodeMap, "port", "3306"),
			User:    getStringValue(nodeMap, "user", "root"),
			Pass:    getStringValue(nodeMap, "pass", ""),
			Name:    getStringValue(nodeMap, "name", ""),
			Type:    getStringValue(nodeMap, "type", "mysql"),
			Role:    database.Role(getStringValue(nodeMap, "role", "master")),
			Debug:   getBoolValue(nodeMap, "debug", false),
			Prefix:  getStringValue(nodeMap, "prefix", ""),
			Charset: getStringValue(nodeMap, "charset", "utf8"),
		}

		if err := database.AddConfigNode(name, configNode); err != nil {
			return fmt.Errorf("add config node %s failed: %w", name, err)
		}
	}

	return nil
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

// 辅助函数：从 map 中获取布尔值
func getBoolValue(m map[string]any, key string, defaultValue bool) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// 表名转结构体名
func tableNameToStructName(tableName string) string {
	parts := strings.Split(tableName, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			result.WriteString(part[1:])
		}
	}
	return result.String()
}

// 生成 Entity 文件
func generateEntity(tableName, entityName string, fields map[string]*database.TableField) {
	filename := fmt.Sprintf("./internal/model/entity/%s.go", tableName)

	var content strings.Builder
	content.WriteString("package entity\n\n")
	content.WriteString("// Auto-generated by gin-base gen dao tool\n\n")
	content.WriteString(fmt.Sprintf("// %s represents the entity for table %s\n", entityName, tableName))
	content.WriteString(fmt.Sprintf("type %s struct {\n", entityName))

	for _, field := range fields {
		fieldName := fieldNameToStructName(field.Name)
		goType := dbTypeToGoType(field.Type)
		jsonTag := field.Name

		content.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\" description:\"%s\"`\n",
			fieldName, goType, jsonTag, field.Comment))
	}

	content.WriteString("}\n")

	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		fmt.Printf("      ❌ 写入文件失败: %v\n", err)
	} else {
		fmt.Printf("      📄 生成 Entity: %s\n", filename)
	}
}

// 生成 DO 文件
func generateDO(tableName, entityName string, fields map[string]*database.TableField) {
	filename := fmt.Sprintf("./internal/model/do/%s.go", tableName)

	var content strings.Builder
	content.WriteString("package do\n\n")
	content.WriteString("import \"github.com/gogf/gf/v2/frame/g\"\n\n")
	content.WriteString("// Auto-generated by gin-base gen dao tool\n\n")
	content.WriteString(fmt.Sprintf("// %s represents the data object for table %s\n", entityName, tableName))
	content.WriteString(fmt.Sprintf("type %s struct {\n\tg.Meta `orm:\"table:%s, do:true\"`\n\n", entityName, tableName))

	for _, field := range fields {
		fieldName := fieldNameToStructName(field.Name)
		goType := dbTypeToGoType(field.Type)

		content.WriteString(fmt.Sprintf("\t%s *%s `json:\"%s,omitempty\"`\n",
			fieldName, goType, field.Name))
	}

	content.WriteString("}\n")

	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		fmt.Printf("      ❌ 写入文件失败: %v\n", err)
	} else {
		fmt.Printf("      📄 生成 DO: %s\n", filename)
	}
}

// 生成 DAO 文件
func generateDAO(tableName, entityName string) {
	filename := fmt.Sprintf("./internal/dao/%s.go", tableName)
	lowerName := strings.ToLower(entityName[:1]) + entityName[1:]

	var content strings.Builder
	content.WriteString("package dao\n\n")
	content.WriteString("import (\n")
	content.WriteString("\t\"git.magicany.cc/black1552/gin-base/database\"\n")
	content.WriteString(fmt.Sprintf("\t\"git.magicany.cc/black1552/gin-base/internal/model/entity\"\n"))
	content.WriteString(")\n\n")
	content.WriteString("// Auto-generated by gin-base gen dao tool\n\n")
	content.WriteString(fmt.Sprintf("// %s is the DAO for table %s\n", entityName, tableName))
	content.WriteString(fmt.Sprintf("var %s = New%s()\n\n", lowerName, entityName))
	content.WriteString(fmt.Sprintf("// %s creates and returns a new DAO instance\n", entityName))
	content.WriteString(fmt.Sprintf("func New%s() *%sDao {\n", entityName, entityName))
	content.WriteString(fmt.Sprintf("\treturn &%sDao{\n", entityName))
	content.WriteString("\t\ttable: \"" + tableName + "\",\n")
	content.WriteString("\t}\n")
	content.WriteString("}\n\n")
	content.WriteString(fmt.Sprintf("// %sDao is the data access object for %s\n", entityName, tableName))
	content.WriteString(fmt.Sprintf("type %sDao struct {\n", entityName))
	content.WriteString("\ttable string\n")
	content.WriteString("}\n\n")
	content.WriteString("// Table returns the table name\n")
	content.WriteString(fmt.Sprintf("func (d *%sDao) Table() string {\n", entityName))
	content.WriteString("\treturn d.table\n")
	content.WriteString("}\n\n")
	content.WriteString("// DB returns the database instance\n")
	content.WriteString("func (d *DB) DB() database.DB {\n")
	content.WriteString("\treturn database.Database()\n")
	content.WriteString("}\n")

	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		fmt.Printf("      ❌ 写入文件失败: %v\n", err)
	} else {
		fmt.Printf("      📄 生成 DAO: %s\n", filename)
	}
}

// 生成 Table 文件
func generateTable(tableName, entityName string, fields map[string]*database.TableField) {
	filename := fmt.Sprintf("./internal/model/table/%s.go", tableName)

	var content strings.Builder
	content.WriteString("package table\n\n")
	content.WriteString("// Auto-generated by gin-base gen dao tool\n\n")
	content.WriteString("const (\n")
	content.WriteString(fmt.Sprintf("\t// %s is the table name\n", entityName))
	content.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", entityName, tableName))
	content.WriteString(")\n\n")
	content.WriteString("// Columns defines all columns of the table\n")
	content.WriteString("var Columns = struct {\n")

	for _, field := range fields {
		fieldName := fieldNameToConstName(field.Name)
		content.WriteString(fmt.Sprintf("\t%s string\n", fieldName))
	}
	content.WriteString("}{\n")

	for _, field := range fields {
		fieldName := fieldNameToConstName(field.Name)
		content.WriteString(fmt.Sprintf("\t%s: \"%s\",\n", fieldName, field.Name))
	}
	content.WriteString("}\n")

	if err := os.WriteFile(filename, []byte(content.String()), 0644); err != nil {
		fmt.Printf("      ❌ 写入文件失败: %v\n", err)
	} else {
		fmt.Printf("      📄 生成 Table: %s\n", filename)
	}
}

// 字段名转结构体字段名
func fieldNameToStructName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			result.WriteString(part[1:])
		}
	}
	return result.String()
}

// 字段名转常量名
func fieldNameToConstName(fieldName string) string {
	parts := strings.Split(fieldName, "_")
	var result strings.Builder
	for i, part := range parts {
		if i > 0 {
			result.WriteString("_")
		}
		result.WriteString(strings.ToUpper(part))
	}
	return result.String()
}

// 数据库类型转 Go 类型
func dbTypeToGoType(dbType string) string {
	dbType = strings.ToLower(dbType)

	switch {
	case strings.Contains(dbType, "int"):
		if strings.Contains(dbType, "bigint") {
			return "int64"
		}
		return "int"
	case strings.Contains(dbType, "float"), strings.Contains(dbType, "double"), strings.Contains(dbType, "decimal"):
		return "float64"
	case strings.Contains(dbType, "bool"):
		return "bool"
	case strings.Contains(dbType, "datetime"), strings.Contains(dbType, "timestamp"):
		return "*gtime.Time"
	case strings.Contains(dbType, "date"):
		return "*gtime.Time"
	case strings.Contains(dbType, "text"), strings.Contains(dbType, "char"), strings.Contains(dbType, "varchar"):
		return "string"
	case strings.Contains(dbType, "blob"), strings.Contains(dbType, "binary"):
		return "[]byte"
	default:
		return "string"
	}
}

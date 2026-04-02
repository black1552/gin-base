package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"git.magicany.cc/black1552/gin-base/db/config"
	"git.magicany.cc/black1552/gin-base/db/generator"
	"git.magicany.cc/black1552/gin-base/db/introspector"
)

// 设置 Windows 控制台编码为 UTF-8
func init() {
	// 在 Windows 上设置控制台输出代码页为 UTF-8 (65001)
	// 这样可以避免中文乱码问题
}

const version = "1.0.0"

func main() {
	// 定义命令行参数
	versionFlag := flag.Bool("version", false, "显示版本号")
	vFlag := flag.Bool("v", false, "显示版本号（简写）")
	helpFlag := flag.Bool("help", false, "显示帮助信息")
	hFlag := flag.Bool("h", false, "显示帮助信息（简写）")
	outputDir := flag.String("o", "./model", "输出目录")
	allFlag := flag.Bool("all", false, "生成所有预设的表（user, product, order）")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Magic-ORM 代码生成器 - 快速生成 Model 和 DAO 代码

用法:
  gendb [选项] <表名> [列定义...]
  gendb [选项] -all

选项:
`)
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, `
示例:
  # 生成 user 表代码（自动推断常用列）
  gendb user
  
  # 指定输出目录
  gendb -o ./models user product
  
  # 自定义列定义
  gendb user id:int64:primary username:string email:string created_at:time.Time
  
  # 批量生成多个表
  gendb user product order
  
  # 生成所有预设的表（user, product, order）
  gendb -all

列定义格式:
  字段名：类型 [:primary] [:nullable]
  
支持的类型:
  int64, string, time.Time, bool, float64, int

更多信息:
  https://github.com/your-repo/magic-orm
`)
	}

	flag.Parse()

	// 检查版本参数
	if *versionFlag || *vFlag {
		fmt.Printf("Magic-ORM Code Generator v%s\n", version)
		return
	}

	// 检查帮助参数
	if *helpFlag || *hFlag {
		flag.Usage()
		return
	}

	// 检查 -all 参数
	if *allFlag {
		generateAllTablesFromDB(*outputDir)
		return
	}

	// 获取参数
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "错误：请指定至少一个表名")
		fmt.Fprintln(os.Stderr, "使用 'gendb -h' 查看帮助")
		fmt.Fprintln(os.Stderr, "或者使用 'gendb -all' 生成所有预设表")
		os.Exit(1)
	}

	tableNames := args

	// 创建代码生成器
	cg := generator.NewCodeGenerator(*outputDir)

	fmt.Printf("[Magic-ORM Code Generator v%s]\n", version)
	fmt.Printf("[Output Directory: %s]\n", *outputDir)
	fmt.Println()

	// 处理每个表
	for _, tableName := range tableNames {
		// 跳过看起来像列定义的参数
		if strings.Contains(tableName, ":") {
			continue
		}

		fmt.Printf("[Generating table '%s'...]\n", tableName)

		// 解析列定义（如果有提供）
		columns := parseColumns(tableNames, tableName)

		// 如果没有自定义列定义，使用默认列
		if len(columns) == 0 {
			columns = getDefaultColumns(tableName)
		}

		// 生成代码
		err := cg.GenerateAll(tableName, columns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Error] Generation failed: %v\n", err)
			continue
		}

		fmt.Printf("[Success] Generated %s.go and %s_dao.go\n", tableName, tableName)
	}

	fmt.Println()
	fmt.Println("[Complete] Code generation finished!")
	fmt.Printf("[Location] Files are in: %s directory\n", *outputDir)
}

// parseColumns 解析列定义
func parseColumns(args []string, currentTable string) []generator.ColumnInfo {
	// 查找当前表的列定义
	found := false
	columnDefs := []string{}

	for i, arg := range args {
		if arg == currentTable && !found {
			found = true
			// 收集后续的列定义
			for j := i + 1; j < len(args); j++ {
				if strings.Contains(args[j], ":") {
					columnDefs = append(columnDefs, args[j])
				} else {
					break // 遇到下一个表名
				}
			}
			break
		}
	}

	if len(columnDefs) == 0 {
		return nil
	}

	columns := []generator.ColumnInfo{}
	for _, def := range columnDefs {
		parts := strings.Split(def, ":")
		if len(parts) < 2 {
			continue
		}

		colName := parts[0]
		fieldType := parts[1]
		isPrimary := false
		isNullable := false

		// 检查修饰符
		for i := 2; i < len(parts); i++ {
			switch strings.ToLower(parts[i]) {
			case "primary":
				isPrimary = true
			case "nullable":
				isNullable = true
			}
		}

		// 转换为 Go 字段名（驼峰）
		fieldName := toCamelCase(colName)

		// 映射类型
		goType := mapType(fieldType)

		columns = append(columns, generator.ColumnInfo{
			ColumnName: colName,
			FieldName:  fieldName,
			FieldType:  goType,
			JSONName:   colName,
			IsPrimary:  isPrimary,
			IsNullable: isNullable,
		})
	}

	return columns
}

// getDefaultColumns 获取默认的列定义（根据表名推断）
func getDefaultColumns(tableName string) []generator.ColumnInfo {
	columns := []generator.ColumnInfo{
		{
			ColumnName: "id",
			FieldName:  "ID",
			FieldType:  "int64",
			JSONName:   "id",
			IsPrimary:  true,
		},
	}

	// 根据表名添加常见字段
	switch tableName {
	case "user", "users":
		columns = append(columns,
			generator.ColumnInfo{ColumnName: "username", FieldName: "Username", FieldType: "string", JSONName: "username"},
			generator.ColumnInfo{ColumnName: "email", FieldName: "Email", FieldType: "string", JSONName: "email", IsNullable: true},
			generator.ColumnInfo{ColumnName: "password", FieldName: "Password", FieldType: "string", JSONName: "password"},
			generator.ColumnInfo{ColumnName: "status", FieldName: "Status", FieldType: "int", JSONName: "status"},
			generator.ColumnInfo{ColumnName: "created_at", FieldName: "CreatedAt", FieldType: "time.Time", JSONName: "created_at"},
			generator.ColumnInfo{ColumnName: "updated_at", FieldName: "UpdatedAt", FieldType: "time.Time", JSONName: "updated_at"},
		)
	case "product", "products":
		columns = append(columns,
			generator.ColumnInfo{ColumnName: "name", FieldName: "Name", FieldType: "string", JSONName: "name"},
			generator.ColumnInfo{ColumnName: "price", FieldName: "Price", FieldType: "float64", JSONName: "price"},
			generator.ColumnInfo{ColumnName: "stock", FieldName: "Stock", FieldType: "int", JSONName: "stock"},
			generator.ColumnInfo{ColumnName: "description", FieldName: "Description", FieldType: "string", JSONName: "description", IsNullable: true},
			generator.ColumnInfo{ColumnName: "created_at", FieldName: "CreatedAt", FieldType: "time.Time", JSONName: "created_at"},
		)
	case "order", "orders":
		columns = append(columns,
			generator.ColumnInfo{ColumnName: "order_no", FieldName: "OrderNo", FieldType: "string", JSONName: "order_no"},
			generator.ColumnInfo{ColumnName: "user_id", FieldName: "UserID", FieldType: "int64", JSONName: "user_id"},
			generator.ColumnInfo{ColumnName: "total_amount", FieldName: "TotalAmount", FieldType: "float64", JSONName: "total_amount"},
			generator.ColumnInfo{ColumnName: "status", FieldName: "Status", FieldType: "int", JSONName: "status"},
			generator.ColumnInfo{ColumnName: "created_at", FieldName: "CreatedAt", FieldType: "time.Time", JSONName: "created_at"},
		)
	default:
		// 默认添加通用字段
		columns = append(columns,
			generator.ColumnInfo{ColumnName: "name", FieldName: "Name", FieldType: "string", JSONName: "name"},
			generator.ColumnInfo{ColumnName: "status", FieldName: "Status", FieldType: "int", JSONName: "status"},
			generator.ColumnInfo{ColumnName: "created_at", FieldName: "CreatedAt", FieldType: "time.Time", JSONName: "created_at"},
			generator.ColumnInfo{ColumnName: "updated_at", FieldName: "UpdatedAt", FieldType: "time.Time", JSONName: "updated_at"},
		)
	}

	return columns
}

// mapType 将类型字符串映射到 Go 类型
func mapType(typeStr string) string {
	typeMap := map[string]string{
		"int":       "int64",
		"integer":   "int64",
		"bigint":    "int64",
		"string":    "string",
		"text":      "string",
		"varchar":   "string",
		"time.Time": "time.Time",
		"time":      "time.Time",
		"datetime":  "time.Time",
		"bool":      "bool",
		"boolean":   "bool",
		"float":     "float64",
		"float64":   "float64",
		"double":    "float64",
		"decimal":   "string",
	}

	if goType, ok := typeMap[strings.ToLower(typeStr)]; ok {
		return goType
	}
	return "string" // 默认返回 string
}

// toCamelCase 转换为驼峰命名
func toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	result := ""

	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + part[1:]
		}
	}

	return result
}

// generateAllTablesFromDB 从数据库读取所有表并生成代码
func generateAllTablesFromDB(outputDir string) {
	fmt.Printf("[Magic-ORM Code Generator v%s]\n", version)
	fmt.Println()

	// 1. 加载配置文件
	fmt.Println("[Step 1] Loading configuration file...")
	cfg, err := loadDatabaseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Failed to load config: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[Info] Database type: %s\n", cfg.Type)
	fmt.Printf("[Info] Database name: %s\n", cfg.Name)
	fmt.Println()

	// 2. 连接数据库并获取所有表
	fmt.Println("[Step 2] Connecting to database and fetching table structure...")
	intro, err := introspector.NewIntrospector(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer intro.Close()

	tableNames, err := intro.GetTableNames()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Error] Failed to get table names: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[Info] Found %d tables\n", len(tableNames))
	fmt.Println()

	// 3. 创建代码生成器
	cg := generator.NewCodeGenerator(outputDir)

	// 4. 为每个表生成代码
	for _, tableName := range tableNames {
		fmt.Printf("[Generating] Table '%s'...\n", tableName)

		// 获取表详细信息
		tableInfo, err := intro.GetTableInfo(tableName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Error] Failed to get table info: %v\n", err)
			continue
		}

		// 转换为 generator.ColumnInfo
		columns := make([]generator.ColumnInfo, len(tableInfo.Columns))
		for i, col := range tableInfo.Columns {
			columns[i] = generator.ColumnInfo{
				ColumnName: col.ColumnName,
				FieldName:  col.FieldName,
				FieldType:  col.GoType,
				JSONName:   col.JSONName,
				IsPrimary:  col.IsPrimary,
				IsNullable: col.IsNullable,
			}
		}

		// 生成代码
		err = cg.GenerateAll(tableName, columns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[Error] Generation failed: %v\n", err)
			continue
		}

		fmt.Printf("[Success] Generated %s.go and %s_dao.go\n", tableName, tableName)
	}

	fmt.Println()
	fmt.Println("[Complete] Code generation finished!")
	fmt.Printf("[Location] Files are in: %s directory\n", outputDir)
}

// loadDatabaseConfig 加载数据库配置
func loadDatabaseConfig() (*config.DatabaseConfig, error) {
	// 自动查找配置文件
	configPath, err := config.FindConfigFile("")
	if err != nil {
		return nil, fmt.Errorf("查找配置文件失败：%w", err)
	}

	fmt.Printf("[Info] Using config file: %s\n", configPath)

	// 从文件加载配置
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败：%w", err)
	}

	return &cfg.Database, nil
}

// generateAllTables 生成所有预设的表
func generateAllTables(outputDir string) {
	fmt.Printf("🚀 Magic-ORM 代码生成器 v%s\n", version)
	fmt.Printf("📁 输出目录：%s\n", outputDir)
	fmt.Println()

	// 预设的所有表
	presetTables := []string{"user", "product", "order"}

	// 创建代码生成器
	cg := generator.NewCodeGenerator(outputDir)

	// 处理每个表
	for _, tableName := range presetTables {
		fmt.Printf("📝 生成表 '%s' 的代码...\n", tableName)

		// 使用默认列定义
		columns := getDefaultColumns(tableName)

		// 生成代码
		err := cg.GenerateAll(tableName, columns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 生成失败：%v\n", err)
			continue
		}

		fmt.Printf("✅ 成功生成 %s.go 和 %s_dao.go\n", tableName, tableName)
	}

	fmt.Println()
	fmt.Println("✨ 代码生成完成！")
	fmt.Printf("📂 生成的文件在：%s 目录下\n", outputDir)
}

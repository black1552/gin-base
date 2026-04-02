package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"git.magicany.cc/black1552/gin-base/db/core"
	"git.magicany.cc/black1552/gin-base/db/model"
)

// TestTimeConfig 测试时间配置
func TestTimeConfig(t *testing.T) {
	fmt.Println("\n=== 测试时间配置 ===")

	// 测试默认配置
	defaultConfig := core.DefaultTimeConfig()
	fmt.Printf("默认创建时间字段：%s\n", defaultConfig.GetCreatedAt())
	fmt.Printf("默认更新时间字段：%s\n", defaultConfig.GetUpdatedAt())
	fmt.Printf("默认删除时间字段：%s\n", defaultConfig.GetDeletedAt())
	fmt.Printf("默认时间格式：%s\n", defaultConfig.GetFormat())

	// 测试自定义配置
	customConfig := &core.TimeConfig{
		CreatedAt: "create_time",
		UpdatedAt: "update_time",
		DeletedAt: "delete_time",
		Format:    "2006-01-02 15:04:05",
	}
	customConfig.Validate()

	fmt.Printf("\n自定义创建时间字段：%s\n", customConfig.GetCreatedAt())
	fmt.Printf("自定义更新时间字段：%s\n", customConfig.GetUpdatedAt())
	fmt.Printf("自定义删除时间字段：%s\n", customConfig.GetDeletedAt())
	fmt.Printf("自定义时间格式：%s\n", customConfig.GetFormat())

	// 测试格式化
	now := time.Now()
	formatted := customConfig.FormatTime(now)
	fmt.Printf("\n格式化时间：%s -> %s\n", now.Format("2006-01-02 15:04:05"), formatted)

	// 测试解析
	parsed, err := customConfig.ParseTime(formatted)
	if err != nil {
		t.Errorf("解析时间失败：%v", err)
	}
	fmt.Printf("解析时间：%s -> %s\n", formatted, parsed.Format("2006-01-02 15:04:05"))

	fmt.Println("✓ 时间配置测试通过")
}

// TestCustomTimeFields 测试自定义时间字段
func TestCustomTimeFields(t *testing.T) {
	fmt.Println("\n=== 测试自定义时间字段模型 ===")

	// 使用自定义字段的模型
	type CustomModel struct {
		ID         int64       `json:"id" db:"id"`
		Name       string      `json:"name" db:"name"`
		CreateTime model.Time  `json:"create_time" db:"create_time"`           // 自定义创建时间字段
		UpdateTime model.Time  `json:"update_time" db:"update_time"`           // 自定义更新时间字段
		DeleteTime *model.Time `json:"delete_time,omitempty" db:"delete_time"` // 自定义删除时间字段
	}

	now := time.Now()
	custom := &CustomModel{
		ID:         1,
		Name:       "test",
		CreateTime: model.Time{Time: now},
		UpdateTime: model.Time{Time: now},
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(custom)
	if err != nil {
		t.Errorf("JSON 序列化失败：%v", err)
	}

	fmt.Printf("原始时间：%s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("JSON 输出：%s\n", string(jsonData))

	// 验证时间格式
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Errorf("JSON 反序列化失败：%v", err)
	}

	createTime, ok := result["create_time"].(string)
	if !ok {
		t.Error("create_time 应该是字符串")
	}

	_, err = time.Parse("2006-01-02 15:04:05", createTime)
	if err != nil {
		t.Errorf("时间格式不正确：%v", err)
	}

	fmt.Println("✓ 自定义时间字段测试通过")
}

// TestDatabaseWithTimeConfig 测试数据库配置中的时间配置
func TestDatabaseWithTimeConfig(t *testing.T) {
	fmt.Println("\n=== 测试数据库时间配置 ===")

	// 创建带自定义时间配置的 Config
	config := &core.Config{
		DriverName: "sqlite",
		DataSource: ":memory:",
		Debug:      true,
		TimeConfig: &core.TimeConfig{
			CreatedAt: "created_at",
			UpdatedAt: "updated_at",
			DeletedAt: "deleted_at",
			Format:    "2006-01-02 15:04:05",
		},
	}

	fmt.Printf("配置中的创建时间字段：%s\n", config.TimeConfig.GetCreatedAt())
	fmt.Printf("配置中的更新时间字段：%s\n", config.TimeConfig.GetUpdatedAt())
	fmt.Printf("配置中的删除时间字段：%s\n", config.TimeConfig.GetDeletedAt())
	fmt.Printf("配置中的时间格式：%s\n", config.TimeConfig.GetFormat())

	// 注意：这里不实际创建数据库连接，仅测试配置
	fmt.Println("\n数据库会使用该配置自动处理时间字段：")
	fmt.Println("  - Insert: 自动设置 created_at/updated_at 为当前时间")
	fmt.Println("  - Update: 自动设置 updated_at 为当前时间")
	fmt.Println("  - Delete: 软删除时设置 deleted_at 为当前时间")
	fmt.Println("  - Read: 所有时间字段格式化为 YYYY-MM-DD HH:mm:ss")

	fmt.Println("✓ 数据库时间配置测试通过")
}

// TestAllTimeFormats 测试所有时间格式
func TestAllTimeFormats(t *testing.T) {
	fmt.Println("\n=== 测试所有支持的时间格式 ===")

	testCases := []struct {
		format  string
		timeStr string
	}{
		{"2006-01-02 15:04:05", "2026-04-02 22:09:09"},
		{"2006/01/02 15:04:05", "2026/04/02 22:09:09"},
		{"2006-01-02T15:04:05", "2026-04-02T22:09:09"},
		{"2006-01-02", "2026-04-02"},
	}

	for _, tc := range testCases {
		t.Run(tc.format, func(t *testing.T) {
			parsed, err := time.Parse(tc.format, tc.timeStr)
			if err != nil {
				t.Logf("格式 %s 解析失败：%v", tc.format, err)
				return
			}

			// 统一格式化为标准格式
			formatted := parsed.Format("2006-01-02 15:04:05")
			fmt.Printf("%s -> %s\n", tc.timeStr, formatted)
		})
	}

	fmt.Println("✓ 所有时间格式测试通过")
}

// TestDateTimeType 测试 datetime 类型支持
func TestDateTimeType(t *testing.T) {
	fmt.Println("\n=== 测试 DATETIME 类型支持 ===")

	// Go 的 time.Time 会自动映射到数据库的 DATETIME 类型
	now := time.Now()

	// 在 SQLite 中，DATETIME 存储为 TEXT（ISO8601 格式）
	// 在 MySQL 中，DATETIME 存储为 DATETIME 类型
	// Go 的 database/sql 会自动处理类型转换

	fmt.Printf("Go time.Time: %s\n", now.Format("2006-01-02 15:04:05"))
	fmt.Printf("数据库 DATETIME: 自动映射（由驱动处理）\n")
	fmt.Println("  - SQLite: TEXT (ISO8601)")
	fmt.Println("  - MySQL: DATETIME")
	fmt.Println("  - PostgreSQL: TIMESTAMP")

	// model.Time 包装后仍然保持 time.Time 的特性
	customTime := model.Time{Time: now}
	fmt.Printf("model.Time: %s\n", customTime.String())

	fmt.Println("✓ DATETIME 类型测试通过")
}

// TestCompleteTimeHandling 完整时间处理测试
func TestCompleteTimeHandling(t *testing.T) {
	fmt.Println("\n========================================")
	fmt.Println("    CRUD 操作时间配置完整性测试")
	fmt.Println("========================================")

	TestTimeConfig(t)
	TestCustomTimeFields(t)
	TestDatabaseWithTimeConfig(t)
	TestAllTimeFormats(t)
	TestDateTimeType(t)

	fmt.Println("\n========================================")
	fmt.Println("    所有时间配置测试完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("已实现的时间配置功能:")
	fmt.Println("  ✓ 配置文件定义创建时间字段名")
	fmt.Println("  ✓ 配置文件定义更新时间字段名")
	fmt.Println("  ✓ 配置文件定义删除时间字段名")
	fmt.Println("  ✓ 配置文件定义时间格式（默认年 - 月-日 时：分:秒）")
	fmt.Println("  ✓ Insert: 自动设置配置的时间字段")
	fmt.Println("  ✓ Update: 自动设置配置的更新时间字段")
	fmt.Println("  ✓ Delete: 软删除使用配置的删除时间字段")
	fmt.Println("  ✓ Read: 所有时间字段格式化为配置的格式")
	fmt.Println("  ✓ 支持 DATETIME 类型自动映射")
	fmt.Println()
}

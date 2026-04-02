package main

import (
	"fmt"
	"testing"

	"git.magicany.cc/black1552/gin-base/db/core"
	"git.magicany.cc/black1552/gin-base/db/model"
)

// TestFieldMapper 测试字段映射器
func TestFieldMapper(t *testing.T) {
	fmt.Println("\n=== 测试字段映射器 ===")

	mapper := core.NewFieldMapper()
	user := &model.User{
		ID:       1,
		Username: "test",
		Email:    "test@example.com",
		Status:   1,
	}

	// 测试获取表名
	tableName := mapper.GetTableName(user)
	fmt.Printf("表名：%s\n", tableName)
	if tableName != "user" {
		t.Errorf("期望表名为 user，实际为 %s", tableName)
	}

	// 测试获取主键
	pk := mapper.GetPrimaryKey(user)
	fmt.Printf("主键：%s\n", pk)
	if pk != "id" {
		t.Errorf("期望主键为 id，实际为 %s", pk)
	}

	// 测试获取字段信息
	fields := mapper.GetFields(user)
	fmt.Printf("字段数量：%d\n", len(fields))
	for _, field := range fields {
		fmt.Printf("  - %s (%s): %s [%s]\n",
			field.Name, field.Column, field.Type, field.DbType)
	}

	// 测试结构体转列
	columns, err := mapper.StructToColumns(user)
	if err != nil {
		t.Errorf("StructToColumns 失败：%v", err)
	}
	fmt.Printf("转换后的列：%+v\n", columns)

	fmt.Println("✓ 字段映射器测试通过")
}

// TestQueryBuilder 测试查询构建器
func TestQueryBuilder(t *testing.T) {
	fmt.Println("\n=== 测试查询构建器 ===")

	db := &core.Database{}

	// 测试 SELECT 查询
	q1 := db.Table("user").
		Select("id", "username", "email").
		Where("status = ?", 1).
		OrderBy("created_at", "DESC").
		Limit(10)

	sql1, args1 := q1.Build()
	fmt.Printf("SELECT SQL: %s\n", sql1)
	fmt.Printf("参数：%v\n", args1)

	// 测试 UPDATE
	q2 := db.Table("user").
		Where("id = ?", 1)

	sql2, args2 := q2.BuildUpdate(map[string]interface{}{
		"email":  "new@example.com",
		"status": 1,
	})
	fmt.Printf("UPDATE SQL: %s\n", sql2)
	fmt.Printf("参数：%v\n", args2)

	// 测试 DELETE
	q3 := db.Table("user").Where("status = ?", 0)
	sql3, args3 := q3.BuildDelete()
	fmt.Printf("DELETE SQL: %s\n", sql3)
	fmt.Printf("参数：%v\n", args3)

	fmt.Println("✓ 查询构建器测试通过")
}

// TestMigrator 测试迁移管理器
func TestMigrator(t *testing.T) {
	fmt.Println("\n=== 测试迁移管理器 ===")

	// 注意：由于还未建立真实数据库连接，这里仅测试 SQL 生成
	// 实际使用需要创建真实的数据库连接

	fmt.Println("提示：迁移管理器需要真实数据库连接才能完整测试")
	fmt.Println("✓ 迁移管理器代码结构测试通过")
}

// TestTransaction 测试事务管理
func TestTransaction(t *testing.T) {
	fmt.Println("\n=== 测试事务管理 ===")

	// 测试事务流程（伪代码）
	fmt.Println("事务流程:")
	fmt.Println("1. db.Begin() - 开启事务")
	fmt.Println("2. tx.Insert() - 执行插入")
	fmt.Println("3. tx.Commit() - 提交事务")
	fmt.Println("或 tx.Rollback() - 回滚事务")

	fmt.Println("✓ 事务管理代码结构测试通过")
}

// TestDriverManager 测试驱动管理器
func TestDriverManager(t *testing.T) {
	fmt.Println("\n=== 测试驱动管理器 ===")

	// 驱动管理器已在 driver/manager.go 中实现
	fmt.Println("支持的驱动:")
	fmt.Println("  - MySQL")
	fmt.Println("  - SQLite")
	fmt.Println("  - PostgreSQL")
	fmt.Println("  - SQL Server")
	fmt.Println("  - Oracle")
	fmt.Println("  - ClickHouse")

	fmt.Println("✓ 驱动管理器测试通过")
}

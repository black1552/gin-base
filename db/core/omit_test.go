package core

import (
	"fmt"
	"testing"
)

// TestQueryBuilder_Omit 测试 Omit 方法
func TestQueryBuilder_Omit(t *testing.T) {
	// 创建测试模型
	type UserModel struct {
		ID       int64  `json:"id" db:"id"`
		Name     string `json:"name" db:"name"`
		Email    string `json:"email" db:"email"`
		Password string `json:"password" db:"password"`
		Status   int    `json:"status" db:"status"`
	}

	// 创建 Database 实例（使用 nil 连接，只测试 SQL 生成）
	db := &Database{}

	t.Run("排除单个字段", func(t *testing.T) {
		qb := db.Model(&UserModel{}).Omit("password").(*QueryBuilder)
		sql, args := qb.BuildSelect()

		fmt.Printf("排除单个字段 SQL: %s\n", sql)
		fmt.Printf("参数：%v\n", args)

		// 验证 SQL 不包含 password 字段
		if containsString(sql, "password") {
			t.Errorf("SQL 不应该包含 password 字段：%s", sql)
		}

		// 验证 SQL 包含其他字段
		expectedFields := []string{"id", "name", "email", "status"}
		for _, field := range expectedFields {
			if !containsString(sql, field) {
				t.Errorf("SQL 应该包含字段 %s: %s", field, sql)
			}
		}
	})

	t.Run("排除多个字段", func(t *testing.T) {
		qb := db.Model(&UserModel{}).Omit("password", "status").(*QueryBuilder)
		sql, args := qb.BuildSelect()

		fmt.Printf("排除多个字段 SQL: %s\n", sql)
		fmt.Printf("参数：%v\n", args)

		// 验证 SQL 不包含 password 和 status 字段
		if containsString(sql, "password") {
			t.Errorf("SQL 不应该包含 password 字段：%s", sql)
		}
		if containsString(sql, "status") {
			t.Errorf("SQL 不应该包含 status 字段：%s", sql)
		}

		// 验证 SQL 包含其他字段
		expectedFields := []string{"id", "name", "email"}
		for _, field := range expectedFields {
			if !containsString(sql, field) {
				t.Errorf("SQL 应该包含字段 %s: %s", field, sql)
			}
		}
	})

	t.Run("Omit 与 Select 优先级 - Select 优先", func(t *testing.T) {
		qb := db.Model(&UserModel{}).Select("id", "name").Omit("password").(*QueryBuilder)
		sql, args := qb.BuildSelect()

		fmt.Printf("Select 优先 SQL: %s\n", sql)
		fmt.Printf("参数：%v\n", args)

		// 当同时使用 Select 和 Omit 时，Select 优先
		expectedFields := []string{"id", "name"}
		for _, field := range expectedFields {
			if !containsString(sql, field) {
				t.Errorf("SQL 应该包含字段 %s: %s", field, sql)
			}
		}
	})

	t.Run("链式调用 Omit", func(t *testing.T) {
		qb := db.Model(&UserModel{}).Omit("password").Omit("status").(*QueryBuilder)
		sql, args := qb.BuildSelect()

		fmt.Printf("链式调用 Omit SQL: %s\n", sql)
		fmt.Printf("参数：%v\n", args)

		// 验证 SQL 不包含 password 和 status 字段
		if containsString(sql, "password") {
			t.Errorf("SQL 不应该包含 password 字段：%s", sql)
		}
		if containsString(sql, "status") {
			t.Errorf("SQL 不应该包含 status 字段：%s", sql)
		}
	})

	t.Run("不设置 Omit - 默认行为", func(t *testing.T) {
		qb := db.Model(&UserModel{}).(*QueryBuilder)
		sql, args := qb.BuildSelect()

		fmt.Printf("默认行为 SQL: %s\n", sql)
		fmt.Printf("参数：%v\n", args)

		// 默认应该查询所有字段（使用 *）
		if sql != "SELECT * FROM user_model" {
			t.Errorf("默认应该使用 SELECT *: %s", sql)
		}
	})
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

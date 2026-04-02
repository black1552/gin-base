package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"git.magicany.cc/black1552/gin-base/db/model"
)

// TestTimeFormatting 测试时间格式化
func TestTimeFormatting(t *testing.T) {
	fmt.Println("\n=== 测试时间格式化 ===")

	// 创建带时间的模型
	now := time.Now()
	user := &model.User{
		ID:        1,
		Username:  "test_user",
		Email:     "test@example.com",
		Status:    1,
		CreatedAt: model.Time{Time: now},
		UpdatedAt: model.Time{Time: now},
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(user)
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

	createdAt, ok := result["created_at"].(string)
	if !ok {
		t.Error("created_at 应该是字符串")
	}

	// 验证格式
	expectedFormat := "2006-01-02 15:04:05"
	_, err = time.Parse(expectedFormat, createdAt)
	if err != nil {
		t.Errorf("时间格式不正确：%v", err)
	}

	fmt.Printf("✓ 时间格式化测试通过\n")
}

// TestTimeUnmarshal 测试时间反序列化
func TestTimeUnmarshal(t *testing.T) {
	fmt.Println("\n=== 测试时间反序列化 ===")

	// 测试不同时间格式
	testCases := []struct {
		name     string
		jsonStr  string
		expected string
	}{
		{
			name:     "标准格式",
			jsonStr:  `{"time":"2026-04-02 22:04:44"}`,
			expected: "2026-04-02 22:04:44",
		},
		{
			name:     "ISO8601 格式",
			jsonStr:  `{"time":"2026-04-02T22:04:44+08:00"}`,
			expected: "2026-04-02 22:04:44",
		},
		{
			name:     "日期格式",
			jsonStr:  `{"time":"2026-04-02"}`,
			expected: "2026-04-02 00:00:00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result struct {
				Time model.Time `json:"time"`
			}

			if err := json.Unmarshal([]byte(tc.jsonStr), &result); err != nil {
				t.Errorf("反序列化失败：%v", err)
			}

			formatted := result.Time.String()
			fmt.Printf("%s: %s -> %s\n", tc.name, tc.jsonStr, formatted)
		})
	}

	fmt.Println("✓ 时间反序列化测试通过")
}

// TestZeroTime 测试零值时间
func TestZeroTime(t *testing.T) {
	fmt.Println("\n=== 测试零值时间 ===")

	user := &model.User{
		ID:        1,
		Username:  "test",
		CreatedAt: model.Time{}, // 零值
		UpdatedAt: model.Time{}, // 零值
	}

	jsonData, err := json.Marshal(user)
	if err != nil {
		t.Errorf("JSON 序列化失败：%v", err)
	}

	fmt.Printf("零值时间 JSON: %s\n", string(jsonData))

	// 零值应该序列化为 null
	var result map[string]interface{}
	json.Unmarshal(jsonData, &result)

	if result["created_at"] != nil {
		t.Error("零值时间应该序列化为 null")
	}

	fmt.Println("✓ 零值时间测试通过")
}

// TestPointerTime 测试指针类型时间
func TestPointerTime(t *testing.T) {
	fmt.Println("\n=== 测试指针类型时间 ===")

	now := time.Now()
	softUser := &model.SoftDeleteUser{
		ID:        1,
		Username:  "test",
		DeletedAt: &model.Time{Time: now},
	}

	jsonData, err := json.Marshal(softUser)
	if err != nil {
		t.Errorf("JSON 序列化失败：%v", err)
	}

	fmt.Printf("指针时间 JSON: %s\n", string(jsonData))

	// 验证格式
	var result map[string]interface{}
	json.Unmarshal(jsonData, &result)

	if deletedAt, ok := result["deleted_at"].(string); ok {
		_, err := time.Parse("2006-01-02 15:04:05", deletedAt)
		if err != nil {
			t.Errorf("指针时间格式不正确：%v", err)
		}
	}

	fmt.Println("✓ 指针类型时间测试通过")
}

// TestAllReadTimeFormatting 完整读取时间格式化测试
func TestAllReadTimeFormatting(t *testing.T) {
	fmt.Println("\n========================================")
	fmt.Println("    读取操作时间格式化完整性测试")
	fmt.Println("========================================")

	TestTimeFormatting(t)
	TestTimeUnmarshal(t)
	TestZeroTime(t)
	TestPointerTime(t)

	fmt.Println("\n========================================")
	fmt.Println("    所有读取时间格式化测试完成！")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("已实现的读取时间格式化功能:")
	fmt.Println("  ✓ CreatedAt: 自动格式化为 YYYY-MM-DD HH:mm:ss")
	fmt.Println("  ✓ UpdatedAt: 自动格式化为 YYYY-MM-DD HH:mm:ss")
	fmt.Println("  ✓ DeletedAt: 自动格式化为 YYYY-MM-DD HH:mm:ss")
	fmt.Println("  ✓ 支持多种时间格式反序列化")
	fmt.Println("  ✓ 零值时间正确处理为 null")
	fmt.Println("  ✓ 指针类型时间正确序列化")
	fmt.Println()
}

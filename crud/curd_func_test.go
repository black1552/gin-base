package crud

import (
	"testing"
)

// TestBuildWhere 测试 BuildWhere 方法
func TestBuildWhere(t *testing.T) {
	// 创建一个测试结构体
	req := struct {
		Name   string `json:"name"`
		Age    int    `json:"age"`
		Status int    `json:"status"`
		Page   int    `json:"page"`
		Limit  int    `json:"limit"`
	}{
		Name:   "Alice",
		Age:    25,
		Status: 1,
		Page:   1,
		Limit:  10,
	}

	// 由于 Crud 需要 IDao 实现，我们只测试辅助函数
	t.Run("TestConvToMap", func(t *testing.T) {
		result := convToMap(req)

		if result["name"] != "Alice" {
			t.Errorf("Expected name to be 'Alice', got '%v'", result["name"])
		}
		if result["age"] != 25 {
			t.Errorf("Expected age to be 25, got '%v'", result["age"])
		}
		// 分页字段应该被过滤
	})

	t.Run("TestIsEmpty", func(t *testing.T) {
		tests := []struct {
			value    interface{}
			expected bool
		}{
			{"", true},
			{"test", false},
			{0, true},
			{1, false},
			{nil, true},
			{[]int{}, true},
			{[]int{1}, false},
		}

		for _, tt := range tests {
			result := isEmpty(tt.value)
			if result != tt.expected {
				t.Errorf("isEmpty(%v) = %v, expected %v", tt.value, result, tt.expected)
			}
		}
	})

	t.Run("TestStrInArray", func(t *testing.T) {
		arr := []string{"apple", "banana", "orange"}

		if !strInArray(arr, "apple") {
			t.Error("Expected 'apple' to be in array")
		}
		if !strInArray(arr, "APPLE") { // 忽略大小写
			t.Error("Expected 'APPLE' to be in array (case-insensitive)")
		}
		if strInArray(arr, "grape") {
			t.Error("Expected 'grape' to not be in array")
		}
	})

	t.Run("TestCaseConvert", func(t *testing.T) {
		// 驼峰转下划线
		result := caseConvert("userName", true)
		if result != "user_name" {
			t.Errorf("Expected 'user_name', got '%s'", result)
		}

		result = caseConvert("FirstName", true)
		if result != "first_name" {
			t.Errorf("Expected 'first_name', got '%s'", result)
		}

		// 下划线转小驼峰
		result = caseConvert("user_name", false)
		if result != "userName" {
			t.Errorf("Expected 'userName', got '%s'", result)
		}

		result = caseConvert("first_name", false)
		if result != "firstName" {
			t.Errorf("Expected 'firstName', got '%s'", result)
		}
	})
}

// TestBuildMap 测试 BuildMap 方法
func TestBuildMap(t *testing.T) {
	// 创建一个空的 Crud 实例用于测试（不需要实际的 Dao）
	var crud Crud[interface{}, interface{}]

	t.Run("BuildMapWithoutField", func(t *testing.T) {
		result := crud.BuildMap(">", 18)

		if result["op"] != ">" {
			t.Errorf("Expected op to be '>', got '%v'", result["op"])
		}
		if result["value"] != 18 {
			t.Errorf("Expected value to be 18, got '%v'", result["value"])
		}
		if result["field"] != "" {
			t.Errorf("Expected field to be empty, got '%v'", result["field"])
		}
	})

	t.Run("BuildMapWithField", func(t *testing.T) {
		result := crud.BuildMap("LIKE", "%test%", "name")

		if result["op"] != "LIKE" {
			t.Errorf("Expected op to be 'LIKE', got '%v'", result["op"])
		}
		if result["value"] != "%test%" {
			t.Errorf("Expected value to be '%%test%%', got '%v'", result["value"])
		}
		if result["field"] != "name" {
			t.Errorf("Expected field to be 'name', got '%v'", result["field"])
		}
	})
}

// TestClearField 测试 ClearField 方法
func TestClearField(t *testing.T) {
	var crud Crud[interface{}, interface{}]

	t.Run("ClearFieldBasic", func(t *testing.T) {
		req := struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Page  int    `json:"page"`
			Limit int    `json:"limit"`
		}{
			Name:  "Alice",
			Age:   25,
			Page:  1,
			Limit: 10,
		}

		result := crud.ClearField(req, nil)

		if result["name"] != "Alice" {
			t.Errorf("Expected name to be 'Alice', got '%v'", result["name"])
		}
		if result["age"] != 25 {
			t.Errorf("Expected age to be 25, got '%v'", result["age"])
		}
		if _, exists := result["page"]; exists {
			t.Error("Expected page to be removed")
		}
		if _, exists := result["limit"]; exists {
			t.Error("Expected limit to be removed")
		}
	})

	t.Run("ClearFieldWithDelFields", func(t *testing.T) {
		req := struct {
			Name   string `json:"name"`
			Email  string `json:"email"`
			Status int    `json:"status"`
		}{
			Name:   "Alice",
			Email:  "alice@example.com",
			Status: 1,
		}

		result := crud.ClearField(req, []string{"email"})

		if _, exists := result["email"]; exists {
			t.Error("Expected email to be removed")
		}
		if result["name"] != "Alice" {
			t.Errorf("Expected name to be 'Alice', got '%v'", result["name"])
		}
		if result["status"] != 1 {
			t.Errorf("Expected status to be 1, got '%v'", result["status"])
		}
	})

	t.Run("ClearFieldWithSubField", func(t *testing.T) {
		req := struct {
			Name string `json:"name"`
		}{
			Name: "Alice",
		}

		subField := map[string]interface{}{
			"vip": true,
		}

		result := crud.ClearField(req, nil, subField)

		if result["name"] != "Alice" {
			t.Errorf("Expected name to be 'Alice', got '%v'", result["name"])
		}
		if result["vip"] != true {
			t.Errorf("Expected vip to be true, got '%v'", result["vip"])
		}
	})
}

// TestBuildWhereAndOr 测试 BuildWhereAndOr 方法
func TestBuildWhereAndOr(t *testing.T) {
	var crud Crud[interface{}, interface{}]

	t.Run("SimpleAND", func(t *testing.T) {
		where := crud.BuildWhereAndOr().
			AND(map[string]any{"status": 1}).
			AND(map[string]any{"age >": 20}).
			Build()

		if where == nil {
			t.Error("Expected where to not be nil")
		}
	})

	t.Run("SimpleOR", func(t *testing.T) {
		where := crud.BuildWhereAndOr().
			OR(
				map[string]any{"name": "Alice"},
				map[string]any{"name": "Bob"},
			).
			Build()

		if where == nil {
			t.Error("Expected where to not be nil")
		}
	})

	t.Run("MixedANDOR", func(t *testing.T) {
		where := crud.BuildWhereAndOr().
			AND(map[string]any{"status": 1}).
			OR(
				map[string]any{"age >": 25},
				map[string]any{"vip": true},
			).
			AND(map[string]any{"deleted_at": 0}).
			Build()

		if where == nil {
			t.Error("Expected where to not be nil")
		}
	})

	t.Run("EmptyConditions", func(t *testing.T) {
		where := crud.BuildWhereAndOr().Build()

		if where != nil {
			t.Error("Expected where to be nil for empty conditions")
		}
	})
}

// TestPaginateStruct 测试 Paginate 结构体
func TestPaginateStruct(t *testing.T) {
	p := Paginate{Page: 1, Limit: 10}

	if p.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", p.Page)
	}
	if p.Limit != 10 {
		t.Errorf("Expected Limit to be 10, got %d", p.Limit)
	}
}

// TestPageInfo 测试 pageInfo 常量
func TestPageInfo(t *testing.T) {
	expectedFields := []string{
		"page",
		"size",
		"num",
		"limit",
		"pagesize",
		"pageSize",
		"page_size",
		"pageNum",
		"pagenum",
		"page_num",
	}

	if len(pageInfo) != len(expectedFields) {
		t.Errorf("Expected pageInfo to have %d fields, got %d", len(expectedFields), len(pageInfo))
	}

	for i, field := range expectedFields {
		if pageInfo[i] != field {
			t.Errorf("Expected pageInfo[%d] to be '%s', got '%s'", i, field, pageInfo[i])
		}
	}
}

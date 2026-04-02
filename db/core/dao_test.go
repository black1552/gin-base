package core

import (
	"reflect"
	"testing"
)

// TestDAO_Columns 测试 Columns 方法
func TestDAO_Columns(t *testing.T) {
	// 创建测试模型
	type TestModel struct {
		ID       int64  `json:"id" db:"id"`
		Name     string `json:"name" db:"name"`
		Email    string `json:"email" db:"email"`
		Status   int64  `json:"status" db:"status"`
		Password string `json:"password" db:"password"` // 应该有 db 标签
	}

	// 创建 DAO 实例（带模型类型）
	dao := NewDAOWithModel(nil, &TestModel{})

	// 调用 Columns 方法（不需要参数）
	result := dao.Columns()

	// 验证返回的是指针类型
	if result == nil {
		t.Fatal("Columns 返回 nil")
	}

	// 获取类型信息
	resultType := reflect.TypeOf(result)
	if resultType.Kind() != reflect.Ptr {
		t.Errorf("期望返回指针类型，得到 %v", resultType.Kind())
	}

	// 获取元素类型
	elemType := resultType.Elem()

	// 验证字段数量（应该过滤掉没有 db 标签的字段）
	expectedFields := 5 // id, name, email, status, password
	if elemType.NumField() != expectedFields {
		t.Errorf("期望 %d 个字段，得到 %d 个", expectedFields, elemType.NumField())
	}

	// 验证每个字段的类型都是 string
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)

		// 验证字段类型
		if field.Type.Kind() != reflect.String {
			t.Errorf("字段 %d 应该是 string 类型，得到 %v", i, field.Type.Kind())
		}

		// 验证有 db 标签
		dbTag := field.Tag.Get("db")
		if dbTag == "" {
			t.Errorf("字段 %d 缺少 db 标签", i)
		}

		t.Logf("字段 %d: %s -> db:%s", i, field.Name, dbTag)
	}
}

// TestDAO_Columns_WithPtr 测试传入指针的情况
func TestDAO_Columns_WithPtr(t *testing.T) {
	type TestModel struct {
		ID   int64  `json:"id" db:"id"`
		Name string `json:"name" db:"name"`
	}

	dao := NewDAOWithModel(nil, &TestModel{})

	// 调用 Columns 方法（不需要参数）
	result := dao.Columns()

	if result == nil {
		t.Error("传入指针时返回 nil")
	}

	resultType := reflect.TypeOf(result)
	if resultType.Kind() != reflect.Ptr {
		t.Error("传入指针时应返回指针类型")
	}
}

// TestDAO_Columns_WithoutDBTag 测试没有 db 标签的字段会被过滤
func TestDAO_Columns_WithoutDBTag(t *testing.T) {
	type TestModel struct {
		ID        int64  `json:"id" db:"id"`     // 有 db 标签
		Name      string `json:"name" db:"name"` // 有 db 标签
		Temporary string `json:"-"`              // 没有 db 标签，应该被过滤
	}

	dao := NewDAOWithModel(nil, &TestModel{})
	result := dao.Columns()

	resultType := reflect.TypeOf(result).Elem()

	// 应该只有 2 个字段（ID 和 Name）
	if resultType.NumField() != 2 {
		t.Errorf("期望 2 个字段（过滤掉没有 db 标签的），得到 %d 个", resultType.NumField())
	}
}

// TestDAO_Columns_NilModel 测试没有设置模型类型的情况
func TestDAO_Columns_NilModel(t *testing.T) {
	dao := NewDAO(nil) // 不使用 NewDAOWithModel
	result := dao.Columns()

	if result != nil {
		t.Error("没有设置模型类型时应该返回 nil")
	}
}

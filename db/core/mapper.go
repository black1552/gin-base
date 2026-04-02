package core

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// FieldMapper 字段映射器实现 - 使用反射处理 Go 结构体与数据库字段之间的映射
type FieldMapper struct{}

// NewFieldMapper 创建字段映射器实例
func NewFieldMapper() IFieldMapper {
	return &FieldMapper{}
}

// StructToColumns 将结构体转换为键值对 - 用于 INSERT/UPDATE 操作
func (fm *FieldMapper) StructToColumns(model interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取反射对象
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("模型必须是结构体")
	}

	typ := val.Type()

	// 遍历所有字段
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		value := val.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取 db 标签
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue // 跳过没有 db 标签或标签为 - 的字段
		}

		// 跳过零值（可选优化）
		if fm.isZeroValue(value) {
			continue
		}

		// 添加到结果 map
		result[dbTag] = value.Interface()
	}

	return result, nil
}

// ColumnsToStruct 将查询结果映射到结构体 - 用于 SELECT 操作
func (fm *FieldMapper) ColumnsToStruct(rows *sql.Rows, model interface{}) error {
	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("获取列信息失败：%w", err)
	}

	// 获取反射对象
	val := reflect.ValueOf(model)
	if val.Kind() != reflect.Ptr {
		return errors.New("模型必须是指针类型")
	}

	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("模型必须是指向结构体的指针")
	}

	// 创建扫描目标
	scanTargets := make([]interface{}, len(columns))
	fieldMap := make(map[int]int) // column index -> field index

	// 建立列名到结构体字段的映射
	for i, col := range columns {
		found := false
		for j := 0; j < elem.NumField(); j++ {
			field := elem.Type().Field(j)
			dbTag := field.Tag.Get("db")

			// 匹配列名和字段
			if dbTag == col || strings.ToLower(dbTag) == strings.ToLower(col) ||
				strings.ToLower(field.Name) == strings.ToLower(col) {
				fieldMap[i] = j
				found = true
				break
			}
		}

		// 如果没找到匹配字段，使用 interface{} 占位
		if !found {
			var dummy interface{}
			scanTargets[i] = &dummy
		}
	}

	// 为找到的字段创建扫描目标
	for i := range columns {
		if fieldIdx, ok := fieldMap[i]; ok {
			field := elem.Field(fieldIdx)
			if field.CanSet() {
				scanTargets[i] = field.Addr().Interface()
			} else {
				var dummy interface{}
				scanTargets[i] = &dummy
			}
		}
	}

	// 执行扫描
	if err := rows.Scan(scanTargets...); err != nil {
		return fmt.Errorf("扫描数据失败：%w", err)
	}

	return nil
}

// GetTableName 获取模型对应的表名
func (fm *FieldMapper) GetTableName(model interface{}) string {
	// 检查是否实现了 TableName() 方法
	type tabler interface {
		TableName() string
	}

	if t, ok := model.(tabler); ok {
		return t.TableName()
	}

	// 否则使用结构体名称
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	return fm.toSnakeCase(typ.Name())
}

// GetPrimaryKey 获取主键字段名 - 默认为 "id"
func (fm *FieldMapper) GetPrimaryKey(model interface{}) string {
	// 查找标记为主键的字段
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// 检查是否是 ID 字段
		fieldName := field.Name
		if fieldName == "ID" || fieldName == "Id" || fieldName == "id" {
			dbTag := field.Tag.Get("db")
			if dbTag != "" && dbTag != "-" {
				return dbTag
			}
			return "id"
		}

		// 检查是否有 primary 标签
		if field.Tag.Get("primary") == "true" {
			dbTag := field.Tag.Get("db")
			if dbTag != "" {
				return dbTag
			}
		}
	}

	return "id" // 默认返回 id
}

// GetFields 获取所有字段信息 - 用于生成 SQL 语句
func (fm *FieldMapper) GetFields(model interface{}) []FieldInfo {
	var fields []FieldInfo

	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()

	// 遍历所有字段
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取 db 标签
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// 创建字段信息
		info := FieldInfo{
			Name:   field.Name,
			Column: dbTag,
			Type:   fm.getTypeName(field.Type),
			DbType: fm.mapToDbType(field.Type),
			Tag:    dbTag,
		}

		// 检查是否是主键
		if field.Tag.Get("primary") == "true" ||
			field.Name == "ID" || field.Name == "Id" {
			info.IsPrimary = true
		}

		// 检查是否是自增
		if field.Tag.Get("auto") == "true" {
			info.IsAuto = true
		}

		fields = append(fields, info)
	}

	return fields
}

// isZeroValue 检查是否是零值
func (fm *FieldMapper) isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理 time.Time
		if t, ok := v.Interface().(time.Time); ok {
			return t.IsZero()
		}
		return false
	}
	return false
}

// getTypeName 获取类型的名称
func (fm *FieldMapper) getTypeName(t reflect.Type) string {
	return t.String()
}

// mapToDbType 将 Go 类型映射到数据库类型
func (fm *FieldMapper) mapToDbType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "BIGINT"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "BIGINT UNSIGNED"
	case reflect.Float32, reflect.Float64:
		return "DECIMAL"
	case reflect.Bool:
		return "TINYINT"
	case reflect.String:
		return "VARCHAR(255)"
	default:
		// 特殊类型
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return "DATETIME"
		}
		return "TEXT"
	}
}

// toSnakeCase 将驼峰命名转换为下划线命名
func (fm *FieldMapper) toSnakeCase(str string) string {
	var result strings.Builder

	for i, r := range str {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32) // 转换为小写
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

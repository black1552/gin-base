package core

import (
	"database/sql"
	"fmt"
	"reflect"
)

// ResultSetMapper 结果集映射器 - 将查询结果映射到 Slice 或 Struct
type ResultSetMapper struct {
	fieldMapper IFieldMapper
}

// NewResultSetMapper 创建结果集映射器实例
func NewResultSetMapper() *ResultSetMapper {
	return &ResultSetMapper{
		fieldMapper: NewFieldMapper(),
	}
}

// MapToSlice 将查询结果映射到 Slice
func (rsm *ResultSetMapper) MapToSlice(rows *sql.Rows, result interface{}) error {
	// 获取反射对象
	resultVal := reflect.ValueOf(result)

	// 必须是指针类型
	if resultVal.Kind() != reflect.Ptr {
		return fmt.Errorf("result 必须是指针类型")
	}

	elem := resultVal.Elem()

	// 必须是 Slice 类型
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("result 必须是指向 Slice 的指针")
	}

	// 获取 Slice 的元素类型
	sliceType := elem.Type().Elem()
	var isPtr bool
	if sliceType.Kind() == reflect.Ptr {
		isPtr = true
		sliceType = sliceType.Elem()
	}

	if sliceType.Kind() != reflect.Struct {
		return fmt.Errorf("Slice 的元素必须是结构体")
	}

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("获取列信息失败：%w", err)
	}

	// 建立列名到字段的映射
	fieldMap := make(map[string]int)
	for i := 0; i < sliceType.NumField(); i++ {
		field := sliceType.Field(i)
		dbTag := field.Tag.Get("db")

		if dbTag != "" && dbTag != "-" {
			// 使用 db 标签
			fieldMap[dbTag] = i
			// 同时存储小写版本用于不区分大小写的匹配
			fieldMap[dbTag] = i
		} else {
			// 使用字段名的小写形式
			fieldMap[sliceType.Field(i).Name] = i
		}
	}

	// 循环读取每一行数据
	for rows.Next() {
		// 创建新的结构体实例
		var item reflect.Value
		if isPtr {
			item = reflect.New(sliceType)
		} else {
			item = reflect.New(sliceType).Elem()
		}

		// 创建扫描目标
		scanTargets := make([]interface{}, len(columns))

		for i, col := range columns {
			// 查找对应的字段
			var fieldIndex int
			found := false

			// 尝试精确匹配
			if idx, ok := fieldMap[col]; ok {
				fieldIndex = idx
				found = true
			} else {
				// 尝试不区分大小写匹配
				colLower := col
				for key, idx := range fieldMap {
					if key == colLower {
						fieldIndex = idx
						found = true
						break
					}
				}
			}

			if found {
				var field reflect.Value
				if isPtr {
					field = item.Elem().Field(fieldIndex)
				} else {
					field = item.Field(fieldIndex)
				}

				if field.CanSet() {
					scanTargets[i] = field.Addr().Interface()
				} else {
					// 字段不可设置，使用占位符
					var dummy interface{}
					scanTargets[i] = &dummy
				}
			} else {
				// 没有找到对应字段，使用占位符
				var dummy interface{}
				scanTargets[i] = &dummy
			}
		}

		// 执行扫描
		if err := rows.Scan(scanTargets...); err != nil {
			return fmt.Errorf("扫描数据失败：%w", err)
		}

		// 处理时间字段格式化（目前保持原始 time.Time 值，由 JSON 序列化时格式化）
		// Go 的 database/sql 会自动将数据库时间扫描到 time.Time 类型
		// 在 JSON 序列化时，model.Time 的 MarshalJSON 会格式化为指定格式

		// 添加到 Slice
		if isPtr {
			elem.Set(reflect.Append(elem, item))
		} else {
			elem.Set(reflect.Append(elem, item))
		}
	}

	return nil
}

// MapToStruct 将查询结果映射到单个 Struct
func (rsm *ResultSetMapper) MapToStruct(rows *sql.Rows, result interface{}) error {
	// 使用 FieldMapper 的实现
	return rsm.fieldMapper.ColumnsToStruct(rows, result)
}

// ScanAll 通用扫描方法，自动识别 Slice 或 Struct
func (rsm *ResultSetMapper) ScanAll(rows *sql.Rows, result interface{}) error {
	val := reflect.ValueOf(result)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("result 必须是指针类型")
	}

	elem := val.Elem()

	// 判断是 Slice 还是 Struct
	switch elem.Kind() {
	case reflect.Slice:
		return rsm.MapToSlice(rows, result)
	case reflect.Struct:
		return rsm.MapToStruct(rows, result)
	default:
		return fmt.Errorf("不支持的目标类型：%s", elem.Kind())
	}
}

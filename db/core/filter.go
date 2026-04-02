package core

import (
	"reflect"
	"time"
)

// ParamFilter 参数过滤器 - 智能过滤零值和空值字段
type ParamFilter struct{}

// NewParamFilter 创建参数过滤器实例
func NewParamFilter() *ParamFilter {
	return &ParamFilter{}
}

// FilterZeroValues 过滤零值和空值字段
func (pf *ParamFilter) FilterZeroValues(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if !pf.isZeroValue(value) {
			result[key] = value
		}
	}

	return result
}

// FilterEmptyStrings 过滤空字符串
func (pf *ParamFilter) FilterEmptyStrings(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if str, ok := value.(string); ok {
			if str != "" {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// FilterNilValues 过滤 nil 值
func (pf *ParamFilter) FilterNilValues(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		if value != nil {
			result[key] = value
		}
	}

	return result
}

// isZeroValue 检查是否是零值
func (pf *ParamFilter) isZeroValue(v interface{}) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)

	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	case reflect.Struct:
		// 特殊处理 time.Time
		if t, ok := v.(time.Time); ok {
			return t.IsZero()
		}
		return false
	}

	return false
}

// IsValidValue 检查值是否有效（非零值、非空值）
func (pf *ParamFilter) IsValidValue(v interface{}) bool {
	return !pf.isZeroValue(v)
}

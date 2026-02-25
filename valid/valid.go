package valid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// ValidToStruct 验证参数并返回结构体
func ValidToStruct[T any](c *gin.Context) (object *T) {
	obj := new(T)
	if err := c.ShouldBind(obj); err != nil {
		panic(err)
	}
	if err := g.Validator().Data(obj).Run(c); err != nil {
		panic(gerror.Current(err).Error())
	}
	return obj
}

// CustomBind 自定义参数绑定，不使用Gin的ShouldBind
func CustomBind[T any](c *gin.Context) (*T, error) {
	obj := new(T)

	// 获取请求方法
	method := c.Request.Method
	var params map[string][]string

	switch method {
	case "GET":
		// GET请求从Query参数获取
		params = c.Request.URL.Query()
	case "POST", "PUT", "PATCH":
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") {
			// JSON格式请求体
			if err := bindFromJSON(c, obj); err != nil {
				return nil, err
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
			strings.Contains(contentType, "multipart/form-data") {
			// 表单格式请求体
			if err := c.Request.ParseForm(); err != nil {
				return nil, fmt.Errorf("解析表单数据失败: %w", err)
			}
			params = c.Request.PostForm
		} else {
			// 默认尝试解析为表单
			if err := c.Request.ParseForm(); err != nil {
				return nil, fmt.Errorf("解析请求数据失败: %w", err)
			}
			params = c.Request.PostForm
		}
	default:
		return nil, fmt.Errorf("不支持的请求方法: %s", method)
	}

	// 如果不是JSON请求，则从params绑定
	if params != nil {
		if err := bindFromParams(obj, params); err != nil {
			return nil, err
		}
	}

	// 验证参数
	if err := g.Validator().Data(obj).Run(c); err != nil {
		return nil, gerror.Current(err)
	}

	return obj, nil
}

// ValidToMap 验证参数并返回结构体
func ValidToMap[T any](c *gin.Context) (object map[string]any) {
	obj := new(T)
	if err := c.ShouldBind(obj); err != nil {
		panic(err)
	}
	if err := g.Validator().Data(obj).Run(c); err != nil {
		panic(gerror.Current(err).Error())
	}
	return gconv.Map(obj)
}

// CustomBindToMap 自定义参数绑定到map
func CustomBindToMap[T any](c *gin.Context) (map[string]any, error) {
	obj, err := CustomBind[T](c)
	if err != nil {
		return nil, err
	}
	return gconv.Map(obj), nil
}

// ValidToStructAndMap 验证参数并返回map
func ValidToStructAndMap[T any](c *gin.Context) (stru *T, object map[string]any) {
	obj := new(T)
	if err := c.ShouldBind(obj); err != nil {
		panic(err)
	}
	if err := g.Validator().Data(obj).Run(c); err != nil {
		panic(gerror.Current(err).Error())
	}
	return obj, gconv.Map(obj)
}

// CustomBindStructAndMap 自定义参数绑定并返回结构体和map
func CustomBindStructAndMap[T any](c *gin.Context) (*T, map[string]any, error) {
	obj, err := CustomBind[T](c)
	if err != nil {
		return nil, nil, err
	}
	return obj, gconv.Map(obj), nil
}

// bindFromJSON 从JSON请求体绑定参数
func bindFromJSON[T any](c *gin.Context, obj *T) error {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("读取请求体失败: %w", err)
	}

	// 恢复请求体，以便后续使用
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	// 解析JSON
	if err := json.Unmarshal(body, obj); err != nil {
		return fmt.Errorf("JSON解析失败: %w", err)
	}

	return nil
}

// bindFromParams 从参数映射绑定到结构体
func bindFromParams(obj any, params map[string][]string) error {
	objValue := reflect.ValueOf(obj)
	if objValue.Kind() != reflect.Ptr {
		return fmt.Errorf("目标必须是指针类型")
	}

	objValue = objValue.Elem()
	if objValue.Kind() != reflect.Struct {
		return fmt.Errorf("目标必须是指向结构体的指针")
	}

	objType := objValue.Type()

	// 遍历结构体字段
	for i := 0; i < objValue.NumField(); i++ {
		field := objType.Field(i)
		fieldValue := objValue.Field(i)

		// 跳过不可设置的字段
		if !fieldValue.CanSet() {
			continue
		}

		// 获取字段标签
		jsonTag := field.Tag.Get("json")
		formTag := field.Tag.Get("form")

		// 确定参数名
		paramName := ""
		if jsonTag != "" && jsonTag != "-" {
			paramName = strings.Split(jsonTag, ",")[0]
		} else if formTag != "" && formTag != "-" {
			paramName = strings.Split(formTag, ",")[0]
		} else {
			paramName = field.Name
		}

		// 查找对应的参数值
		paramValues, exists := params[paramName]
		if !exists || len(paramValues) == 0 {
			continue
		}

		// 获取第一个值
		paramValue := paramValues[0]
		if paramValue == "" {
			continue
		}

		// 根据字段类型进行转换
		if err := setFieldValue(fieldValue, paramValue); err != nil {
			return fmt.Errorf("设置字段 %s 失败: %w", paramName, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Slice:
		// 处理切片类型
		newValue := reflect.MakeSlice(field.Type(), 1, 1)

		// 创建元素并设置值
		element := newValue.Index(0)
		if err := setFieldValue(element, value); err != nil {
			return err
		}
		field.Set(newValue)
	default:
		return fmt.Errorf("不支持的字段类型: %v", field.Kind())
	}

	return nil
}

// GetQueryParam 获取查询参数
func GetQueryParam(c *gin.Context, key string, defaultValue ...string) string {
	value := c.Query(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// GetPostParam 获取POST参数
func GetPostParam(c *gin.Context, key string, defaultValue ...string) string {
	value := c.PostForm(key)
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

// GetPathParam 获取路径参数
func GetPathParam(c *gin.Context, key string) string {
	return c.Param(key)
}

package crud

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// 定义上下文别名，保持原代码风格
type ctx = context.Context

// IDao -------------------------- 核心接口定义 --------------------------
// IDao GORM 版本的Dao接口，提供GORM DB实例和表相关信息
type IDao interface {
	DB() *gorm.DB                                                                          // 返回GORM的DB实例
	Table() string                                                                         // 返回表名
	PrimaryKey() string                                                                    // 返回主键字段名（如：id）
	Ctx(ctx context.Context) *gorm.DB                                                      // 绑定上下文的DB实例
	Transaction(ctx context.Context, f func(ctx context.Context, tx *gorm.DB) error) error // 事务方法
}

// Paginate -------------------------- 分页结构体定义 --------------------------
// Paginate 分页参数结构体（补全原代码中缺失的定义，保持功能完整）
type Paginate struct {
	Page  int // 页码（从1开始）
	Limit int // 每页条数
}

// -------------------------- 全局常量 --------------------------
// 分页相关字段，用于清理请求参数
var pageInfo = []string{
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

// Crud -------------------------- 泛型CURD核心结构体 --------------------------
// Crud GORM 版本的泛型CURD封装，R为对应的模型结构体
type Crud[R any] struct {
	Dao IDao
}

// -------------------------- 工具方法：字段名转换（保持原代码的命名风格转换） --------------------------
// caseConvert 字段名风格转换（下划线 <-> 小驼峰）
func caseConvert(key string, toSnake bool) string {
	if toSnake {
		// 驼峰转下划线（参考GORM的命名策略）
		return schema.NamingStrategy{}.ColumnName("", key)
	}
	// 下划线转小驼峰
	var result strings.Builder
	upperNext := false
	for i, c := range key {
		if c == '_' && i < len(key)-1 {
			upperNext = true
			continue
		}
		if upperNext {
			result.WriteRune(rune(strings.ToUpper(string(c))[0]))
			upperNext = false
		} else {
			result.WriteRune(c)
		}
	}
	return result.String()
}

// BuildWhere -------------------------- 原BuildWhere对应实现：构建查询条件map --------------------------
func (c Crud[R]) BuildWhere(req any, changeWhere any, subWhere any, removeFields []string, isSnake ...bool) map[string]any {
	// 默认使用小写下划线方式
	toSnake := true
	if len(isSnake) > 0 && !isSnake[0] {
		toSnake = false
	}

	// 1. 转换req为map并清理无效数据
	reqMap := convToMap(req)
	cleanedReq := make(map[string]any)

	for k, v := range reqMap {
		// 清理空值
		if isEmpty(v) {
			continue
		}
		// 清理分页字段
		if strInArray(pageInfo, k) {
			continue
		}
		// 清理指定移除字段
		if len(removeFields) > 0 && strInArray(removeFields, k) {
			continue
		}
		// 转换字段名风格并存入
		cleanedReq[caseConvert(k, toSnake)] = v
	}

	// 2. 处理changeWhere（修改查询操作符，如：eq -> gt）
	if changeWhere != nil {
		changeMap := convToMap(changeWhere)
		for k, v := range changeMap {
			// 跳过不存在于cleanedReq的字段
			if _, exists := cleanedReq[k]; !exists {
				continue
			}
			// 跳过指定移除的字段
			if len(removeFields) > 0 && strInArray(removeFields, k) {
				continue
			}

			vMap := convToMap(v)
			value, hasValue := vMap["value"]
			op, hasOp := vMap["op"]

			if hasValue {
				// 存在操作符则重构字段名（GORM支持 "字段名 >" 这种格式作为where key）
				if hasOp && op != "" {
					newKey := fmt.Sprintf("%s %s", k, op)
					delete(cleanedReq, k)
					cleanedReq[newKey] = value
				} else {
					cleanedReq[k] = value
				}
			}
		}
	}

	// 3. 字段名风格最终转换（确保一致性）
	resultMap := make(map[string]any)
	for k, v := range cleanedReq {
		// 拆分字段名和操作符
		parts := strings.SplitN(k, " ", 2)
		fieldName := parts[0]
		opStr := ""
		if len(parts) == 2 {
			opStr = parts[1]
		}

		// 转换字段名风格
		convertedField := caseConvert(fieldName, toSnake)

		// 重构带操作符的key
		if opStr != "" {
			resultMap[fmt.Sprintf("%s %s", convertedField, opStr)] = v
		} else {
			resultMap[convertedField] = v
		}
	}

	// 4. 合并subWhere附加条件
	if subWhere != nil {
		subMap := convToMap(subWhere)
		for k, v := range subMap {
			resultMap[caseConvert(k, toSnake)] = v
		}
	}

	return resultMap
}

// BuildMap -------------------------- 原BuildMap对应实现：构建变更条件map --------------------------
func (c Crud[R]) BuildMap(op string, value any, field ...string) map[string]any {
	res := map[string]any{
		"op":    op,
		"field": "",
		"value": value,
	}
	if len(field) > 0 {
		res["field"] = field[0]
	}
	return res
}

// ClearField -------------------------- 原ClearField对应实现：清理请求参数并返回有效map --------------------------
func (c Crud[R]) ClearField(req any, delField []string, subField ...map[string]any) map[string]any {
	reqMap := convToMap(req)
	resultMap := make(map[string]any)

	// 过滤无效数据和指定删除字段
	for k, v := range reqMap {
		if isEmpty(v) {
			continue
		}
		if strInArray(pageInfo, k) {
			continue
		}
		if len(delField) > 0 && strInArray(delField, k) {
			continue
		}
		resultMap[k] = v
	}

	// 合并附加字段
	if len(subField) > 0 && subField[0] != nil {
		for k, v := range subField[0] {
			resultMap[k] = v
		}
	}

	return resultMap
}

// ClearFieldPage -------------------------- 原ClearFieldPage对应实现：清理参数+分页查询 --------------------------
func (c Crud[R]) ClearFieldPage(ctx ctx, req any, delField []string, where any, page *Paginate, order any, with map[string]func(db *gorm.DB) *gorm.DB) (items []*R, total int64, err error) {
	// 1. 清理请求参数
	filterMap := c.ClearField(req, delField)

	// 2. 初始化GORM查询
	db := c.Dao.Ctx(ctx)

	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}

	// 3. 构建查询条件
	db = db.Model(new(R)).Where(filterMap)
	if where != nil {
		db = db.Where(where)
	}

	// 4. 排序
	if order != nil {
		db = db.Order(order)
	}

	// 5. 统计总数
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 6. 分页查询
	if page != nil && page.Limit > 0 {
		offset := (page.Page - 1) * page.Limit
		db = db.Offset(offset).Limit(page.Limit)
	}

	// 7. 执行查询
	err = db.Find(&items).Error
	return
}

// ClearFieldList -------------------------- 原ClearFieldList对应实现：清理参数+列表查询（不分页） --------------------------
func (c Crud[R]) ClearFieldList(ctx ctx, req any, delField []string, where any, order any, with map[string]func(db *gorm.DB) *gorm.DB) (items []*R, err error) {
	filterMap := c.ClearField(req, delField)
	db := c.Dao.Ctx(ctx).Model(new(R))

	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}
	if where != nil {
		db = db.Where(where)
	}
	if order != nil {
		db = db.Order(order)
	}

	err = db.Where(filterMap).Find(&items).Error
	return
}

// ClearFieldOne -------------------------- 原ClearFieldOne对应实现：清理参数+单条查询 --------------------------
func (c Crud[R]) ClearFieldOne(ctx ctx, req any, delField []string, where any, order any, with map[string]func(db *gorm.DB) *gorm.DB) (item *R, err error) {
	item = new(R)
	filterMap := c.ClearField(req, delField)
	db := c.Dao.Ctx(ctx).Model(item)

	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}
	if where != nil {
		db = db.Where(where)
	}
	if order != nil {
		db = db.Order(order)
	}

	err = db.Where(filterMap).First(item).Error
	// 处理记录不存在的情况（GORM会返回ErrRecordNotFound）
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic("未找到数据")
	}
	return
}

// Value -------------------------- 原Value对应实现：查询单个字段值 --------------------------
func (c Crud[R]) Value(ctx ctx, where any, field any) (interface{}, error) {
	var result interface{}
	db := c.Dao.Ctx(ctx).Model(new(R)).Where(where)

	// 处理字段参数
	if field != nil {
		fieldStr, ok := field.(string)
		if !ok || fieldStr == "" {
			fieldStr = "*"
		}
		db = db.Select(fieldStr)
	} else {
		db = db.Select("*")
	}

	// 执行查询（取第一条记录的指定字段）
	err := db.First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic("未找到数据")
	}
	return result, err
}

// DeletePri -------------------------- 原DeletePri对应实现：按主键删除 --------------------------
func (c Crud[R]) DeletePri(ctx ctx, primaryKey any) error {
	db := c.Dao.Ctx(ctx).Model(new(R))
	// 按主键字段构建查询
	pk := c.Dao.PrimaryKey()
	if pk == "" {
		panic("主键字段未配置")
	}
	return db.Where(fmt.Sprintf("%s = ?", pk), primaryKey).Delete(new(R)).Error
}

// DeleteWhere -------------------------- 原DeleteWhere对应实现：按条件删除 --------------------------
func (c Crud[R]) DeleteWhere(ctx ctx, where any) error {
	return c.Dao.Ctx(ctx).Model(new(R)).Where(where).Delete(new(R)).Error
}

// Sum -------------------------- 原Sum对应实现：字段求和 --------------------------
func (c Crud[R]) Sum(ctx ctx, where any, field string) float64 {
	var sum float64
	if field == "" {
		panic("求和字段不能为空")
	}

	err := c.Dao.Ctx(ctx).Model(new(R)).Where(where).Select(fmt.Sprintf("SUM(%s) as sum", field)).Scan(&sum).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic("未找到数据")
	}
	return sum
}

// ArrayField -------------------------- 原ArrayField对应实现：查询指定字段数组 --------------------------
func (c Crud[R]) ArrayField(ctx ctx, where any, field any) []interface{} {
	var result []interface{}
	db := c.Dao.Ctx(ctx).Model(new(R)).Where(where)

	// 处理字段参数
	if field != nil {
		fieldStr, ok := field.(string)
		if !ok || fieldStr == "" {
			fieldStr = "*"
		}
		db = db.Select(fieldStr)
	} else {
		db = db.Select("*")
	}

	// 执行查询
	err := db.Find(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic("未找到数据")
	}
	return result
}

// FindPri -------------------------- 原FindPri对应实现：按主键查询单条记录 --------------------------
func (c Crud[R]) FindPri(ctx ctx, primaryKey any, with map[string]func(db *gorm.DB) *gorm.DB) (model *R) {
	model = new(R)
	db := c.Dao.Ctx(ctx).Model(model)
	pk := c.Dao.PrimaryKey()

	if pk == "" {
		panic("主键字段未配置")
	}
	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}

	// 按主键查询
	err := db.Where(fmt.Sprintf("%s = ?", pk), primaryKey).First(model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic("未找到数据")
	}
	return
}

// -------------------------- 原First对应实现：按条件查询第一条记录 --------------------------
func (c Crud[R]) First(ctx ctx, where any, order any, with map[string]func(db *gorm.DB) *gorm.DB) (model *R) {
	model = new(R)
	db := c.Dao.Ctx(ctx).Model(model)

	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}
	if where != nil {
		db = db.Where(where)
	}
	if order != nil {
		db = db.Order(order)
	}

	err := db.First(model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic("未找到数据")
	}
	return
}

// -------------------------- 原Exists对应实现：判断记录是否存在 --------------------------
func (c Crud[R]) Exists(ctx ctx, where any) (exists bool) {
	var count int64
	err := c.Dao.Ctx(ctx).Model(new(R)).Where(where).Count(&count).Error
	if err != nil {
		panic(fmt.Sprintf("Exists查询错误: %v", err))
	}
	return count > 0
}

// -------------------------- 原All对应实现：查询所有符合条件的记录 --------------------------
func (c Crud[R]) All(ctx ctx, where any, order any, with map[string]func(db *gorm.DB) *gorm.DB) (items []*R) {
	db := c.Dao.Ctx(ctx).Model(new(R))

	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}
	if where != nil {
		db = db.Where(where)
	}
	if order != nil {
		db = db.Order(order)
	}

	err := db.Find(&items).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic(fmt.Sprintf("All查询错误: %v", err))
	}
	return
}

// -------------------------- 原Count对应实现：统计记录总数 --------------------------
func (c Crud[R]) Count(ctx ctx, where any) (count int64) {
	err := c.Dao.Ctx(ctx).Model(new(R)).Where(where).Count(&count).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		panic(fmt.Sprintf("Count查询错误: %v", err))
	}
	return
}

// -------------------------- 原Save对应实现：新增/更新记录（对应GORM的Save） --------------------------
func (c Crud[R]) Save(ctx ctx, data any) {
	err := c.Dao.Ctx(ctx).Model(new(R)).Create(data).Error
	if err != nil {
		panic(fmt.Sprintf("Save保存错误: %v", err))
	}
}

// -------------------------- 原Update对应实现：按条件更新记录 --------------------------
func (c Crud[R]) Update(ctx ctx, where any, data any) (count int64) {
	result := c.Dao.Ctx(ctx).Model(new(R)).Where(where).Updates(data)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		panic(fmt.Sprintf("Update更新错误: %v", result.Error.Error()))
	}
	return result.RowsAffected
}

// -------------------------- 原UpdatePri对应实现：按主键更新记录 --------------------------
func (c Crud[R]) UpdatePri(ctx ctx, primaryKey any, data any) (count int64) {
	db := c.Dao.Ctx(ctx).Model(new(R))
	pk := c.Dao.PrimaryKey()

	if pk == "" {
		panic("主键字段未配置")
	}

	result := db.Where(fmt.Sprintf("%s = ?", pk), primaryKey).Updates(data)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		panic(fmt.Sprintf("UpdatePri更新错误: %v", result.Error.Error()))
	}
	return result.RowsAffected
}

// -------------------------- 原Paginate对应实现：分页查询 --------------------------
func (c Crud[R]) Paginate(ctx context.Context, where any, p Paginate, with map[string]func(db *gorm.DB) *gorm.DB, order any) (items []*R, total int64) {
	db := c.Dao.Ctx(ctx).Model(new(R))

	// 1. 构建查询条件
	if where != nil {
		db = db.Where(where)
	}

	// 2. 统计总数
	if err := db.Count(&total).Error; err != nil {
		panic(fmt.Sprintf("Paginate查询错误: %v", err))
	}

	// 3. 关联查询
	if with != nil {
		for k, v := range with {
			db = db.Preload(k, v)
		}
	}

	// 4. 排序
	if order != nil {
		db = db.Order(order)
	}

	// 5. 分页（offset = (页码-1)*每页条数）
	if p.Limit > 0 {
		offset := (p.Page - 1) * p.Limit
		db = db.Offset(offset).Limit(p.Limit)
	}

	// 6. 执行查询
	err := db.Find(&items).Error
	if err != nil || errors.Is(err, gorm.ErrRecordNotFound) {
		panic(fmt.Sprintf("Paginate查询错误: %v", err))
	}
	return
}

// -------------------------- 内部辅助工具函数 --------------------------
// convToMap 将任意类型转换为map[string]any（简化版，适配常见场景）
func convToMap(v any) map[string]any {
	if v == nil {
		return make(map[string]any)
	}

	val := reflect.ValueOf(v)
	// 处理指针类型
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 只处理结构体和map类型
	if val.Kind() != reflect.Struct && val.Kind() != reflect.Map {
		return make(map[string]any)
	}

	result := make(map[string]any)

	if val.Kind() == reflect.Map {
		// 处理map类型
		for _, key := range val.MapKeys() {
			keyStr, ok := key.Interface().(string)
			if !ok {
				continue
			}
			result[keyStr] = val.MapIndex(key).Interface()
		}
	} else {
		// 处理结构体类型
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := val.Field(i)

			// 获取json标签作为key（优先），否则用字段名
			jsonTag := field.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				jsonTag = field.Name
			} else {
				// 分割json标签（忽略omitempty等选项）
				jsonTag = strings.Split(jsonTag, ",")[0]
			}

			result[jsonTag] = fieldVal.Interface()
		}
	}

	return result
}

// isEmpty 判断值是否为空
func isEmpty(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.String:
		return val.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return false
	}
}

// strInArray 判断字符串是否在数组中
func strInArray(arr []string, str string) bool {
	for _, v := range arr {
		if strings.EqualFold(v, str) { // 忽略大小写比较
			return true
		}
	}
	return false
}

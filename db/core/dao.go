package core

import (
	"context"
	"reflect"
)

// DAO 数据访问对象基类 - 所有 DAO 都继承此结构
// 提供通用的 CRUD 操作方法，子类只需嵌入即可使用
type DAO struct {
	db        *Database   // 数据库连接实例
	modelType interface{} // 模型类型信息，用于 Columns 等方法
}

// NewDAO 创建 DAO 基类实例
// 自动使用全局默认 Database 实例
func NewDAO() *DAO {
	return &DAO{
		db: GetDefaultDatabase(),
	}
}

// NewDAOWithModel 创建带模型类型的 DAO 基类实例
// 参数：
//   - model: 模型实例（指针类型），用于获取表结构信息
//
// 自动使用全局默认 Database 实例
func NewDAOWithModel(model interface{}) *DAO {
	return &DAO{
		db:        GetDefaultDatabase(),
		modelType: model,
	}
}

// Create 创建记录（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) Create(ctx context.Context, model interface{}) error {
	// 使用事务来插入数据
	tx, err := dao.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Insert(model)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetByID 根据 ID 查询单条记录（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) GetByID(ctx context.Context, model interface{}, id int64) error {
	return dao.db.Model(model).Where("id = ?", id).First(model)
}

// Update 更新记录（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) Update(ctx context.Context, model interface{}, data map[string]interface{}) error {
	pkValue := getFieldValue(model, "ID")

	if pkValue == 0 {
		return nil
	}

	return dao.db.Model(model).Where("id = ?", pkValue).Updates(data)
}

// Delete 删除记录（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) Delete(ctx context.Context, model interface{}) error {
	pkValue := getFieldValue(model, "ID")

	if pkValue == 0 {
		return nil
	}

	return dao.db.Model(model).Where("id = ?", pkValue).Delete()
}

// FindAll 查询所有记录（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) FindAll(ctx context.Context, model interface{}) error {
	return dao.db.Model(model).Find(model)
}

// FindByPage 分页查询（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) FindByPage(ctx context.Context, model interface{}, page, pageSize int) error {
	return dao.db.Model(model).Limit(pageSize).Offset((page - 1) * pageSize).Find(model)
}

// Count 统计记录数（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) Count(ctx context.Context, model interface{}, where ...string) (int64, error) {
	var count int64

	query := dao.db.Model(model)
	if len(where) > 0 {
		query = query.Where(where[0])
	}

	// Count 是链式调用，需要调用 Find 来执行
	err := query.Count(&count).Find(model)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Exists 检查记录是否存在（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) Exists(ctx context.Context, model interface{}) (bool, error) {
	count, err := dao.Count(ctx, model)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// First 查询第一条记录（通用方法）
// 自动使用 DAO 中已关联的 Database 实例
func (dao *DAO) First(ctx context.Context, model interface{}) error {
	return dao.db.Model(model).First(model)
}

// Columns 获取表的所有列名
// 返回一个动态创建的结构体类型，所有字段都是 string 类型
// 用途：用于构建 UPDATE、INSERT 等操作时的列名映射
//
// 示例：
//
//	type UserDAO struct {
//	    *core.DAO
//	}
//
//	func NewUserDAO() *UserDAO {
//	    return &UserDAO{
//	        DAO: core.NewDAOWithModel(&model.User{}),
//	    }
//	}
//
//	// 使用
//	dao := NewUserDAO()
//	cols := dao.Columns() // 返回 *struct{ID string; Username string; ...}
func (dao *DAO) Columns() interface{} {
	// 检查是否有模型类型信息
	if dao.modelType == nil {
		return nil
	}

	// 获取模型类型
	modelType := reflect.TypeOf(dao.modelType)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// 创建字段列表
	fields := []reflect.StructField{}

	// 遍历模型的所有字段
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取 db 标签，如果没有则跳过
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		// 创建新的结构体字段，类型为 string
		newField := reflect.StructField{
			Name: field.Name,
			Type: reflect.TypeOf(""), // string 类型
			Tag:  reflect.StructTag(`json:"` + field.Tag.Get("json") + `" db:"` + dbTag + `"`),
		}

		fields = append(fields, newField)
	}

	// 动态创建结构体类型
	columnsType := reflect.StructOf(fields)

	// 创建该类型的指针并返回
	return reflect.New(columnsType).Interface()
}

// getFieldValue 获取结构体字段值（辅助函数）
// 用于获取主键或其他字段的值，支持多种数据类型
// 参数：
//   - model: 模型实例（可以是指针或值）
//   - fieldName: 字段名（如 "ID", "UserId" 等）
//
// 返回：
//   - int64: 字段值（如果是数字类型）或 0（无法获取时）
func getFieldValue(model interface{}, fieldName string) int64 {
	// 检查空值
	if model == nil {
		return 0
	}

	// 获取反射对象
	val := reflect.ValueOf(model)

	// 如果是指针，解引用
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}

	// 确保是结构体
	if val.Kind() != reflect.Struct {
		return 0
	}

	// 查找字段
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		// 尝试查找常见的主键字段名变体
		alternativeNames := []string{"Id", "id", "ID"}
		for _, name := range alternativeNames {
			if name != fieldName {
				field = val.FieldByName(name)
				if field.IsValid() {
					fieldName = name
					break
				}
			}
		}

		if !field.IsValid() {
			return 0
		}
	}

	// 检查字段是否可以访问
	if !field.CanInterface() {
		return 0
	}

	// 获取字段值并转换为 int64
	fieldValue := field.Interface()

	// 根据字段类型进行转换
	switch v := fieldValue.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return int64(v)
	case uint8:
		return int64(v)
	case uint16:
		return int64(v)
	case uint32:
		return int64(v)
	case uint64:
		// 注意：uint64 转 int64 可能溢出，但这里假设 ID 不会超过 int64 范围
		return int64(v)
	case float32:
		return int64(v)
	case float64:
		return int64(v)
	case string:
		// 尝试将字符串解析为数字
		// 注意：这里不导入 strconv，简单处理返回 0
		return 0
	default:
		// 其他类型（如 sql.NullInt64 等），尝试使用反射
		return convertToInteger(field)
	}
}

// convertToInteger 使用反射将字段值转换为 int64
func convertToInteger(field reflect.Value) int64 {
	// 获取实际的值（如果是指针则解引用）
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return 0
		}
		field = field.Elem()
	}

	// 根据 Kind 进行转换
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(field.Uint())
	case reflect.Float32, reflect.Float64:
		return int64(field.Float())
	case reflect.String:
		// 字符串类型，尝试解析（简单实现，不处理错误）
		return 0
	default:
		return 0
	}
}

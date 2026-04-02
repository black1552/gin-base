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
func NewDAO(db *Database) *DAO {
	return &DAO{db: db}
}

// NewDAOWithModel 创建带模型类型的 DAO 基类实例
// 参数：
//   - db: 数据库连接实例
//   - model: 模型实例（指针类型），用于获取表结构信息
func NewDAOWithModel(db *Database, model interface{}) *DAO {
	return &DAO{
		db:        db,
		modelType: model,
	}
}

// Create 创建记录（通用方法）
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
func (dao *DAO) GetByID(ctx context.Context, model interface{}, id int64) error {
	return dao.db.Model(model).Where("id = ?", id).First(model)
}

// Update 更新记录（通用方法）
func (dao *DAO) Update(ctx context.Context, model interface{}, data map[string]interface{}) error {
	pkValue := getFieldValue(model, "ID")

	if pkValue == 0 {
		return nil
	}

	return dao.db.Model(model).Where("id = ?", pkValue).Updates(data)
}

// Delete 删除记录（通用方法）
func (dao *DAO) Delete(ctx context.Context, model interface{}) error {
	pkValue := getFieldValue(model, "ID")

	if pkValue == 0 {
		return nil
	}

	return dao.db.Model(model).Where("id = ?", pkValue).Delete()
}

// FindAll 查询所有记录（通用方法）
func (dao *DAO) FindAll(ctx context.Context, model interface{}) error {
	return dao.db.Model(model).Find(model)
}

// FindByPage 分页查询（通用方法）
func (dao *DAO) FindByPage(ctx context.Context, model interface{}, page, pageSize int) error {
	return dao.db.Model(model).Limit(pageSize).Offset((page - 1) * pageSize).Find(model)
}

// Count 统计记录数（通用方法）
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
func (dao *DAO) Exists(ctx context.Context, model interface{}) (bool, error) {
	count, err := dao.Count(ctx, model)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// First 查询第一条记录（通用方法）
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
//	func NewUserDAO(db *core.Database) *UserDAO {
//	    return &UserDAO{
//	        DAO: core.NewDAOWithModel(db, &model.User{}),
//	    }
//	}
//
//	// 使用
//	dao := NewUserDAO(db)
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
func getFieldValue(model interface{}, fieldName string) int64 {
	// TODO: 使用反射获取字段值
	// 这里是简化实现，实际需要根据情况完善
	return 0
}

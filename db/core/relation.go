package core

import (
	"fmt"
	"reflect"
	"strings"
)

// RelationType 关联类型
type RelationType int

const (
	HasOne     RelationType = iota // 一对一
	HasMany                        // 一对多
	BelongsTo                      // 多对一
	ManyToMany                     // 多对多
)

// RelationInfo 关联信息
type RelationInfo struct {
	Type       RelationType // 关联类型
	Field      string       // 字段名
	Model      interface{}  // 关联的模型
	FK         string       // 外键
	PK         string       // 主键
	JoinTable  string       // 中间表（多对多）
	JoinFK     string       // 中间表外键
	JoinJoinFK string       // 中间表关联外键
}

// RelationLoader 关联加载器 - 处理模型关联的预加载
type RelationLoader struct {
	db *Database
}

// NewRelationLoader 创建关联加载器实例
func NewRelationLoader(db *Database) *RelationLoader {
	return &RelationLoader{db: db}
}

// Preload 预加载关联数据
func (rl *RelationLoader) Preload(models interface{}, relation string, conditions ...interface{}) error {
	// 获取反射对象
	modelsVal := reflect.ValueOf(models)
	if modelsVal.Kind() != reflect.Ptr {
		return fmt.Errorf("models 必须是指针类型")
	}

	elem := modelsVal.Elem()
	if elem.Kind() != reflect.Slice {
		return fmt.Errorf("models 必须是指向 Slice 的指针")
	}

	if elem.Len() == 0 {
		return nil // 空 Slice，无需加载
	}

	// 解析关联关系
	relationInfo, err := rl.parseRelation(elem.Index(0).Interface(), relation)
	if err != nil {
		return fmt.Errorf("解析关联失败：%w", err)
	}

	// 根据关联类型加载数据
	switch relationInfo.Type {
	case HasOne:
		return rl.loadHasOne(elem, relationInfo)
	case HasMany:
		return rl.loadHasMany(elem, relationInfo)
	case BelongsTo:
		return rl.loadBelongsTo(elem, relationInfo)
	case ManyToMany:
		return rl.loadManyToMany(elem, relationInfo)
	default:
		return fmt.Errorf("不支持的关联类型：%v", relationInfo.Type)
	}
}

// parseRelation 解析关联关系
func (rl *RelationLoader) parseRelation(model interface{}, relation string) (*RelationInfo, error) {
	// TODO: 从结构体标签中解析关联信息
	// 示例：
	// type Order struct {
	//     User   User   `gorm:"ForeignKey:user_id;References:id"`
	//     Items  []Item `gorm:"ForeignKey:order_id;References:id"`
	// }

	// 这里提供简化的实现
	return &RelationInfo{
		Type:  HasOne, // 默认假设为一对一
		Field: relation,
	}, nil
}

// loadHasOne 加载一对一关联
func (rl *RelationLoader) loadHasOne(models reflect.Value, relation *RelationInfo) error {
	// 收集所有主键值
	pkValues := make([]interface{}, 0, models.Len())
	for i := 0; i < models.Len(); i++ {
		model := models.Index(i).Interface()
		pk := rl.getFieldValue(model, "ID")
		if pk != nil {
			pkValues = append(pkValues, pk)
		}
	}

	if len(pkValues) == 0 {
		return nil
	}

	// 查询关联数据
	query := rl.db.Model(relation.Model)
	query.Where(fmt.Sprintf("%s IN ?", relation.FK), pkValues)

	// TODO: 执行查询并映射到模型

	return nil
}

// loadHasMany 加载一对多关联
func (rl *RelationLoader) loadHasMany(models reflect.Value, relation *RelationInfo) error {
	// 类似 HasOne，但结果需要映射到 Slice
	return rl.loadHasOne(models, relation)
}

// loadBelongsTo 加载多对一关联
func (rl *RelationLoader) loadBelongsTo(models reflect.Value, relation *RelationInfo) error {
	// 收集所有外键值
	fkValues := make([]interface{}, 0, models.Len())
	for i := 0; i < models.Len(); i++ {
		model := models.Index(i).Interface()
		fk := rl.getFieldValue(model, relation.FK)
		if fk != nil {
			fkValues = append(fkValues, fk)
		}
	}

	if len(fkValues) == 0 {
		return nil
	}

	// 查询关联数据
	query := rl.db.Model(relation.Model)
	query.Where(fmt.Sprintf("id IN ?"), fkValues)

	// TODO: 执行查询并映射到模型

	return nil
}

// loadManyToMany 加载多对多关联
func (rl *RelationLoader) loadManyToMany(models reflect.Value, relation *RelationInfo) error {
	// 多对多需要通过中间表查询
	// SELECT * FROM table WHERE id IN (
	//     SELECT join_fk FROM join_table WHERE fk IN (pk_values)
	// )

	return fmt.Errorf("多对多关联暂未实现")
}

// getFieldValue 获取字段的值
func (rl *RelationLoader) getFieldValue(model interface{}, fieldName string) interface{} {
	val := reflect.ValueOf(model)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	field := val.FieldByName(fieldName)
	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}

	return nil
}

// getRelationTags 从结构体字段提取关联标签信息
func getRelationTags(structType reflect.Type, fieldName string) map[string]string {
	tags := make(map[string]string)

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Name == fieldName {
			gormTag := field.Tag.Get("gorm")
			if gormTag != "" {
				// 解析 GORM 风格的标签
				parts := strings.Split(gormTag, ";")
				for _, part := range parts {
					kv := strings.Split(part, ":")
					if len(kv) == 2 {
						tags[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}
			break
		}
	}

	return tags
}

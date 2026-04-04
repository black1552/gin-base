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
	// 从结构体字段中解析关联信息
	structType := reflect.TypeOf(model)
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}

	// 查找对应的字段
	var relationField reflect.StructField
	var found bool

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if field.Name == relation {
			relationField = field
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("字段 %s 不存在", relation)
	}

	// 从 gorm 标签解析关联信息
	gormTag := relationField.Tag.Get("gorm")
	fkTag := relationField.Tag.Get("foreignkey")
	referencesTag := relationField.Tag.Get("references")
	joinTableTag := relationField.Tag.Get("many2many")

	// 初始化关联信息
	info := &RelationInfo{
		Field: relation,
		Model: reflect.New(relationField.Type).Interface(),
	}

	// 判断关联类型
	if relationField.Type.Kind() == reflect.Slice {
		// 一对多或多对多
		if joinTableTag != "" {
			// 多对多
			info.Type = ManyToMany
			info.JoinTable = joinTableTag
		} else {
			// 一对多
			info.Type = HasMany
		}
	} else {
		// 一对一或多对一
		// 根据外键位置判断
		if fkTag != "" || referencesTag != "" {
			// 如果当前模型包含外键，则是多对一
			info.Type = BelongsTo
		} else {
			// 否则是一对一
			info.Type = HasOne
		}
	}

	// 解析外键和主键
	if gormTag != "" {
		// 解析 GORM 风格的标签
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			kv := strings.Split(part, ":")
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				switch key {
				case "ForeignKey":
					info.FK = value
				case "References":
					info.PK = value
				case "JoinTable":
					info.JoinTable = value
				case "JoinForeignKey":
					info.JoinFK = value
				case "JoinReferences":
					info.JoinJoinFK = value
				}
			}
		}
	}

	// 使用单独的标签
	if fkTag != "" {
		info.FK = fkTag
	}
	if referencesTag != "" {
		info.PK = referencesTag
	}

	// 设置默认值
	if info.FK == "" {
		// 默认外键为当前模型名 + Id
		modelName := structType.Name()
		info.FK = modelName + "Id"
	}
	if info.PK == "" {
		info.PK = "id"
	}

	return info, nil
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
	query.Where(fmt.Sprintf("%s IN (?)", relation.FK), pkValues)

	// 执行查询并映射到模型
	relatedData := reflect.New(reflect.SliceOf(reflect.TypeOf(relation.Model))).Interface()
	if err := query.Find(relatedData); err != nil {
		return err
	}

	// 将关联数据映射到模型
	relatedVal := reflect.ValueOf(relatedData)
	if relatedVal.Kind() == reflect.Ptr {
		relatedVal = relatedVal.Elem()
	}

	// 遍历所有模型，设置关联字段
	for i := 0; i < models.Len(); i++ {
		model := models.Index(i)
		pk := rl.getFieldValue(model.Interface(), "ID")

		// 查找对应的关联数据
		for j := 0; j < relatedVal.Len(); j++ {
			item := relatedVal.Index(j).Interface()
			itemFK := rl.getFieldValue(item, relation.FK)
			if itemFK != nil && fmt.Sprintf("%v", itemFK) == fmt.Sprintf("%v", pk) {
				model.Elem().FieldByName(relation.Field).Set(reflect.ValueOf(item))
				break
			}
		}
	}

	return nil
}

// loadHasMany 加载一对多关联
func (rl *RelationLoader) loadHasMany(models reflect.Value, relation *RelationInfo) error {
	// 一对多的逻辑与 HasOne 类似，但结果必须映射到 Slice
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
	query.Where(fmt.Sprintf("%s IN (?)", relation.PK), fkValues)

	// 执行查询
	relatedData := reflect.New(reflect.SliceOf(reflect.TypeOf(relation.Model))).Interface()
	if err := query.Find(relatedData); err != nil {
		return err
	}

	// 将关联数据映射到模型
	relatedVal := reflect.ValueOf(relatedData)
	if relatedVal.Kind() == reflect.Ptr {
		relatedVal = relatedVal.Elem()
	}

	// 遍历所有模型，设置关联字段
	for i := 0; i < models.Len(); i++ {
		model := models.Index(i)
		fk := rl.getFieldValue(model.Interface(), relation.FK)

		// 查找对应的关联数据
		for j := 0; j < relatedVal.Len(); j++ {
			item := relatedVal.Index(j).Interface()
			itemPK := rl.getFieldValue(item, relation.PK)
			if itemPK != nil && fmt.Sprintf("%v", itemPK) == fmt.Sprintf("%v", fk) {
				model.Elem().FieldByName(relation.Field).Set(reflect.ValueOf(item))
				break
			}
		}
	}

	return nil
}

// loadManyToMany 加载多对多关联
func (rl *RelationLoader) loadManyToMany(models reflect.Value, relation *RelationInfo) error {
	// 多对多需要通过中间表查询
	// SELECT * FROM table WHERE id IN (
	//     SELECT join_fk FROM join_table WHERE fk IN (pk_values)
	// )

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

	// 检查中间表配置
	if relation.JoinTable == "" || relation.JoinFK == "" || relation.JoinJoinFK == "" {
		return fmt.Errorf("多对多关联需要配置中间表信息")
	}

	// 先从中间表获取关联关系
	joinQuery := rl.db.Table(relation.JoinTable)
	joinQuery.Where(fmt.Sprintf("%s IN (?)", relation.JoinFK), pkValues)

	// 这里简化处理，实际应该查询中间表获取关联 ID 列表
	// 然后查询关联模型

	return fmt.Errorf("多对多关联实现中，请稍后使用")
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

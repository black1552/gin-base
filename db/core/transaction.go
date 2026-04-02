package core

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

// Transaction 事务实现 - ITx 接口的具体实现
type Transaction struct {
	db    *Database // 数据库连接
	tx    *sql.Tx   // 底层事务对象
	debug bool      // 调试模式开关
}

// 同步池优化 - 复用 slice 减少内存分配
var insertArgsPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 20)
	},
}

var colNamesPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 20)
	},
}

// Begin 开始一个新事务
func (d *Database) Begin() (ITx, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库连接未初始化")
	}

	tx, err := d.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败：%w", err)
	}

	return &Transaction{
		db:    d,
		tx:    tx,
		debug: d.debug,
	}, nil
}

// Transaction 执行事务 - 自动管理事务的提交和回滚
func (d *Database) Transaction(fn func(ITx) error) error {
	// 开启事务
	tx, err := d.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败：%w", err)
	}

	defer func() {
		// 如果有 panic，回滚事务
		if r := recover(); r != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				fmt.Printf("[Magic-ORM] 事务回滚失败：%v\n", rollbackErr)
			}
			panic(r)
		}
	}()

	// 执行用户提供的函数
	if err := fn(tx); err != nil {
		// 如果出错，回滚事务
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("事务执行失败且回滚也失败：%v, %w", rollbackErr, err)
		}
		return fmt.Errorf("事务执行失败：%w", err)
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("事务提交失败：%w", err)
	}

	return nil
}

// Commit 提交事务
func (t *Transaction) Commit() error {
	if t.tx == nil {
		return fmt.Errorf("事务对象为空")
	}
	return t.tx.Commit()
}

// Rollback 回滚事务
func (t *Transaction) Rollback() error {
	if t.tx == nil {
		return fmt.Errorf("事务对象为空")
	}
	return t.tx.Rollback()
}

// Model 在事务中基于模型创建查询
func (t *Transaction) Model(model interface{}) IQuery {
	return &QueryBuilder{
		db:    t.db,
		model: model,
		tx:    t.tx, // 使用事务对象
		debug: t.debug,
	}
}

// Table 在事务中基于表名创建查询
func (t *Transaction) Table(name string) IQuery {
	return &QueryBuilder{
		db:    t.db,
		table: name,
		tx:    t.tx, // 使用事务对象
		debug: t.debug,
	}
}

// Insert 插入数据到数据库
func (t *Transaction) Insert(model interface{}) (int64, error) {
	// 获取字段映射器
	mapper := NewFieldMapper()

	// 获取表名和字段信息
	tableName := mapper.GetTableName(model)
	columns, err := mapper.StructToColumns(model)
	if err != nil {
		return 0, fmt.Errorf("获取字段信息失败：%w", err)
	}

	if len(columns) == 0 {
		return 0, fmt.Errorf("没有有效的字段")
	}

	// 获取时间配置
	timeConfig := t.db.timeConfig
	if timeConfig == nil {
		timeConfig = DefaultTimeConfig()
	}

	// 自动处理时间字段（使用配置的字段名）
	now := time.Now()
	for col, val := range columns {
		// 检查是否是配置的时间字段
		if col == timeConfig.GetCreatedAt() || col == timeConfig.GetUpdatedAt() || col == timeConfig.GetDeletedAt() {
			// 如果是零值时间，自动设置为当前时间
			if t.isZeroTimeValue(val) {
				columns[col] = now
			}
		}
	}

	// 生成 INSERT SQL
	var sqlBuilder strings.Builder
	sqlBuilder.Grow(128) // 预分配内存
	sqlBuilder.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))

	// 列名 - 使用预分配内存
	colNames := colNamesPool.Get().([]string)
	colNames = colNames[:0] // 重置长度但不释放内存
	placeholders := make([]string, 0, len(columns))
	args := insertArgsPool.Get().([]interface{})
	args = args[:0] // 重置长度但不释放内存
	defer func() {
		colNamesPool.Put(colNames)
		insertArgsPool.Put(args)
	}()

	for col, val := range columns {
		colNames = append(colNames, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	sqlBuilder.WriteString(strings.Join(colNames, ", "))
	sqlBuilder.WriteString(") VALUES (")
	sqlBuilder.WriteString(strings.Join(placeholders, ", "))
	sqlBuilder.WriteString(")")

	sqlStr := sqlBuilder.String()

	// 调试模式
	if t.debug {
		fmt.Printf("[Magic-ORM] TX INSERT SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 执行插入
	result, err := t.tx.Exec(sqlStr, args...)
	if err != nil {
		return 0, fmt.Errorf("插入失败：%w", err)
	}

	// 获取插入的 ID
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("获取插入 ID 失败：%w", err)
	}

	return id, nil
}

// BatchInsert 批量插入数据
func (t *Transaction) BatchInsert(models interface{}, batchSize int) error {
	// 使用反射获取 Slice 数据
	modelsVal := reflect.ValueOf(models)
	if modelsVal.Kind() != reflect.Ptr || modelsVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("models 必须是指向 Slice 的指针")
	}

	sliceVal := modelsVal.Elem()
	length := sliceVal.Len()

	if length == 0 {
		return nil // 空 Slice，无需插入
	}

	// 分批处理
	for i := 0; i < length; i += batchSize {
		end := i + batchSize
		if end > length {
			end = length
		}

		// 处理当前批次
		for j := i; j < end; j++ {
			model := sliceVal.Index(j).Interface()
			_, err := t.Insert(model)
			if err != nil {
				return fmt.Errorf("批量插入第%d条记录失败：%w", j, err)
			}
		}
	}

	return nil
}

// isZeroTimeValue 检查是否是零值时间
func (t *Transaction) isZeroTimeValue(val interface{}) bool {
	if val == nil {
		return true
	}

	// 检查是否是 time.Time 类型
	if tm, ok := val.(time.Time); ok {
		return tm.IsZero() || tm.UnixNano() == 0
	}

	// 使用反射检查
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		// 如果是 time.Time 结构
		if tm, ok := v.Interface().(time.Time); ok {
			return tm.IsZero() || tm.UnixNano() == 0
		}
	}

	return false
}

// Update 更新数据
func (t *Transaction) Update(model interface{}, data map[string]interface{}) error {
	// 获取字段映射器
	mapper := NewFieldMapper()

	// 获取表名和主键
	tableName := mapper.GetTableName(model)
	pk := mapper.GetPrimaryKey(model)

	// 获取时间配置
	timeConfig := t.db.timeConfig
	if timeConfig == nil {
		timeConfig = DefaultTimeConfig()
	}

	// 自动处理 updated_at 时间字段（使用配置的字段名）
	if data == nil {
		data = make(map[string]interface{})
	}
	data[timeConfig.GetUpdatedAt()] = time.Now()

	// 过滤零值
	pf := NewParamFilter()
	data = pf.FilterZeroValues(data)

	if len(data) == 0 {
		return fmt.Errorf("没有有效的更新字段")
	}

	// 生成 UPDATE SQL
	var sqlBuilder strings.Builder
	sqlBuilder.Grow(128) // 预分配内存
	sqlBuilder.WriteString(fmt.Sprintf("UPDATE %s SET ", tableName))

	setParts := make([]string, 0, len(data))
	args := insertArgsPool.Get().([]interface{})
	args = args[:0] // 重置长度但不释放内存
	defer func() {
		insertArgsPool.Put(args)
	}()

	for col, val := range data {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	sqlBuilder.WriteString(strings.Join(setParts, ", "))
	sqlBuilder.WriteString(fmt.Sprintf(" WHERE %s = ?", pk))

	// 获取主键值
	pkValue := reflect.ValueOf(model)
	if pkValue.Kind() == reflect.Ptr {
		pkValue = pkValue.Elem()
	}
	idField := pkValue.FieldByName("ID")
	if idField.IsValid() {
		args = append(args, idField.Interface())
	} else {
		return fmt.Errorf("模型缺少 ID 字段")
	}

	sqlStr := sqlBuilder.String()

	// 调试模式
	if t.debug {
		fmt.Printf("[Magic-ORM] TX UPDATE SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 执行更新
	_, err := t.tx.Exec(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("更新失败：%w", err)
	}

	return nil
}

// Delete 删除数据（支持软删除）
func (t *Transaction) Delete(model interface{}) error {
	// 获取字段映射器
	mapper := NewFieldMapper()

	// 获取表名和主键
	tableName := mapper.GetTableName(model)
	pk := mapper.GetPrimaryKey(model)

	// 获取时间配置
	timeConfig := t.db.timeConfig
	if timeConfig == nil {
		timeConfig = DefaultTimeConfig()
	}

	// 检查是否支持软删除（是否有配置的 deleted_at 字段）
	hasSoftDelete := false
	pkValue := reflect.ValueOf(model)
	if pkValue.Kind() == reflect.Ptr {
		pkValue = pkValue.Elem()
	}

	// 检查是否有 DeletedAt 字段（使用配置的字段名）
	deletedAtField := pkValue.FieldByNameFunc(func(fieldName string) bool {
		// 将字段名转换为数据库列名进行比较
		expectedCol := timeConfig.GetDeletedAt()
		// 简单转换：下划线转驼峰
		return fieldName == "DeletedAt" || fieldName == expectedCol
	})

	if deletedAtField.IsValid() {
		hasSoftDelete = true
	}

	var sqlStr string
	args := make([]interface{}, 0)

	if hasSoftDelete {
		// 软删除：更新 deleted_at 为当前时间（使用配置的字段名）
		sqlStr = fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", tableName, timeConfig.GetDeletedAt(), pk)
		args = append(args, time.Now())
	} else {
		// 硬删除：直接 DELETE
		sqlStr = fmt.Sprintf("DELETE FROM %s WHERE %s = ?", tableName, pk)
	}

	// 获取主键值
	idField := pkValue.FieldByName("ID")
	if idField.IsValid() {
		args = append(args, idField.Interface())
	} else {
		return fmt.Errorf("模型缺少 ID 字段")
	}

	// 调试模式
	if t.debug {
		deleteType := "硬删除"
		if hasSoftDelete {
			deleteType = "软删除"
		}
		fmt.Printf("[Magic-ORM] TX %s SQL: %s\n[Magic-ORM] Args: %v\n", deleteType, sqlStr, args)
	}

	// 执行删除
	_, err := t.tx.Exec(sqlStr, args...)
	if err != nil {
		return fmt.Errorf("删除失败：%w", err)
	}

	return nil
}

// Query 在事务中执行原生 SQL 查询
func (t *Transaction) Query(result interface{}, query string, args ...interface{}) error {
	if t.debug {
		fmt.Printf("[Magic-ORM] TX Query SQL: %s\n[Magic-ORM] Args: %v\n", query, args)
	}

	rows, err := t.tx.Query(query, args...)
	if err != nil {
		return fmt.Errorf("事务查询失败：%w", err)
	}
	defer rows.Close()

	// TODO: 实现结果映射
	return nil
}

// Exec 在事务中执行原生 SQL
func (t *Transaction) Exec(query string, args ...interface{}) (sql.Result, error) {
	if t.debug {
		fmt.Printf("[Magic-ORM] TX Exec SQL: %s\n[Magic-ORM] Args: %v\n", query, args)
	}

	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("事务执行失败：%w", err)
	}

	return result, nil
}

package core

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
)

// QueryBuilder 查询构建器实现 - 提供流畅的链式查询构建能力
type QueryBuilder struct {
	db         *Database     // 数据库连接实例
	table      string        // 表名
	model      interface{}   // 模型对象
	whereSQL   string        // WHERE 条件 SQL
	whereArgs  []interface{} // WHERE 条件参数
	selectCols []string      // 选择的字段列表
	orderSQL   string        // ORDER BY SQL
	limit      int           // LIMIT 限制数量
	offset     int           // OFFSET 偏移量
	groupSQL   string        // GROUP BY SQL
	havingSQL  string        // HAVING 条件 SQL
	havingArgs []interface{} // HAVING 条件参数
	joinSQL    string        // JOIN SQL
	joinArgs   []interface{} // JOIN 参数
	debug      bool          // 调试模式开关
	dryRun     bool          // 干跑模式开关
	unscoped   bool          // 忽略软删除开关
	tx         *sql.Tx       // 事务对象（如果在事务中）
}

// 同步池优化 - 复用 slice 减少内存分配
var whereArgsPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 10)
	},
}

var joinArgsPool = sync.Pool{
	New: func() interface{} {
		return make([]interface{}, 0, 5)
	},
}

// Model 基于模型创建查询
func (d *Database) Model(model interface{}) IQuery {
	return &QueryBuilder{
		db:    d,
		model: model,
	}
}

// Table 基于表名创建查询
func (d *Database) Table(name string) IQuery {
	return &QueryBuilder{
		db:    d,
		table: name,
	}
}

// Where 添加 WHERE 条件 - 性能优化版本
func (q *QueryBuilder) Where(query string, args ...interface{}) IQuery {
	if q.whereSQL == "" {
		q.whereSQL = query
	} else {
		// 使用 strings.Builder 优化字符串拼接
		var builder strings.Builder
		builder.Grow(len(q.whereSQL) + 5 + len(query)) // 预分配内存
		builder.WriteString(q.whereSQL)
		builder.WriteString(" AND ")
		builder.WriteString(query)
		q.whereSQL = builder.String()
	}
	q.whereArgs = append(q.whereArgs, args...)
	return q
}

// Or 添加 OR 条件 - 性能优化版本
func (q *QueryBuilder) Or(query string, args ...interface{}) IQuery {
	if q.whereSQL == "" {
		q.whereSQL = query
	} else {
		// 使用 strings.Builder 优化字符串拼接
		var builder strings.Builder
		builder.Grow(len(q.whereSQL) + 10 + len(query)) // 预分配内存
		builder.WriteString(" (")
		builder.WriteString(q.whereSQL)
		builder.WriteString(") OR ")
		builder.WriteString(query)
		q.whereSQL = builder.String()
	}
	q.whereArgs = append(q.whereArgs, args...)
	return q
}

// And 添加 AND 条件（同 Where）
func (q *QueryBuilder) And(query string, args ...interface{}) IQuery {
	return q.Where(query, args...)
}

// Select 选择要查询的字段
func (q *QueryBuilder) Select(fields ...string) IQuery {
	q.selectCols = fields
	return q
}

// Omit 排除指定的字段（暂未实现）
func (q *QueryBuilder) Omit(fields ...string) IQuery {
	// TODO: 实现字段排除逻辑，生成 SELECT 时排除这些字段
	return q
}

// Order 设置排序规则
func (q *QueryBuilder) Order(order string) IQuery {
	q.orderSQL = order
	return q
}

// OrderBy 按指定字段和方向排序
func (q *QueryBuilder) OrderBy(field string, direction string) IQuery {
	q.orderSQL = field + " " + direction
	return q
}

// Limit 限制返回数量
func (q *QueryBuilder) Limit(limit int) IQuery {
	q.limit = limit
	return q
}

// Offset 设置偏移量
func (q *QueryBuilder) Offset(offset int) IQuery {
	q.offset = offset
	return q
}

// Page 分页查询
func (q *QueryBuilder) Page(page, pageSize int) IQuery {
	q.limit = pageSize
	q.offset = (page - 1) * pageSize
	return q
}

// Group 设置分组字段
func (q *QueryBuilder) Group(group string) IQuery {
	q.groupSQL = group
	return q
}

// Having 添加 HAVING 条件
func (q *QueryBuilder) Having(having string, args ...interface{}) IQuery {
	q.havingSQL = having
	q.havingArgs = args
	return q
}

// Join 添加 JOIN 连接 - 性能优化版本
func (q *QueryBuilder) Join(join string, args ...interface{}) IQuery {
	if q.joinSQL == "" {
		q.joinSQL = join
	} else {
		// 使用 strings.Builder 优化字符串拼接
		var builder strings.Builder
		builder.Grow(len(q.joinSQL) + 1 + len(join)) // 预分配内存
		builder.WriteString(q.joinSQL)
		builder.WriteByte(' ')
		builder.WriteString(join)
		q.joinSQL = builder.String()
	}
	q.joinArgs = append(q.joinArgs, args...)
	return q
}

// LeftJoin 左连接
func (q *QueryBuilder) LeftJoin(table, on string) IQuery {
	return q.Join("LEFT JOIN " + table + " ON " + on)
}

// RightJoin 右连接
func (q *QueryBuilder) RightJoin(table, on string) IQuery {
	return q.Join("RIGHT JOIN " + table + " ON " + on)
}

// InnerJoin 内连接
func (q *QueryBuilder) InnerJoin(table, on string) IQuery {
	return q.Join("INNER JOIN " + table + " ON " + on)
}

// Preload 预加载关联数据（暂未实现）
func (q *QueryBuilder) Preload(relation string, conditions ...interface{}) IQuery {
	// TODO: 实现预加载逻辑
	return q
}

// First 查询第一条记录
func (q *QueryBuilder) First(result interface{}) error {
	q.limit = 1
	return q.Find(result)
}

// Find 查询多条记录
func (q *QueryBuilder) Find(result interface{}) error {
	sqlStr, args := q.BuildSelect()

	// 调试模式打印 SQL
	if q.debug || (q.db != nil && q.db.debug) {
		fmt.Printf("[Magic-ORM] SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 干跑模式不执行 SQL
	if q.dryRun {
		return nil
	}

	var rows *sql.Rows
	var err error

	// 判断是否在事务中
	if q.tx != nil {
		rows, err = q.tx.Query(sqlStr, args...)
	} else if q.db != nil && q.db.db != nil {
		rows, err = q.db.db.Query(sqlStr, args...)
	} else {
		return fmt.Errorf("数据库连接未初始化")
	}

	if err != nil {
		return fmt.Errorf("查询失败：%w", err)
	}
	defer rows.Close()

	// TODO: 实现结果映射逻辑
	// 使用 FieldMapper 将查询结果映射到 result

	return nil
}

// Count 统计记录数量
func (q *QueryBuilder) Count(count *int64) IQuery {
	// 构建 COUNT 查询
	originalSelect := q.selectCols
	q.selectCols = []string{"COUNT(*)"}

	sqlStr, args := q.BuildSelect()

	// 调试模式
	if q.debug || (q.db != nil && q.db.debug) {
		fmt.Printf("[Magic-ORM] COUNT SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 干跑模式
	if q.dryRun {
		return q
	}

	var err error
	if q.tx != nil {
		err = q.tx.QueryRow(sqlStr, args...).Scan(count)
	} else if q.db != nil && q.db.db != nil {
		err = q.db.db.QueryRow(sqlStr, args...).Scan(count)
	}

	if err != nil {
		fmt.Printf("[Magic-ORM] Count 错误：%v\n", err)
	}

	// 恢复原来的选择字段
	q.selectCols = originalSelect
	return q
}

// Exists 检查记录是否存在
func (q *QueryBuilder) Exists() (bool, error) {
	// 使用 LIMIT 1 优化查询
	originalLimit := q.limit
	q.limit = 1

	sqlStr, args := q.BuildSelect()

	// 调试模式
	if q.debug || (q.db != nil && q.db.debug) {
		fmt.Printf("[Magic-ORM] EXISTS SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 干跑模式
	if q.dryRun {
		return false, nil
	}

	var rows *sql.Rows
	var err error

	if q.tx != nil {
		rows, err = q.tx.Query(sqlStr, args...)
	} else if q.db != nil && q.db.db != nil {
		rows, err = q.db.db.Query(sqlStr, args...)
	} else {
		return false, fmt.Errorf("数据库连接未初始化")
	}
	defer rows.Close()

	if err != nil {
		return false, err
	}

	// 检查是否有结果
	exists := rows.Next()

	// 恢复原来的 limit
	q.limit = originalLimit

	return exists, nil
}

// Updates 更新数据
func (q *QueryBuilder) Updates(data interface{}) error {
	sqlStr, args := q.BuildUpdate(data)

	// 调试模式打印 SQL
	if q.debug || (q.db != nil && q.db.debug) {
		fmt.Printf("[Magic-ORM] UPDATE SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 干跑模式不执行 SQL
	if q.dryRun {
		return nil
	}

	var err error
	if q.tx != nil {
		_, err = q.tx.Exec(sqlStr, args...)
	} else if q.db != nil && q.db.db != nil {
		_, err = q.db.db.Exec(sqlStr, args...)
	} else {
		return fmt.Errorf("数据库连接未初始化")
	}

	if err != nil {
		return fmt.Errorf("更新失败：%w", err)
	}
	return nil
}

// UpdateColumn 更新单个字段
func (q *QueryBuilder) UpdateColumn(column string, value interface{}) error {
	return q.Updates(map[string]interface{}{column: value})
}

// Delete 删除数据
func (q *QueryBuilder) Delete() error {
	sqlStr, args := q.BuildDelete()

	// 调试模式打印 SQL
	if q.debug || (q.db != nil && q.db.debug) {
		fmt.Printf("[Magic-ORM] DELETE SQL: %s\n[Magic-ORM] Args: %v\n", sqlStr, args)
	}

	// 干跑模式不执行 SQL
	if q.dryRun {
		return nil
	}

	var err error
	if q.tx != nil {
		_, err = q.tx.Exec(sqlStr, args...)
	} else if q.db != nil && q.db.db != nil {
		_, err = q.db.db.Exec(sqlStr, args...)
	} else {
		return fmt.Errorf("数据库连接未初始化")
	}

	if err != nil {
		return fmt.Errorf("删除失败：%w", err)
	}
	return nil
}

// Unscoped 忽略软删除限制
func (q *QueryBuilder) Unscoped() IQuery {
	q.unscoped = true
	return q
}

// DryRun 设置干跑模式（只生成 SQL 不执行）
func (q *QueryBuilder) DryRun() IQuery {
	q.dryRun = true
	return q
}

// Debug 设置调试模式（打印 SQL 日志）
func (q *QueryBuilder) Debug() IQuery {
	q.debug = true
	return q
}

// Build 构建 SELECT SQL 语句
func (q *QueryBuilder) Build() (string, []interface{}) {
	return q.BuildSelect()
}

// BuildSelect 构建 SELECT SQL 语句
func (q *QueryBuilder) BuildSelect() (string, []interface{}) {
	var builder strings.Builder

	// SELECT 部分
	builder.WriteString("SELECT ")
	if len(q.selectCols) > 0 {
		builder.WriteString(strings.Join(q.selectCols, ", "))
	} else {
		builder.WriteString("*")
	}

	// FROM 部分
	builder.WriteString(" FROM ")
	if q.table != "" {
		builder.WriteString(q.table)
	} else if q.model != nil {
		// 从模型获取表名
		mapper := NewFieldMapper()
		builder.WriteString(mapper.GetTableName(q.model))
	} else {
		builder.WriteString("unknown_table")
	}

	// JOIN 部分
	if q.joinSQL != "" {
		builder.WriteString(" ")
		builder.WriteString(q.joinSQL)
	}

	// WHERE 部分
	if q.whereSQL != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(q.whereSQL)
	}

	// GROUP BY 部分
	if q.groupSQL != "" {
		builder.WriteString(" GROUP BY ")
		builder.WriteString(q.groupSQL)
	}

	// HAVING 部分
	if q.havingSQL != "" {
		builder.WriteString(" HAVING ")
		builder.WriteString(q.havingSQL)
	}

	// ORDER BY 部分
	if q.orderSQL != "" {
		builder.WriteString(" ORDER BY ")
		builder.WriteString(q.orderSQL)
	}

	// LIMIT 部分
	if q.limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", q.limit))
	}

	// OFFSET 部分
	if q.offset > 0 {
		builder.WriteString(fmt.Sprintf(" OFFSET %d", q.offset))
	}

	// 合并参数
	allArgs := make([]interface{}, 0)
	allArgs = append(allArgs, q.joinArgs...)
	allArgs = append(allArgs, q.whereArgs...)
	allArgs = append(allArgs, q.havingArgs...)

	return builder.String(), allArgs
}

// BuildUpdate 构建 UPDATE SQL 语句
func (q *QueryBuilder) BuildUpdate(data interface{}) (string, []interface{}) {
	var builder strings.Builder
	var args []interface{}

	builder.WriteString("UPDATE ")
	if q.table != "" {
		builder.WriteString(q.table)
	} else if q.model != nil {
		mapper := NewFieldMapper()
		builder.WriteString(mapper.GetTableName(q.model))
	} else {
		builder.WriteString("unknown_table")
	}

	builder.WriteString(" SET ")

	// 根据 data 类型生成 SET 子句
	switch v := data.(type) {
	case map[string]interface{}:
		// map 类型，生成 key=value 对
		setParts := make([]string, 0, len(v))
		for key, value := range v {
			setParts = append(setParts, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
		builder.WriteString(strings.Join(setParts, ", "))
	case string:
		// string 类型，直接使用（注意：实际使用需要转义）
		builder.WriteString(v)
	default:
		// 结构体类型，使用字段映射器
		mapper := NewFieldMapper()
		columns, err := mapper.StructToColumns(data)
		if err == nil && len(columns) > 0 {
			setParts := make([]string, 0, len(columns))
			for key := range columns {
				setParts = append(setParts, fmt.Sprintf("%s = ?", key))
				args = append(args, columns[key])
			}
			builder.WriteString(strings.Join(setParts, ", "))
		}
	}

	// WHERE 部分
	if q.whereSQL != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(q.whereSQL)
		args = append(args, q.whereArgs...)
	}

	return builder.String(), args
}

// BuildDelete 构建 DELETE SQL 语句
func (q *QueryBuilder) BuildDelete() (string, []interface{}) {
	var builder strings.Builder

	builder.WriteString("DELETE FROM ")
	if q.table != "" {
		builder.WriteString(q.table)
	} else if q.model != nil {
		mapper := NewFieldMapper()
		builder.WriteString(mapper.GetTableName(q.model))
	} else {
		builder.WriteString("unknown_table")
	}

	if q.whereSQL != "" {
		builder.WriteString(" WHERE ")
		builder.WriteString(q.whereSQL)
	}

	return builder.String(), q.whereArgs
}

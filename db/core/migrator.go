package core

import (
	"fmt"
	"strings"
)

// Migrator 迁移管理器实现 - 处理数据库架构的自动迁移
type Migrator struct {
	db *Database // 数据库连接实例
}

// NewMigrator 创建迁移管理器实例
func NewMigrator(db *Database) IMigrator {
	return &Migrator{db: db}
}

// AutoMigrate 自动迁移 - 根据模型自动创建或更新数据库表结构
func (m *Migrator) AutoMigrate(models ...interface{}) error {
	for _, model := range models {
		if err := m.CreateTable(model); err != nil {
			return fmt.Errorf("创建表失败：%w", err)
		}
	}
	return nil
}

// CreateTable 创建表 - 根据模型创建数据库表
func (m *Migrator) CreateTable(model interface{}) error {
	mapper := NewFieldMapper()

	// 获取表名
	tableName := mapper.GetTableName(model)

	// 获取字段信息
	fields := mapper.GetFields(model)
	if len(fields) == 0 {
		return fmt.Errorf("模型没有有效的字段")
	}

	// 生成 CREATE TABLE SQL
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName))

	columnDefs := make([]string, 0)
	for _, field := range fields {
		colDef := fmt.Sprintf("%s %s", field.Column, field.DbType)

		// 添加主键约束
		if field.IsPrimary {
			colDef += " PRIMARY KEY"
			if field.IsAuto {
				colDef += " AUTOINCREMENT"
			}
		}

		// 添加 NOT NULL 约束（可选）
		// colDef += " NOT NULL"

		columnDefs = append(columnDefs, colDef)
	}

	sqlBuilder.WriteString(strings.Join(columnDefs, ", "))
	sqlBuilder.WriteString(")")

	createSQL := sqlBuilder.String()

	if m.db.debug {
		fmt.Printf("[Magic-ORM] CREATE TABLE SQL: %s\n", createSQL)
	}

	// 执行 SQL
	_, err := m.db.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("执行 CREATE TABLE 失败：%w", err)
	}

	return nil
}

// DropTable 删除表 - 删除指定的数据库表
func (m *Migrator) DropTable(model interface{}) error {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)

	if m.db.debug {
		fmt.Printf("[Magic-ORM] DROP TABLE SQL: %s\n", dropSQL)
	}

	_, err := m.db.db.Exec(dropSQL)
	if err != nil {
		return fmt.Errorf("执行 DROP TABLE 失败：%w", err)
	}

	return nil
}

// HasTable 检查表是否存在 - 验证数据库中是否已存在指定表
func (m *Migrator) HasTable(model interface{}) (bool, error) {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	// SQLite 检查表是否存在的 SQL
	checkSQL := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`

	var count int
	err := m.db.db.QueryRow(checkSQL, tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查表是否存在失败：%w", err)
	}

	return count > 0, nil
}

// RenameTable 重命名表 - 修改数据库表的名称
func (m *Migrator) RenameTable(oldName, newName string) error {
	renameSQL := fmt.Sprintf("ALTER TABLE %s RENAME TO %s", oldName, newName)

	if m.db.debug {
		fmt.Printf("[Magic-ORM] RENAME TABLE SQL: %s\n", renameSQL)
	}

	_, err := m.db.db.Exec(renameSQL)
	if err != nil {
		return fmt.Errorf("重命名表失败：%w", err)
	}

	return nil
}

// AddColumn 添加列 - 向表中添加新的字段
func (m *Migrator) AddColumn(model interface{}, field string) error {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	// 获取字段信息
	fields := mapper.GetFields(model)
	var targetField *FieldInfo

	for _, f := range fields {
		if f.Name == field || f.Column == field {
			targetField = &f
			break
		}
	}

	if targetField == nil {
		return fmt.Errorf("字段不存在：%s", field)
	}

	addSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		tableName, targetField.Column, targetField.DbType)

	if m.db.debug {
		fmt.Printf("[Magic-ORM] ADD COLUMN SQL: %s\n", addSQL)
	}

	_, err := m.db.db.Exec(addSQL)
	if err != nil {
		return fmt.Errorf("添加列失败：%w", err)
	}

	return nil
}

// DropColumn 删除列 - 从表中删除指定的字段
func (m *Migrator) DropColumn(model interface{}, field string) error {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	// SQLite 不直接支持 DROP COLUMN，需要重建表
	// 这里使用简化方案：创建新表 -> 复制数据 -> 删除旧表 -> 重命名

	_ = tableName // 避免编译错误
	return fmt.Errorf("SQLite 不支持直接删除列，需要手动重建表")
}

// HasColumn 检查列是否存在 - 验证表中是否已存在指定字段
func (m *Migrator) HasColumn(model interface{}, field string) (bool, error) {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	// SQLite 检查列是否存在的 SQL
	checkSQL := `PRAGMA table_info(` + tableName + `)`

	rows, err := m.db.db.Query(checkSQL)
	if err != nil {
		return false, fmt.Errorf("检查列失败：%w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var dfltValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &typ, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}

		if name == field {
			return true, nil
		}
	}

	return false, nil
}

// RenameColumn 重命名列 - 修改表中字段的名称
func (m *Migrator) RenameColumn(model interface{}, oldField, newField string) error {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	// SQLite 3.25.0+ 支持 ALTER TABLE ... RENAME COLUMN
	renameSQL := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
		tableName, oldField, newField)

	if m.db.debug {
		fmt.Printf("[Magic-ORM] RENAME COLUMN SQL: %s\n", renameSQL)
	}

	_, err := m.db.db.Exec(renameSQL)
	if err != nil {
		return fmt.Errorf("重命名列失败：%w", err)
	}

	return nil
}

// CreateIndex 创建索引 - 为表中的字段创建索引
func (m *Migrator) CreateIndex(model interface{}, field string) error {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	indexName := fmt.Sprintf("idx_%s_%s", tableName, field)
	createSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
		indexName, tableName, field)

	if m.db.debug {
		fmt.Printf("[Magic-ORM] CREATE INDEX SQL: %s\n", createSQL)
	}

	_, err := m.db.db.Exec(createSQL)
	if err != nil {
		return fmt.Errorf("创建索引失败：%w", err)
	}

	return nil
}

// DropIndex 删除索引 - 删除表中的指定索引
func (m *Migrator) DropIndex(model interface{}, field string) error {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	indexName := fmt.Sprintf("idx_%s_%s", tableName, field)
	dropSQL := fmt.Sprintf("DROP INDEX IF EXISTS %s", indexName)

	if m.db.debug {
		fmt.Printf("[Magic-ORM] DROP INDEX SQL: %s\n", dropSQL)
	}

	_, err := m.db.db.Exec(dropSQL)
	if err != nil {
		return fmt.Errorf("删除索引失败：%w", err)
	}

	return nil
}

// HasIndex 检查索引是否存在 - 验证表中是否已存在指定索引
func (m *Migrator) HasIndex(model interface{}, field string) (bool, error) {
	mapper := NewFieldMapper()
	tableName := mapper.GetTableName(model)

	indexName := fmt.Sprintf("idx_%s_%s", tableName, field)

	checkSQL := `SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name=?`

	var count int
	err := m.db.db.QueryRow(checkSQL, indexName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("检查索引失败：%w", err)
	}

	return count > 0, nil
}

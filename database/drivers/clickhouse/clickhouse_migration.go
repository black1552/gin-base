// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"git.magicany.cc/black1552/gin-base/database"
)

// Migration implements database migration operations for ClickHouse.
type Migration struct {
	*database.MigrationCore
	*database.AutoMigrateCore
}

// NewMigration creates a new ClickHouse Migration instance.
func NewMigration(db database.DB) *Migration {
	return &Migration{
		MigrationCore:   database.NewMigrationCore(db),
		AutoMigrateCore: database.NewAutoMigrateCore(db),
	}
}

// CreateTable creates a new table with the given name and column definitions.
func (m *Migration) CreateTable(ctx context.Context, table string, columns map[string]*database.ColumnDefinition, options ...database.TableOption) error {
	if len(columns) == 0 {
		return fmt.Errorf("cannot create table without columns")
	}

	var opts database.TableOptions
	for _, opt := range options {
		opt(&opts)
	}

	var sql strings.Builder
	sql.WriteString("CREATE TABLE ")
	if opts.IfNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(database.QuoteIdentifier(table))
	sql.WriteString(" (\n")

	// Add columns
	var colDefs []string
	for name, def := range columns {
		colDef := m.buildColumnDefinition(name, def)
		colDefs = append(colDefs, "    "+colDef)
	}

	sql.WriteString(strings.Join(colDefs, ",\n"))
	sql.WriteString("\n)")

	// Add engine specification (required for ClickHouse)
	sql.WriteString(" ENGINE = MergeTree()")

	if opts.Comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT '%s'", escapeString(opts.Comment)))
	}

	return m.ExecuteSQL(ctx, sql.String())
}

// buildColumnDefinition builds column definition for ClickHouse.
func (m *Migration) buildColumnDefinition(name string, def *database.ColumnDefinition) string {
	var parts []string
	parts = append(parts, database.QuoteIdentifier(name))

	// Handle ClickHouse-specific types
	dbType := def.Type
	parts = append(parts, dbType)

	if !def.Null {
		// ClickHouse uses Nullable type wrapper
		if !strings.HasPrefix(dbType, "Nullable(") {
			// Type is already non-nullable by default in ClickHouse
		}
	} else {
		// Make type nullable if needed
		if !strings.HasPrefix(dbType, "Nullable(") {
			parts[len(parts)-1] = fmt.Sprintf("Nullable(%s)", dbType)
		}
	}

	if def.Default != nil {
		defaultValue := formatDefaultValue(def.Default)
		parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultValue))
	}

	if def.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", escapeString(def.Comment)))
	}

	return strings.Join(parts, " ")
}

// DropTable drops an existing table from the database.
func (m *Migration) DropTable(ctx context.Context, table string, ifExists ...bool) error {
	sql := "DROP TABLE "
	if len(ifExists) > 0 && ifExists[0] {
		sql += "IF EXISTS "
	}
	sql += database.QuoteIdentifier(table)
	return m.ExecuteSQL(ctx, sql)
}

// HasTable checks if a table exists in the database.
func (m *Migration) HasTable(ctx context.Context, table string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "currentDatabase()"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM system.tables WHERE database = %s AND name = '%s'",
		schema,
		table,
	)

	value, err := m.GetDB().GetValue(ctx, query)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}

// RenameTable renames an existing table from oldName to newName.
func (m *Migration) RenameTable(ctx context.Context, oldName, newName string) error {
	sql := fmt.Sprintf(
		"RENAME TABLE %s TO %s",
		database.QuoteIdentifier(oldName),
		database.QuoteIdentifier(newName),
	)
	return m.ExecuteSQL(ctx, sql)
}

// TruncateTable removes all records from a table but keeps the table structure.
func (m *Migration) TruncateTable(ctx context.Context, table string) error {
	sql := fmt.Sprintf("TRUNCATE TABLE %s", database.QuoteIdentifier(table))
	return m.ExecuteSQL(ctx, sql)
}

// AddColumn adds a new column to an existing table.
func (m *Migration) AddColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	colDef := m.buildColumnDefinition(column, definition)
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", database.QuoteIdentifier(table), colDef)
	return m.ExecuteSQL(ctx, sql)
}

// DropColumn removes a column from an existing table.
func (m *Migration) DropColumn(ctx context.Context, table, column string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s DROP COLUMN %s",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(column),
	)
	return m.ExecuteSQL(ctx, sql)
}

// RenameColumn renames a column in an existing table.
func (m *Migration) RenameColumn(ctx context.Context, table, oldName, newName string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s RENAME COLUMN %s TO %s",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(oldName),
		database.QuoteIdentifier(newName),
	)
	return m.ExecuteSQL(ctx, sql)
}

// ModifyColumn modifies an existing column's definition.
func (m *Migration) ModifyColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	colDef := m.buildColumnDefinition(column, definition)
	sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", database.QuoteIdentifier(table), colDef)
	return m.ExecuteSQL(ctx, sql)
}

// HasColumn checks if a column exists in a table.
func (m *Migration) HasColumn(ctx context.Context, table, column string) (bool, error) {
	fields, err := m.GetDB().TableFields(ctx, table)
	if err != nil {
		return false, err
	}
	_, exists := fields[column]
	return exists, nil
}

// CreateIndex creates a new index on the specified table and columns.
// Note: ClickHouse has limited index support compared to traditional databases.
func (m *Migration) CreateIndex(ctx context.Context, table, index string, columns []string, options ...database.IndexOption) error {
	var opts database.IndexOptions
	for _, opt := range options {
		opt(&opts)
	}

	// ClickHouse uses different indexing mechanisms
	// For now, we'll use ALTER TABLE ADD INDEX syntax
	colList := m.BuildIndexColumns(columns)
	sql := fmt.Sprintf(
		"ALTER TABLE %s ADD INDEX %s (%s) TYPE minmax GRANULARITY 1",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(index),
		colList,
	)

	return m.ExecuteSQL(ctx, sql)
}

// DropIndex drops an existing index from a table.
func (m *Migration) DropIndex(ctx context.Context, table, index string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s DROP INDEX %s",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(index),
	)
	return m.ExecuteSQL(ctx, sql)
}

// HasIndex checks if an index exists on a table.
func (m *Migration) HasIndex(ctx context.Context, table, index string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "currentDatabase()"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM system.data_skipping_indices WHERE database = %s AND table = '%s' AND name = '%s'",
		schema,
		table,
		index,
	)

	value, err := m.GetDB().GetValue(ctx, query)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}

// CreateForeignKey is not supported in ClickHouse.
func (m *Migration) CreateForeignKey(ctx context.Context, table, constraint string, columns []string, refTable string, refColumns []string, options ...database.ForeignKeyOption) error {
	// ClickHouse does not support foreign keys
	return fmt.Errorf("ClickHouse does not support foreign key constraints")
}

// DropForeignKey is not supported in ClickHouse.
func (m *Migration) DropForeignKey(ctx context.Context, table, constraint string) error {
	// ClickHouse does not support foreign keys
	return fmt.Errorf("ClickHouse does not support foreign key constraints")
}

// HasForeignKey always returns false as ClickHouse doesn't support foreign keys.
func (m *Migration) HasForeignKey(ctx context.Context, table, constraint string) (bool, error) {
	return false, nil
}

// CreateSchema creates a new database schema.
func (m *Migration) CreateSchema(ctx context.Context, schema string) error {
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database.QuoteIdentifier(schema))
	return m.ExecuteSQL(ctx, sql)
}

// DropSchema drops an existing database schema.
func (m *Migration) DropSchema(ctx context.Context, schema string, cascade ...bool) error {
	sql := fmt.Sprintf("DROP DATABASE IF EXISTS %s", database.QuoteIdentifier(schema))
	return m.ExecuteSQL(ctx, sql)
}

// HasSchema checks if a schema exists.
func (m *Migration) HasSchema(ctx context.Context, schema string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM system.databases WHERE name = '%s'",
		schema,
	)

	value, err := m.GetDB().GetValue(ctx, query)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}

// formatDefaultValue formats the default value for SQL.
func formatDefaultValue(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeString(v))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// escapeString escapes special characters in strings for SQL.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

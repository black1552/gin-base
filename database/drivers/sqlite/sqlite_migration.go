// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/black1552/gin-base/database"
)

// Migration implements database migration operations for SQLite.
type Migration struct {
	*database.MigrationCore
	*database.AutoMigrateCore
}

// NewMigration creates a new SQLite Migration instance.
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
	var primaryKeys []string
	for name, def := range columns {
		colDef := m.buildColumnDefinition(name, def)
		// Only add to primaryKeys if it's not an auto-increment column
		// (auto-increment columns already have PRIMARY KEY in their definition)
		if def.PrimaryKey && !def.AutoIncrement {
			primaryKeys = append(primaryKeys, database.QuoteIdentifier(name))
		}
		colDefs = append(colDefs, "    "+colDef)
	}

	// Add composite primary key constraint if needed (for non-auto-increment keys)
	if len(primaryKeys) > 1 {
		colDefs = append(colDefs, fmt.Sprintf("    PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	sql.WriteString(strings.Join(colDefs, ",\n"))
	sql.WriteString("\n)")

	return m.ExecuteSQL(ctx, sql.String())
}

// buildColumnDefinition builds column definition for SQLite.
func (m *Migration) buildColumnDefinition(name string, def *database.ColumnDefinition) string {
	var parts []string
	parts = append(parts, database.QuoteIdentifier(name))

	// Handle SQLite-specific types for auto-increment primary key
	if def.PrimaryKey && def.AutoIncrement {
		// SQLite requires INTEGER type for AUTOINCREMENT
		parts = append(parts, "INTEGER PRIMARY KEY AUTOINCREMENT")
		return strings.Join(parts, " ")
	}

	// Regular column definition
	parts = append(parts, def.Type)

	if !def.Null {
		parts = append(parts, "NOT NULL")
	}

	if def.Unique && !def.PrimaryKey {
		parts = append(parts, "UNIQUE")
	}

	if def.Default != nil {
		defaultValue := formatDefaultValue(def.Default)
		parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultValue))
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
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='%s'",
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
		"ALTER TABLE %s RENAME TO %s",
		database.QuoteIdentifier(oldName),
		database.QuoteIdentifier(newName),
	)
	return m.ExecuteSQL(ctx, sql)
}

// TruncateTable removes all records from a table but keeps the table structure.
func (m *Migration) TruncateTable(ctx context.Context, table string) error {
	sql := fmt.Sprintf("DELETE FROM %s", database.QuoteIdentifier(table))
	return m.ExecuteSQL(ctx, sql)
}

// AddColumn adds a new column to an existing table.
func (m *Migration) AddColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	colDef := m.buildColumnDefinition(column, definition)
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", database.QuoteIdentifier(table), colDef)
	return m.ExecuteSQL(ctx, sql)
}

// DropColumn removes a column from an existing table.
// Note: SQLite has limited ALTER TABLE support. This may require table recreation.
func (m *Migration) DropColumn(ctx context.Context, table, column string) error {
	// SQLite 3.35.0+ supports DROP COLUMN
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
		database.QuoteIdentifierDouble(newName),
	)
	return m.ExecuteSQL(ctx, sql)
}

// ModifyColumn modifies an existing column's definition.
// Note: SQLite requires table recreation for most column modifications.
func (m *Migration) ModifyColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	// SQLite has very limited ALTER TABLE support
	// This would typically require recreating the table
	return fmt.Errorf("SQLite does not support MODIFY COLUMN directly, table recreation required")
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
func (m *Migration) CreateIndex(ctx context.Context, table, index string, columns []string, options ...database.IndexOption) error {
	var opts database.IndexOptions
	for _, opt := range options {
		opt(&opts)
	}

	var sql strings.Builder
	sql.WriteString("CREATE ")

	if opts.Unique {
		sql.WriteString("UNIQUE ")
	}

	sql.WriteString("INDEX ")
	if opts.IfNotExists {
		sql.WriteString("IF NOT EXISTS ")
	}
	sql.WriteString(database.QuoteIdentifier(index))
	sql.WriteString(" ON ")
	sql.WriteString(database.QuoteIdentifier(table))

	colList := m.BuildIndexColumns(columns)
	sql.WriteString(fmt.Sprintf(" (%s)", colList))

	return m.ExecuteSQL(ctx, sql.String())
}

// DropIndex drops an existing index from a table.
func (m *Migration) DropIndex(ctx context.Context, table, index string) error {
	sql := fmt.Sprintf("DROP INDEX IF EXISTS %s", database.QuoteIdentifier(index))
	return m.ExecuteSQL(ctx, sql)
}

// HasIndex checks if an index exists on a table.
func (m *Migration) HasIndex(ctx context.Context, table, index string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='%s' AND tbl_name='%s'",
		index,
		table,
	)

	value, err := m.GetDB().GetValue(ctx, query)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}

// CreateForeignKey creates a foreign key constraint.
// Note: SQLite requires foreign keys to be enabled with PRAGMA foreign_keys = ON
func (m *Migration) CreateForeignKey(ctx context.Context, table, constraint string, columns []string, refTable string, refColumns []string, options ...database.ForeignKeyOption) error {
	// SQLite doesn't support adding foreign keys to existing tables
	// Foreign keys must be defined during table creation
	return fmt.Errorf("SQLite does not support adding foreign keys to existing tables")
}

// DropForeignKey drops a foreign key constraint.
func (m *Migration) DropForeignKey(ctx context.Context, table, constraint string) error {
	// SQLite doesn't support dropping foreign keys from existing tables
	return fmt.Errorf("SQLite does not support dropping foreign keys from existing tables")
}

// HasForeignKey checks if a foreign key constraint exists.
func (m *Migration) HasForeignKey(ctx context.Context, table, constraint string) (bool, error) {
	// Query pragma to check foreign keys
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", database.QuoteIdentifier(table))

	result, err := m.GetDB().GetAll(ctx, query)
	if err != nil {
		return false, err
	}

	// Check if constraint exists in the result
	for _, row := range result {
		if id, ok := row["id"]; ok && id.String() == constraint {
			return true, nil
		}
	}

	return false, nil
}

// CreateSchema is not applicable for SQLite (single database file).
func (m *Migration) CreateSchema(ctx context.Context, schema string) error {
	// SQLite doesn't support schemas
	return fmt.Errorf("SQLite does not support schemas")
}

// DropSchema is not applicable for SQLite (single database file).
func (m *Migration) DropSchema(ctx context.Context, schema string, cascade ...bool) error {
	// SQLite doesn't support schemas
	return fmt.Errorf("SQLite does not support schemas")
}

// HasSchema checks if a schema exists (always returns false for SQLite).
func (m *Migration) HasSchema(ctx context.Context, schema string) (bool, error) {
	// SQLite doesn't support schemas
	return false, nil
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
	s = strings.ReplaceAll(s, "'", "''")
	return s
}

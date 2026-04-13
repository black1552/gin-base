// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package mysql

import (
	"context"
	"fmt"
	"strings"

	"git.magicany.cc/black1552/gin-base/database"
)

// Migration implements database migration operations for MySQL.
type Migration struct {
	*database.MigrationCore
	*database.AutoMigrateCore
}

// NewMigration creates a new MySQL Migration instance.
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
		colDef := m.BuildColumnDefinition(name, def)
		if def.PrimaryKey {
			primaryKeys = append(primaryKeys, database.QuoteIdentifier(name))
		}
		colDefs = append(colDefs, "    "+colDef)
	}

	// Add primary key constraint if needed
	if len(primaryKeys) > 0 {
		colDefs = append(colDefs, fmt.Sprintf("    PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	sql.WriteString(strings.Join(colDefs, ",\n"))
	sql.WriteString("\n)")

	// Add table options
	if opts.Engine != "" {
		sql.WriteString(fmt.Sprintf(" ENGINE=%s", opts.Engine))
	}
	if opts.Charset != "" {
		sql.WriteString(fmt.Sprintf(" DEFAULT CHARSET=%s", opts.Charset))
	}
	if opts.Collation != "" {
		sql.WriteString(fmt.Sprintf(" COLLATE=%s", opts.Collation))
	}
	if opts.Comment != "" {
		sql.WriteString(fmt.Sprintf(" COMMENT='%s'", escapeString(opts.Comment)))
	}

	return m.ExecuteSQL(ctx, sql.String())
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
		schema = "DATABASE()"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = %s AND table_name = '%s'",
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
	colDef := m.BuildColumnDefinition(column, definition)
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
	// MySQL requires the full column definition when renaming
	fields, err := m.GetDB().TableFields(ctx, table)
	if err != nil {
		return err
	}

	field, ok := fields[oldName]
	if !ok {
		return fmt.Errorf("column %s does not exist in table %s", oldName, table)
	}

	def := &database.ColumnDefinition{
		Type:          field.Type,
		Null:          field.Null,
		Default:       field.Default,
		Comment:       field.Comment,
		AutoIncrement: strings.Contains(field.Extra, "auto_increment"),
	}

	colDef := m.BuildColumnDefinition(newName, def)
	sql := fmt.Sprintf(
		"ALTER TABLE %s CHANGE COLUMN %s %s",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(oldName),
		colDef,
	)
	return m.ExecuteSQL(ctx, sql)
}

// ModifyColumn modifies an existing column's definition.
func (m *Migration) ModifyColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	colDef := m.BuildColumnDefinition(column, definition)
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
	if opts.FullText {
		sql.WriteString("FULLTEXT ")
	}
	if opts.Spatial {
		sql.WriteString("SPATIAL ")
	}

	sql.WriteString("INDEX ")
	sql.WriteString(database.QuoteIdentifier(index))
	sql.WriteString(" ON ")
	sql.WriteString(database.QuoteIdentifier(table))

	colList := m.BuildIndexColumns(columns)
	sql.WriteString(fmt.Sprintf(" (%s)", colList))

	if opts.Using != "" {
		sql.WriteString(fmt.Sprintf(" USING %s", opts.Using))
	}

	return m.ExecuteSQL(ctx, sql.String())
}

// DropIndex drops an existing index from a table.
func (m *Migration) DropIndex(ctx context.Context, table, index string) error {
	sql := fmt.Sprintf(
		"DROP INDEX %s ON %s",
		database.QuoteIdentifier(index),
		database.QuoteIdentifier(table),
	)
	return m.ExecuteSQL(ctx, sql)
}

// HasIndex checks if an index exists on a table.
func (m *Migration) HasIndex(ctx context.Context, table, index string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "DATABASE()"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM information_schema.statistics WHERE table_schema = %s AND table_name = '%s' AND index_name = '%s'",
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

// CreateForeignKey creates a foreign key constraint.
func (m *Migration) CreateForeignKey(ctx context.Context, table, constraint string, columns []string, refTable string, refColumns []string, options ...database.ForeignKeyOption) error {
	var opts database.ForeignKeyOptions
	for _, opt := range options {
		opt(&opts)
	}

	sql := fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY %s REFERENCES %s %s",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(constraint),
		m.BuildForeignKeyColumns(columns),
		database.QuoteIdentifier(refTable),
		m.BuildForeignKeyColumns(refColumns),
	)

	if opts.OnDelete != "" {
		sql += fmt.Sprintf(" ON DELETE %s", opts.OnDelete)
	}
	if opts.OnUpdate != "" {
		sql += fmt.Sprintf(" ON UPDATE %s", opts.OnUpdate)
	}

	return m.ExecuteSQL(ctx, sql)
}

// DropForeignKey drops a foreign key constraint.
func (m *Migration) DropForeignKey(ctx context.Context, table, constraint string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s DROP FOREIGN KEY %s",
		database.QuoteIdentifier(table),
		database.QuoteIdentifier(constraint),
	)
	return m.ExecuteSQL(ctx, sql)
}

// HasForeignKey checks if a foreign key constraint exists.
func (m *Migration) HasForeignKey(ctx context.Context, table, constraint string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "DATABASE()"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM information_schema.table_constraints WHERE constraint_schema = %s AND table_name = '%s' AND constraint_name = '%s' AND constraint_type = 'FOREIGN KEY'",
		schema,
		table,
		constraint,
	)

	value, err := m.GetDB().GetValue(ctx, query)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}

// CreateSchema creates a new database schema.
func (m *Migration) CreateSchema(ctx context.Context, schema string) error {
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database.QuoteIdentifier(schema))
	return m.ExecuteSQL(ctx, sql)
}

// DropSchema drops an existing database schema.
func (m *Migration) DropSchema(ctx context.Context, schema string, cascade ...bool) error {
	sql := "DROP DATABASE "
	if len(cascade) > 0 && cascade[0] {
		sql += "IF EXISTS "
	}
	sql += database.QuoteIdentifier(schema)
	return m.ExecuteSQL(ctx, sql)
}

// HasSchema checks if a schema exists.
func (m *Migration) HasSchema(ctx context.Context, schema string) (bool, error) {
	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '%s'",
		schema,
	)

	value, err := m.GetDB().GetValue(ctx, query)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}

// escapeString escapes special characters in strings for SQL.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package pgsql

import (
	"context"
	"fmt"
	"strings"

	"git.magicany.cc/black1552/gin-base/database"
)

// Migration implements database migration operations for PostgreSQL.
type Migration struct {
	*database.MigrationCore
	*database.AutoMigrateCore
}

// NewMigration creates a new PostgreSQL Migration instance.
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
	sql.WriteString(database.QuoteIdentifierDouble(table))
	sql.WriteString(" (\n")

	// Add columns
	var colDefs []string
	var primaryKeys []string
	for name, def := range columns {
		colDef := m.buildColumnDefinition(name, def)
		if def.PrimaryKey {
			primaryKeys = append(primaryKeys, database.QuoteIdentifierDouble(name))
		}
		colDefs = append(colDefs, "    "+colDef)
	}

	// Add primary key constraint if needed
	if len(primaryKeys) > 0 {
		colDefs = append(colDefs, fmt.Sprintf("    PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	sql.WriteString(strings.Join(colDefs, ",\n"))
	sql.WriteString("\n)")

	// Add table comment if provided
	if opts.Comment != "" {
		commentSQL := fmt.Sprintf(
			"COMMENT ON TABLE %s IS '%s'",
			database.QuoteIdentifierDouble(table),
			escapeString(opts.Comment),
		)
		if err := m.ExecuteSQL(ctx, commentSQL); err != nil {
			return err
		}
	}

	return m.ExecuteSQL(ctx, sql.String())
}

// buildColumnDefinition builds column definition for PostgreSQL.
func (m *Migration) buildColumnDefinition(name string, def *database.ColumnDefinition) string {
	var parts []string
	parts = append(parts, database.QuoteIdentifierDouble(name))

	// Handle PostgreSQL-specific types
	dbType := def.Type
	if def.AutoIncrement {
		if dbType == "INT" || dbType == "INTEGER" {
			dbType = "SERIAL"
		} else if dbType == "BIGINT" {
			dbType = "BIGSERIAL"
		}
	}
	parts = append(parts, dbType)

	if def.PrimaryKey {
		// Primary key is handled separately
	} else {
		if !def.Null {
			parts = append(parts, "NOT NULL")
		}

		if def.Unique {
			parts = append(parts, "UNIQUE")
		}

		if def.Default != nil {
			defaultValue := formatDefaultValue(def.Default)
			parts = append(parts, fmt.Sprintf("DEFAULT %s", defaultValue))
		}
	}

	return strings.Join(parts, " ")
}

// DropTable drops an existing table from the database.
func (m *Migration) DropTable(ctx context.Context, table string, ifExists ...bool) error {
	sql := "DROP TABLE "
	if len(ifExists) > 0 && ifExists[0] {
		sql += "IF EXISTS "
	}
	sql += database.QuoteIdentifierDouble(table)
	return m.ExecuteSQL(ctx, sql)
}

// HasTable checks if a table exists in the database.
func (m *Migration) HasTable(ctx context.Context, table string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "current_schema()"
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
		"ALTER TABLE %s RENAME TO %s",
		database.QuoteIdentifierDouble(oldName),
		database.QuoteIdentifierDouble(newName),
	)
	return m.ExecuteSQL(ctx, sql)
}

// TruncateTable removes all records from a table but keeps the table structure.
func (m *Migration) TruncateTable(ctx context.Context, table string) error {
	sql := fmt.Sprintf("TRUNCATE TABLE %s", database.QuoteIdentifierDouble(table))
	return m.ExecuteSQL(ctx, sql)
}

// AddColumn adds a new column to an existing table.
func (m *Migration) AddColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	colDef := m.buildColumnDefinition(column, definition)
	sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", database.QuoteIdentifierDouble(table), colDef)
	return m.ExecuteSQL(ctx, sql)
}

// DropColumn removes a column from an existing table.
func (m *Migration) DropColumn(ctx context.Context, table, column string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s DROP COLUMN %s",
		database.QuoteIdentifierDouble(table),
		database.QuoteIdentifierDouble(column),
	)
	return m.ExecuteSQL(ctx, sql)
}

// RenameColumn renames a column in an existing table.
func (m *Migration) RenameColumn(ctx context.Context, table, oldName, newName string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s RENAME COLUMN %s TO %s",
		database.QuoteIdentifierDouble(table),
		database.QuoteIdentifierDouble(oldName),
		database.QuoteIdentifierDouble(newName),
	)
	return m.ExecuteSQL(ctx, sql)
}

// ModifyColumn modifies an existing column's definition.
func (m *Migration) ModifyColumn(ctx context.Context, table, column string, definition *database.ColumnDefinition) error {
	// PostgreSQL requires multiple ALTER statements for different modifications
	var statements []string

	if definition.Type != "" {
		statements = append(statements, fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN %s TYPE %s",
			database.QuoteIdentifierDouble(table),
			database.QuoteIdentifierDouble(column),
			definition.Type,
		))
	}

	if definition.Default != nil {
		defaultValue := formatDefaultValue(definition.Default)
		statements = append(statements, fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
			database.QuoteIdentifierDouble(table),
			database.QuoteIdentifierDouble(column),
			defaultValue,
		))
	} else {
		statements = append(statements, fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT",
			database.QuoteIdentifierDouble(table),
			database.QuoteIdentifierDouble(column),
		))
	}

	if !definition.Null {
		statements = append(statements, fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN %s SET NOT NULL",
			database.QuoteIdentifierDouble(table),
			database.QuoteIdentifierDouble(column),
		))
	} else {
		statements = append(statements, fmt.Sprintf(
			"ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL",
			database.QuoteIdentifierDouble(table),
			database.QuoteIdentifierDouble(column),
		))
	}

	// Execute all statements
	for _, stmt := range statements {
		if err := m.ExecuteSQL(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
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
	sql.WriteString(database.QuoteIdentifierDouble(index))
	sql.WriteString(" ON ")
	sql.WriteString(database.QuoteIdentifierDouble(table))

	colList := m.BuildIndexColumnsWithDouble(columns)
	sql.WriteString(fmt.Sprintf(" (%s)", colList))

	if opts.Using != "" {
		sql.WriteString(fmt.Sprintf(" USING %s", opts.Using))
	}

	if opts.Comment != "" {
		// Comment will be added after index creation
	}

	err := m.ExecuteSQL(ctx, sql.String())
	if err != nil {
		return err
	}

	// Add comment if provided
	if opts.Comment != "" {
		commentSQL := fmt.Sprintf(
			"COMMENT ON INDEX %s IS '%s'",
			database.QuoteIdentifierDouble(index),
			escapeString(opts.Comment),
		)
		return m.ExecuteSQL(ctx, commentSQL)
	}

	return nil
}

// BuildIndexColumnsWithDouble builds the column list for index creation using double quotes.
func (m *Migration) BuildIndexColumnsWithDouble(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = database.QuoteIdentifierDouble(col)
	}
	return strings.Join(quoted, ", ")
}

// DropIndex drops an existing index from a table.
func (m *Migration) DropIndex(ctx context.Context, table, index string) error {
	sql := fmt.Sprintf("DROP INDEX %s", database.QuoteIdentifierDouble(index))
	return m.ExecuteSQL(ctx, sql)
}

// HasIndex checks if an index exists on a table.
func (m *Migration) HasIndex(ctx context.Context, table, index string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "current_schema()"
	} else {
		schema = fmt.Sprintf("'%s'", schema)
	}

	query := fmt.Sprintf(
		"SELECT COUNT(*) FROM pg_indexes WHERE schemaname = %s AND tablename = '%s' AND indexname = '%s'",
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
		database.QuoteIdentifierDouble(table),
		database.QuoteIdentifierDouble(constraint),
		m.BuildForeignKeyColumnsWithDouble(columns),
		database.QuoteIdentifierDouble(refTable),
		m.BuildForeignKeyColumnsWithDouble(refColumns),
	)

	if opts.OnDelete != "" {
		sql += fmt.Sprintf(" ON DELETE %s", opts.OnDelete)
	}
	if opts.OnUpdate != "" {
		sql += fmt.Sprintf(" ON UPDATE %s", opts.OnUpdate)
	}
	if opts.Deferrable {
		sql += " DEFERRABLE"
		if opts.InitiallyDeferred {
			sql += " INITIALLY DEFERRED"
		}
	}

	return m.ExecuteSQL(ctx, sql)
}

// BuildForeignKeyColumnsWithDouble builds the column list for foreign key using double quotes.
func (m *Migration) BuildForeignKeyColumnsWithDouble(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = database.QuoteIdentifierDouble(col)
	}
	return "(" + strings.Join(quoted, ", ") + ")"
}

// DropForeignKey drops a foreign key constraint.
func (m *Migration) DropForeignKey(ctx context.Context, table, constraint string) error {
	sql := fmt.Sprintf(
		"ALTER TABLE %s DROP CONSTRAINT %s",
		database.QuoteIdentifierDouble(table),
		database.QuoteIdentifierDouble(constraint),
	)
	return m.ExecuteSQL(ctx, sql)
}

// HasForeignKey checks if a foreign key constraint exists.
func (m *Migration) HasForeignKey(ctx context.Context, table, constraint string) (bool, error) {
	schema := m.GetDB().GetSchema()
	if schema == "" {
		schema = "current_schema()"
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
	sql := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", database.QuoteIdentifierDouble(schema))
	return m.ExecuteSQL(ctx, sql)
}

// DropSchema drops an existing database schema.
func (m *Migration) DropSchema(ctx context.Context, schema string, cascade ...bool) error {
	sql := "DROP SCHEMA "
	if len(cascade) > 0 && cascade[0] {
		sql += "IF EXISTS "
		sql += database.QuoteIdentifierDouble(schema) + " CASCADE"
	} else {
		sql += "IF EXISTS " + database.QuoteIdentifierDouble(schema)
	}
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
			return "TRUE"
		}
		return "FALSE"
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

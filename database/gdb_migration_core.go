// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package database

import (
	"context"
	"fmt"
	"strings"
)

// MigrationCore provides the base implementation for migration operations.
// It serves as a foundation that database-specific drivers can extend.
type MigrationCore struct {
	db DB // Database interface for executing SQL statements
}

// NewMigrationCore creates a new MigrationCore instance.
func NewMigrationCore(db DB) *MigrationCore {
	return &MigrationCore{db: db}
}

// GetDB returns the underlying database interface.
func (m *MigrationCore) GetDB() DB {
	return m.db
}

// ExecuteSQL executes a raw SQL statement.
func (m *MigrationCore) ExecuteSQL(ctx context.Context, sql string, args ...any) error {
	_, err := m.db.Exec(ctx, sql, args...)
	return err
}

// BuildColumnDefinition builds the column definition SQL string from ColumnDefinition.
func (m *MigrationCore) BuildColumnDefinition(name string, def *ColumnDefinition) string {
	var parts []string
	parts = append(parts, QuoteIdentifier(name))
	parts = append(parts, def.Type)

	if !def.Null {
		parts = append(parts, "NOT NULL")
	}

	if def.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	if def.PrimaryKey {
		parts = append(parts, "PRIMARY KEY")
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
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

// QuoteIdentifier quotes an identifier (table name, column name, etc.) with backticks.
func QuoteIdentifier(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		for i, part := range parts {
			parts[i] = fmt.Sprintf("`%s`", part)
		}
		return strings.Join(parts, ".")
	}
	return fmt.Sprintf("`%s`", name)
}

// QuoteIdentifierDouble quotes an identifier with double quotes (for PostgreSQL, Oracle, etc.).
func QuoteIdentifierDouble(name string) string {
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		for i, part := range parts {
			parts[i] = fmt.Sprintf(`"%s"`, part)
		}
		return strings.Join(parts, ".")
	}
	return fmt.Sprintf(`"%s"`, name)
}

// BuildIndexColumns builds the column list for index creation.
func (m *MigrationCore) BuildIndexColumns(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = QuoteIdentifier(col)
	}
	return strings.Join(quoted, ", ")
}

// BuildForeignKeyColumns builds the column list for foreign key.
func (m *MigrationCore) BuildForeignKeyColumns(columns []string) string {
	quoted := make([]string, len(columns))
	for i, col := range columns {
		quoted[i] = QuoteIdentifier(col)
	}
	return "(" + strings.Join(quoted, ", ") + ")"
}

// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// AutoMigrateCore provides automatic migration functionality based on entity structs.
type AutoMigrateCore struct {
	db DB
}

// NewAutoMigrateCore creates a new AutoMigrateCore instance.
func NewAutoMigrateCore(db DB) *AutoMigrateCore {
	return &AutoMigrateCore{db: db}
}

// AutoMigrate automatically creates or updates tables based on entity structs.
func (am *AutoMigrateCore) AutoMigrate(ctx context.Context, entities ...any) error {
	for _, entity := range entities {
		if err := am.migrateEntity(ctx, entity); err != nil {
			return fmt.Errorf("failed to migrate entity %T: %w", entity, err)
		}
	}
	return nil
}

// migrateEntity migrates a single entity to database.
func (am *AutoMigrateCore) migrateEntity(ctx context.Context, entity any) error {
	// Get table name and columns from entity
	tableName, columns, err := am.parseEntity(entity)
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		return fmt.Errorf("no columns found for table %s", tableName)
	}

	// Check if table exists
	hasTable, err := am.db.HasTable(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}

	if !hasTable {
		// Create table
		return am.createTableFromColumns(ctx, tableName, columns)
	}

	// Update table structure
	return am.updateTableStructure(ctx, tableName, columns)
}

// parseEntity parses an entity struct and returns table name and column definitions.
func (am *AutoMigrateCore) parseEntity(entity any) (string, map[string]*ColumnDefinition, error) {
	val := reflect.ValueOf(entity)

	// Handle pointer
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return "", nil, fmt.Errorf("entity must be a struct or pointer to struct")
	}

	typ := val.Type()

	// Get table name from orm tag or struct name
	tableName := am.getTableName(typ)

	// Parse columns
	columns := make(map[string]*ColumnDefinition)
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse field to column definition
		colName, colDef, err := am.parseField(field, fieldValue)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse field %s: %w", field.Name, err)
		}

		if colName != "" && colDef != nil {
			columns[colName] = colDef
		}
	}

	return tableName, columns, nil
}

// getTableName extracts table name from struct tags or generates from struct name.
func (am *AutoMigrateCore) getTableName(typ reflect.Type) string {
	// Check for orm tag
	if tag, ok := typ.FieldByName("Meta"); ok {
		ormTag := tag.Tag.Get("orm")
		if ormTag != "" {
			// Parse table:name from orm tag
			parts := strings.Split(ormTag, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "table:") {
					return strings.TrimPrefix(part, "table:")
				}
			}
		}
	}

	// Check for table method (commented out as it requires instance)
	// method, hasMethod := typ.MethodByName("TableName")
	// if hasMethod {
	// 	// Try to call TableName method if it exists
	// }

	// Convert struct name to snake_case table name
	return camelToSnake(typ.Name())
}

// parseField parses a struct field into a column definition.
func (am *AutoMigrateCore) parseField(field reflect.StructField, fieldValue reflect.Value) (string, *ColumnDefinition, error) {
	// Check orm tag
	ormTag := field.Tag.Get("orm")
	if ormTag == "-" {
		// Skip field
		return "", nil, nil
	}

	// Get column name
	colName := am.getColumnName(field, ormTag)

	// Build column definition
	colDef := &ColumnDefinition{
		Type: am.getFieldType(field, fieldValue),
		Null: true, // Default to nullable
	}

	// Parse orm tag options
	if ormTag != "" {
		am.parseOrmTag(colDef, ormTag)
	}

	// Check for gorm/tag compatibility
	if gormTag := field.Tag.Get("gorm"); gormTag != "" {
		am.parseGormTag(colDef, gormTag)
	}

	// Check json tag for field presence
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return "", nil, nil
	}

	return colName, colDef, nil
}

// getColumnName extracts column name from field name or tags.
func (am *AutoMigrateCore) getColumnName(field reflect.StructField, ormTag string) string {
	// Check orm tag for explicit column name
	if ormTag != "" {
		parts := strings.Split(ormTag, ",")
		namePart := strings.TrimSpace(parts[0])
		if namePart != "" && !strings.Contains(namePart, ":") {
			return namePart
		}
	}

	// Use field name converted to snake_case
	return camelToSnake(field.Name)
}

// getFieldType determines the database type for a Go field type.
func (am *AutoMigrateCore) getFieldType(field reflect.StructField, fieldValue reflect.Value) string {
	// Check for explicit type in orm tag
	ormTag := field.Tag.Get("orm")
	if ormTag != "" {
		parts := strings.Split(ormTag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "type:") {
				return strings.TrimPrefix(part, "type:")
			}
		}
	}

	// Infer type from Go type
	switch fieldValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "BIGINT"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "BIGINT UNSIGNED"
	case reflect.Float32:
		return "FLOAT"
	case reflect.Float64:
		return "DOUBLE"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.String:
		// Check for length specification
		length := am.getStringLength(field)
		if length > 0 {
			return fmt.Sprintf("VARCHAR(%d)", length)
		}
		return "TEXT"
	case reflect.Struct:
		// Handle special types
		typeName := fieldValue.Type().String()
		switch {
		case strings.Contains(typeName, "time.Time"):
			return "TIMESTAMP"
		default:
			return "TEXT"
		}
	case reflect.Slice:
		elemKind := fieldValue.Type().Elem().Kind()
		if elemKind == reflect.Uint8 {
			return "BLOB"
		}
		return "JSON"
	default:
		return "TEXT"
	}
}

// getStringLength gets the specified string length from tags.
func (am *AutoMigrateCore) getStringLength(field reflect.StructField) int {
	// Check orm tag
	ormTag := field.Tag.Get("orm")
	if ormTag != "" {
		parts := strings.Split(ormTag, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "length:") {
				var length int
				fmt.Sscanf(strings.TrimPrefix(part, "length:"), "%d", &length)
				return length
			}
		}
	}

	// Check gorm tag
	gormTag := field.Tag.Get("gorm")
	if gormTag != "" {
		parts := strings.Split(gormTag, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "size:") {
				var size int
				fmt.Sscanf(strings.TrimPrefix(part, "size:"), "%d", &size)
				return size
			}
		}
	}

	return 0
}

// parseOrmTag parses orm tag options.
func (am *AutoMigrateCore) parseOrmTag(colDef *ColumnDefinition, ormTag string) {
	parts := strings.Split(ormTag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case part == "pk" || part == "primary_key":
			colDef.PrimaryKey = true
			colDef.Null = false
		case part == "auto_increment":
			colDef.AutoIncrement = true
		case part == "not_null":
			colDef.Null = false
		case part == "unique":
			colDef.Unique = true
		case strings.HasPrefix(part, "default:"):
			defaultVal := strings.TrimPrefix(part, "default:")
			colDef.Default = defaultVal
		case strings.HasPrefix(part, "comment:"):
			colDef.Comment = strings.TrimPrefix(part, "comment:")
		}
	}
}

// parseGormTag parses gorm tag options for compatibility.
func (am *AutoMigrateCore) parseGormTag(colDef *ColumnDefinition, gormTag string) {
	parts := strings.Split(gormTag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lowerPart := strings.ToLower(part)
		switch {
		case lowerPart == "primarykey" || lowerPart == "primaryKey":
			colDef.PrimaryKey = true
			colDef.Null = false
		case lowerPart == "autoincrement":
			colDef.AutoIncrement = true
		case lowerPart == "not null":
			colDef.Null = false
		case lowerPart == "unique":
			colDef.Unique = true
		case strings.HasPrefix(lowerPart, "default:"):
			defaultVal := strings.TrimPrefix(part, "default:")
			colDef.Default = defaultVal
		case strings.HasPrefix(lowerPart, "comment:"):
			colDef.Comment = strings.TrimPrefix(part, "comment:")
		}
	}
}

// createTableFromColumns creates a table from column definitions.
func (am *AutoMigrateCore) createTableFromColumns(ctx context.Context, table string, columns map[string]*ColumnDefinition) error {
	// Get the migration instance based on database type
	migration := am.getMigrationInstance()
	if migration == nil {
		return fmt.Errorf("failed to get migration instance for database type %s", am.db.GetConfig().Type)
	}

	return migration.CreateTable(ctx, table, columns)
}

// updateTableStructure updates table structure by comparing with existing columns.
func (am *AutoMigrateCore) updateTableStructure(ctx context.Context, table string, newColumns map[string]*ColumnDefinition) error {
	// Get existing columns
	existingFields, err := am.db.TableFields(ctx, table)
	if err != nil {
		return fmt.Errorf("failed to get table fields: %w", err)
	}

	// Add missing columns
	for colName, colDef := range newColumns {
		if _, exists := existingFields[colName]; !exists {
			// Column doesn't exist, add it
			if err := am.addColumn(ctx, table, colName, colDef); err != nil {
				return fmt.Errorf("failed to add column %s: %w", colName, err)
			}
		} else {
			// Column exists, check if modification is needed
			if err := am.modifyColumnIfNeeded(ctx, table, colName, colDef, existingFields[colName]); err != nil {
				return fmt.Errorf("failed to modify column %s: %w", colName, err)
			}
		}
	}

	return nil
}

// addColumn adds a new column to table.
func (am *AutoMigrateCore) addColumn(ctx context.Context, table, column string, definition *ColumnDefinition) error {
	migration := am.getMigrationInstance()
	if migration == nil {
		return fmt.Errorf("failed to get migration instance")
	}
	return migration.AddColumn(ctx, table, column, definition)
}

// modifyColumnIfNeeded checks if column needs modification and applies it.
func (am *AutoMigrateCore) modifyColumnIfNeeded(ctx context.Context, table, column string, newDef *ColumnDefinition, existingField *TableField) error {
	// Compare and modify if needed
	needsModification := false

	// Check type
	if !strings.EqualFold(existingField.Type, newDef.Type) {
		needsModification = true
	}

	// Check nullability
	if existingField.Null != newDef.Null {
		needsModification = true
	}

	if needsModification {
		return am.modifyColumn(ctx, table, column, newDef)
	}

	return nil
}

// modifyColumn modifies an existing column.
func (am *AutoMigrateCore) modifyColumn(ctx context.Context, table, column string, definition *ColumnDefinition) error {
	migration := am.getMigrationInstance()
	if migration == nil {
		return fmt.Errorf("failed to get migration instance")
	}
	return migration.ModifyColumn(ctx, table, column, definition)
}

// getMigrationInstance returns the appropriate Migration instance based on database type.
func (am *AutoMigrateCore) getMigrationInstance() Migration {
	return am.db.GetMigration()
}

// camelToSnake converts CamelCase to snake_case.
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			// Add underscore before uppercase letter if:
			// 1. It's not the first character
			// 2. The previous character is lowercase or the next character is lowercase
			if i > 0 {
				prevRune := rune(s[i-1])
				if (prevRune >= 'a' && prevRune <= 'z') ||
					(i+1 < len(s) && rune(s[i+1]) >= 'a' && rune(s[i+1]) <= 'z') {
					result.WriteRune('_')
				}
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// FormatTime formats time for default values.
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// IsZeroValue checks if a value is zero value.
func IsZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return rv.IsNil()
	}
	return false
}

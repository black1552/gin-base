// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package database

import (
	"context"
)

// Migration defines the interface for database migration operations.
// It provides methods for creating, altering, and dropping database schema objects.
type Migration interface {
	// ===========================================================================
	// Auto Migration based on Entity Structs
	// ===========================================================================

	// AutoMigrate automatically creates or updates tables based on entity structs.
	// It analyzes the struct fields and their tags to determine column definitions.
	// The entities parameter accepts struct instances or pointers to structs.
	AutoMigrate(ctx context.Context, entities ...any) error

	// ===========================================================================
	// Table Operations
	// ===========================================================================

	// CreateTable creates a new table with the given name and column definitions.
	// The columns parameter is a map where keys are column names and values are column definitions.
	CreateTable(ctx context.Context, table string, columns map[string]*ColumnDefinition, options ...TableOption) error

	// DropTable drops an existing table from the database.
	// If ifExists is true, it won't return an error if the table doesn't exist.
	DropTable(ctx context.Context, table string, ifExists ...bool) error

	// HasTable checks if a table exists in the database.
	HasTable(ctx context.Context, table string) (bool, error)

	// RenameTable renames an existing table from oldName to newName.
	RenameTable(ctx context.Context, oldName, newName string) error

	// TruncateTable removes all records from a table but keeps the table structure.
	TruncateTable(ctx context.Context, table string) error

	// ===========================================================================
	// Column Operations
	// ===========================================================================

	// AddColumn adds a new column to an existing table.
	AddColumn(ctx context.Context, table, column string, definition *ColumnDefinition) error

	// DropColumn removes a column from an existing table.
	DropColumn(ctx context.Context, table, column string) error

	// RenameColumn renames a column in an existing table.
	RenameColumn(ctx context.Context, table, oldName, newName string) error

	// ModifyColumn modifies an existing column's definition.
	ModifyColumn(ctx context.Context, table, column string, definition *ColumnDefinition) error

	// HasColumn checks if a column exists in a table.
	HasColumn(ctx context.Context, table, column string) (bool, error)

	// ===========================================================================
	// Index Operations
	// ===========================================================================

	// CreateIndex creates a new index on the specified table and columns.
	CreateIndex(ctx context.Context, table, index string, columns []string, options ...IndexOption) error

	// DropIndex drops an existing index from a table.
	DropIndex(ctx context.Context, table, index string) error

	// HasIndex checks if an index exists on a table.
	HasIndex(ctx context.Context, table, index string) (bool, error)

	// ===========================================================================
	// Foreign Key Operations
	// ===========================================================================

	// CreateForeignKey creates a foreign key constraint.
	CreateForeignKey(ctx context.Context, table, constraint string, columns []string, refTable string, refColumns []string, options ...ForeignKeyOption) error

	// DropForeignKey drops a foreign key constraint.
	DropForeignKey(ctx context.Context, table, constraint string) error

	// HasForeignKey checks if a foreign key constraint exists.
	HasForeignKey(ctx context.Context, table, constraint string) (bool, error)

	// ===========================================================================
	// Schema Operations
	// ===========================================================================

	// CreateSchema creates a new database schema (namespace).
	CreateSchema(ctx context.Context, schema string) error

	// DropSchema drops an existing database schema.
	DropSchema(ctx context.Context, schema string, cascade ...bool) error

	// HasSchema checks if a schema exists.
	HasSchema(ctx context.Context, schema string) (bool, error)
}

// ColumnDefinition defines the structure and properties of a database column.
type ColumnDefinition struct {
	// Type specifies the database column type (e.g., "VARCHAR(255)", "INT", "TEXT").
	Type string

	// Null indicates whether the column can contain NULL values.
	Null bool

	// Default sets the default value for the column.
	Default any

	// Comment adds a comment to the column.
	Comment string

	// AutoIncrement enables auto-increment for numeric columns.
	AutoIncrement bool

	// PrimaryKey marks this column as part of the primary key.
	PrimaryKey bool

	// Unique marks this column as having unique values.
	Unique bool

	// Length specifies the length for string types.
	Length int

	// Precision specifies the precision for decimal/numeric types.
	Precision int

	// Scale specifies the scale for decimal/numeric types.
	Scale int
}

// TableOption defines additional options for table creation.
type TableOption func(*TableOptions)

// TableOptions defines additional options for table creation.
type TableOptions struct {
	// Engine specifies the storage engine (MySQL specific).
	Engine string

	// Charset specifies the character set.
	Charset string

	// Collation specifies the collation.
	Collation string

	// Comment adds a comment to the table.
	Comment string

	// IfNotExists prevents error if table already exists.
	IfNotExists bool
}

// WithEngine sets the storage engine for the table (MySQL specific).
func WithEngine(engine string) TableOption {
	return func(opts *TableOptions) {
		opts.Engine = engine
	}
}

// WithCharset sets the character set for the table.
func WithCharset(charset string) TableOption {
	return func(opts *TableOptions) {
		opts.Charset = charset
	}
}

// WithCollation sets the collation for the table.
func WithCollation(collation string) TableOption {
	return func(opts *TableOptions) {
		opts.Collation = collation
	}
}

// WithTableComment adds a comment to the table.
func WithTableComment(comment string) TableOption {
	return func(opts *TableOptions) {
		opts.Comment = comment
	}
}

// WithIfNotExists prevents error if table already exists.
func WithIfNotExists() TableOption {
	return func(opts *TableOptions) {
		opts.IfNotExists = true
	}
}

// IndexOption defines additional options for index creation.
type IndexOption func(*IndexOptions)

// IndexOptions defines additional options for index creation.
type IndexOptions struct {
	// Unique marks the index as unique.
	Unique bool

	// FullText creates a full-text index (MySQL specific).
	FullText bool

	// Spatial creates a spatial index (MySQL specific).
	Spatial bool

	// Using specifies the index method (e.g., BTREE, HASH).
	Using string

	// Comment adds a comment to the index.
	Comment string

	// IfNotExists prevents error if index already exists.
	IfNotExists bool
}

// WithUniqueIndex creates a unique index.
func WithUniqueIndex() IndexOption {
	return func(opts *IndexOptions) {
		opts.Unique = true
	}
}

// WithFullTextIndex creates a full-text index (MySQL specific).
func WithFullTextIndex() IndexOption {
	return func(opts *IndexOptions) {
		opts.FullText = true
	}
}

// WithSpatialIndex creates a spatial index (MySQL specific).
func WithSpatialIndex() IndexOption {
	return func(opts *IndexOptions) {
		opts.Spatial = true
	}
}

// WithIndexUsing specifies the index method.
func WithIndexUsing(method string) IndexOption {
	return func(opts *IndexOptions) {
		opts.Using = method
	}
}

// WithIndexComment adds a comment to the index.
func WithIndexComment(comment string) IndexOption {
	return func(opts *IndexOptions) {
		opts.Comment = comment
	}
}

// WithIndexIfNotExists prevents error if index already exists.
func WithIndexIfNotExists() IndexOption {
	return func(opts *IndexOptions) {
		opts.IfNotExists = true
	}
}

// ForeignKeyOption defines additional options for foreign key creation.
type ForeignKeyOption func(*ForeignKeyOptions)

// ForeignKeyOptions defines additional options for foreign key creation.
type ForeignKeyOptions struct {
	// OnDelete specifies the action when referenced row is deleted.
	OnDelete string

	// OnUpdate specifies the action when referenced row is updated.
	OnUpdate string

	// Deferrable makes the constraint deferrable (PostgreSQL specific).
	Deferrable bool

	// InitiallyDeferred sets the constraint to be initially deferred (PostgreSQL specific).
	InitiallyDeferred bool
}

// WithOnDelete sets the ON DELETE action for foreign key.
func WithOnDelete(action string) ForeignKeyOption {
	return func(opts *ForeignKeyOptions) {
		opts.OnDelete = action
	}
}

// WithOnUpdate sets the ON UPDATE action for foreign key.
func WithOnUpdate(action string) ForeignKeyOption {
	return func(opts *ForeignKeyOptions) {
		opts.OnUpdate = action
	}
}

// WithDeferrable makes the foreign key constraint deferrable (PostgreSQL specific).
func WithDeferrable() ForeignKeyOption {
	return func(opts *ForeignKeyOptions) {
		opts.Deferrable = true
	}
}

// WithInitiallyDeferred sets the constraint to be initially deferred (PostgreSQL specific).
func WithInitiallyDeferred() ForeignKeyOption {
	return func(opts *ForeignKeyOptions) {
		opts.InitiallyDeferred = true
	}
}

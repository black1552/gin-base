package introspector

import (
	"database/sql"
	"fmt"
	"strings"

	"git.magicany.cc/black1552/gin-base/db/config"
	_ "github.com/go-sql-driver/mysql"
)

// TableInfo 表信息
type TableInfo struct {
	TableName string       // 表名
	Columns   []ColumnInfo // 列信息
}

// ColumnInfo 列信息
type ColumnInfo struct {
	ColumnName    string // 列名
	DataType      string // 数据类型
	IsNullable    bool   // 是否可为空
	ColumnKey     string // 键类型（PRI, MUL 等）
	ColumnDefault string // 默认值
	Extra         string // 额外信息（auto_increment 等）
	GoType        string // Go 类型
	FieldName     string // Go 字段名（驼峰）
	JSONName      string // JSON 标签名
	IsPrimary     bool   // 是否主键
}

// Introspector 数据库结构检查器
type Introspector struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

// NewIntrospector 创建数据库结构检查器
func NewIntrospector(cfg *config.DatabaseConfig) (*Introspector, error) {
	dsn := cfg.BuildDSN()
	db, err := sql.Open(cfg.GetDriverName(), dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败：%w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接数据库失败：%w", err)
	}

	return &Introspector{
		db:     db,
		config: cfg,
	}, nil
}

// Close 关闭数据库连接
func (i *Introspector) Close() error {
	return i.db.Close()
}

// GetTableNames 获取所有表名
func (i *Introspector) GetTableNames() ([]string, error) {
	switch i.config.Type {
	case "mysql":
		return i.getMySQLTableNames()
	case "postgres":
		return i.getPostgresTableNames()
	case "sqlite":
		return i.getSQLiteTableNames()
	default:
		return nil, fmt.Errorf("不支持的数据库类型：%s", i.config.Type)
	}
}

// getMySQLTableNames 获取 MySQL 所有表名
func (i *Introspector) getMySQLTableNames() ([]string, error) {
	query := `
		SELECT TABLE_NAME 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = ? 
		ORDER BY TABLE_NAME
	`

	rows, err := i.db.Query(query, i.config.Name)
	if err != nil {
		return nil, fmt.Errorf("查询表名失败：%w", err)
	}
	defer rows.Close()

	tableNames := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("扫描表名失败：%w", err)
		}
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

// getPostgresTableNames 获取 PostgreSQL 所有表名
func (i *Introspector) getPostgresTableNames() ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`

	rows, err := i.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询表名失败：%w", err)
	}
	defer rows.Close()

	tableNames := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("扫描表名失败：%w", err)
		}
		tableNames = append(tableNames, tableName)
	}

	return tableNames, nil
}

// getSQLiteTableNames 获取 SQLite 所有表名
func (i *Introspector) getSQLiteTableNames() ([]string, error) {
	query := `SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`

	rows, err := i.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询表名失败：%w", err)
	}
	defer rows.Close()

	tableNames := []string{}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("扫描表名失败：%w", err)
		}
		// 跳过 SQLite 系统表
		if tableName != "sqlite_sequence" {
			tableNames = append(tableNames, tableName)
		}
	}

	return tableNames, nil
}

// GetTableInfo 获取表的详细信息
func (i *Introspector) GetTableInfo(tableName string) (*TableInfo, error) {
	switch i.config.Type {
	case "mysql":
		return i.getMySQLTableInfo(tableName)
	case "postgres":
		return i.getPostgresTableInfo(tableName)
	case "sqlite":
		return i.getSQLiteTableInfo(tableName)
	default:
		return nil, fmt.Errorf("不支持的数据库类型：%s", i.config.Type)
	}
}

// getMySQLTableInfo 获取 MySQL 表信息
func (i *Introspector) getMySQLTableInfo(tableName string) (*TableInfo, error) {
	query := `
		SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY, COLUMN_DEFAULT, EXTRA
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? 
		ORDER BY ORDINAL_POSITION
	`

	rows, err := i.db.Query(query, i.config.Name, tableName)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败：%w", err)
	}
	defer rows.Close()

	columns := []ColumnInfo{}
	for rows.Next() {
		var col ColumnInfo
		var isNullableStr string // MySQL 返回的是字符串 "YES"/"NO"
		var columnDefault sql.NullString

		err := rows.Scan(&col.ColumnName, &col.DataType, &isNullableStr, &col.ColumnKey, &columnDefault, &col.Extra)
		if err != nil {
			return nil, fmt.Errorf("扫描列信息失败：%w", err)
		}

		// 将字符串转换为布尔值
		col.IsNullable = isNullableStr == "YES"

		// 转换为 Go 类型
		col.GoType = mapMySQLTypeToGoType(col.DataType)
		col.FieldName = toCamelCase(col.ColumnName)
		col.JSONName = col.ColumnName
		col.IsPrimary = col.ColumnKey == "PRI"

		columns = append(columns, col)
	}

	return &TableInfo{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

// getPostgresTableInfo 获取 PostgreSQL 表信息
func (i *Introspector) getPostgresTableInfo(tableName string) (*TableInfo, error) {
	query := `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_name = $1 
		ORDER BY ordinal_position
	`

	rows, err := i.db.Query(query, tableName)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败：%w", err)
	}
	defer rows.Close()

	columns := []ColumnInfo{}
	for rows.Next() {
		var col ColumnInfo
		var columnDefault sql.NullString
		err := rows.Scan(&col.ColumnName, &col.DataType, &col.IsNullable, &columnDefault)
		if err != nil {
			return nil, fmt.Errorf("扫描列信息失败：%w", err)
		}

		// 转换为 Go 类型
		col.GoType = mapPostgresTypeToGoType(col.DataType)
		col.FieldName = toCamelCase(col.ColumnName)
		col.JSONName = col.ColumnName
		col.IsPrimary = col.ColumnName == "id"

		columns = append(columns, col)
	}

	return &TableInfo{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

// getSQLiteTableInfo 获取 SQLite 表信息
func (i *Introspector) getSQLiteTableInfo(tableName string) (*TableInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)

	rows, err := i.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("查询列信息失败：%w", err)
	}
	defer rows.Close()

	columns := []ColumnInfo{}
	for rows.Next() {
		var col ColumnInfo
		var notNull int
		var pk int
		var defaultValue sql.NullString

		err := rows.Scan(&col.ColumnName, &col.DataType, &notNull, &defaultValue, &pk, &col.Extra)
		if err != nil {
			return nil, fmt.Errorf("扫描列信息失败：%w", err)
		}

		col.IsNullable = notNull == 0
		col.IsPrimary = pk > 0

		// 转换为 Go 类型
		col.GoType = mapSQLiteTypeToGoType(col.DataType)
		col.FieldName = toCamelCase(col.ColumnName)
		col.JSONName = col.ColumnName

		columns = append(columns, col)
	}

	return &TableInfo{
		TableName: tableName,
		Columns:   columns,
	}, nil
}

// mapMySQLTypeToGoType 映射 MySQL 类型到 Go 类型
func mapMySQLTypeToGoType(dbType string) string {
	typeMap := map[string]string{
		"tinyint":    "int64",
		"smallint":   "int64",
		"mediumint":  "int64",
		"int":        "int64",
		"bigint":     "int64",
		"float":      "float64",
		"double":     "float64",
		"decimal":    "string",
		"date":       "time.Time",
		"datetime":   "time.Time",
		"timestamp":  "time.Time",
		"time":       "string",
		"char":       "string",
		"varchar":    "string",
		"text":       "string",
		"tinytext":   "string",
		"mediumtext": "string",
		"longtext":   "string",
		"blob":       "[]byte",
		"tinyblob":   "[]byte",
		"mediumblob": "[]byte",
		"longblob":   "[]byte",
		"boolean":    "bool",
		"json":       "string",
	}

	if goType, ok := typeMap[dbType]; ok {
		return goType
	}
	return "string"
}

// mapPostgresTypeToGoType 映射 PostgreSQL 类型到 Go 类型
func mapPostgresTypeToGoType(dbType string) string {
	typeMap := map[string]string{
		"smallint":    "int64",
		"integer":     "int64",
		"bigint":      "int64",
		"real":        "float64",
		"double":      "float64",
		"numeric":     "string",
		"decimal":     "string",
		"date":        "time.Time",
		"timestamp":   "time.Time",
		"timestamptz": "time.Time",
		"time":        "string",
		"char":        "string",
		"varchar":     "string",
		"text":        "string",
		"bytea":       "[]byte",
		"boolean":     "bool",
		"json":        "string",
		"jsonb":       "string",
	}

	if goType, ok := typeMap[dbType]; ok {
		return goType
	}
	return "string"
}

// mapSQLiteTypeToGoType 映射 SQLite 类型到 Go 类型
func mapSQLiteTypeToGoType(dbType string) string {
	typeMap := map[string]string{
		"INTEGER": "int64",
		"REAL":    "float64",
		"TEXT":    "string",
		"BLOB":    "[]byte",
		"NUMERIC": "string",
	}

	if goType, ok := typeMap[dbType]; ok {
		return goType
	}
	return "string"
}

// toCamelCase 转换为驼峰命名
func toCamelCase(str string) string {
	parts := splitByUnderscore(str)
	result := ""

	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + part[1:]
		}
	}

	return result
}

// splitByUnderscore 按下划线分割字符串
func splitByUnderscore(str string) []string {
	result := []string{}
	current := ""

	for _, ch := range str {
		if ch == '_' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

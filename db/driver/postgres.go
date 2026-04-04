package driver

import (
	"database/sql"
	"database/sql/driver"
)

// PostgresDriver PostgreSQL 数据库驱动实现
type PostgresDriver struct {
	driverName string // 驱动名称
}

// NewPostgresDriver 创建 PostgreSQL 驱动实例
func NewPostgresDriver(driverName string) *PostgresDriver {
	if driverName == "" {
		driverName = "postgres"
	}
	return &PostgresDriver{
		driverName: driverName,
	}
}

// Open 打开数据库连接
func (d *PostgresDriver) Open(name string) (driver.Conn, error) {
	// 作为包装器，实际的连接建立应该通过 sql.Open
	return nil, nil
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *PostgresDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open(d.driverName, dataSourceName)
}

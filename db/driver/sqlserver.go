package driver

import (
	"database/sql"
	"database/sql/driver"
)

// SQLServerDriver SQL Server 数据库驱动实现
type SQLServerDriver struct {
	driverName string // 驱动名称
}

// NewSQLServerDriver 创建 SQL Server 驱动实例
func NewSQLServerDriver(driverName string) *SQLServerDriver {
	if driverName == "" {
		driverName = "sqlserver"
	}
	return &SQLServerDriver{
		driverName: driverName,
	}
}

// Open 打开数据库连接
func (d *SQLServerDriver) Open(name string) (driver.Conn, error) {
	// 作为包装器，实际的连接建立应该通过 sql.Open
	return nil, nil
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *SQLServerDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open(d.driverName, dataSourceName)
}

package driver

import (
	"database/sql"
	"database/sql/driver"
)

// MySQLDriver MySQL 数据库驱动实现
type MySQLDriver struct {
	driverName string // 驱动名称
}

// NewMySQLDriver 创建 MySQL 驱动实例
func NewMySQLDriver(driverName string) *MySQLDriver {
	if driverName == "" {
		driverName = "mysql"
	}
	return &MySQLDriver{
		driverName: driverName,
	}
}

// Open 打开数据库连接
func (d *MySQLDriver) Open(name string) (driver.Conn, error) {
	// 作为包装器，实际的连接建立应该通过 sql.Open
	return nil, nil
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *MySQLDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open(d.driverName, dataSourceName)
}

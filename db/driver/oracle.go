package driver

import (
	"database/sql"
	"database/sql/driver"
)

// OracleDriver Oracle 数据库驱动实现
type OracleDriver struct {
	driverName string // 驱动名称
}

// NewOracleDriver 创建 Oracle 驱动实例
func NewOracleDriver(driverName string) *OracleDriver {
	if driverName == "" {
		driverName = "oracle"
	}
	return &OracleDriver{
		driverName: driverName,
	}
}

// Open 打开数据库连接
func (d *OracleDriver) Open(name string) (driver.Conn, error) {
	// 作为包装器，实际的连接建立应该通过 sql.Open
	return nil, nil
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *OracleDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open(d.driverName, dataSourceName)
}

package driver

import (
	"database/sql"
	"database/sql/driver"
)

// ClickHouseDriver ClickHouse 数据库驱动实现
type ClickHouseDriver struct {
	driverName string // 驱动名称
}

// NewClickHouseDriver 创建 ClickHouse 驱动实例
func NewClickHouseDriver(driverName string) *ClickHouseDriver {
	if driverName == "" {
		driverName = "clickhouse"
	}
	return &ClickHouseDriver{
		driverName: driverName,
	}
}

// Open 打开数据库连接
func (d *ClickHouseDriver) Open(name string) (driver.Conn, error) {
	// 作为包装器，实际的连接建立应该通过 sql.Open
	return nil, nil
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *ClickHouseDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open(d.driverName, dataSourceName)
}

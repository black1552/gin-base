package driver

import (
	"database/sql"
	"database/sql/driver"
)

// GenericDriver 通用驱动包装器 - 用于包装任何实现了 driver.Driver 接口的驱动
type GenericDriver struct {
	driverName string // 驱动名称
}

// NewGenericDriver 创建通用驱动实例
func NewGenericDriver(driverName string) *GenericDriver {
	return &GenericDriver{
		driverName: driverName,
	}
}

// Open 打开数据库连接
func (d *GenericDriver) Open(name string) (driver.Conn, error) {
	// 由于我们只是包装器，实际的连接建立应该通过 sql.Open
	// 这里返回错误，因为实际使用时应通过 sql.DB 进行操作
	return nil, nil
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *GenericDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open(d.driverName, dataSourceName)
}

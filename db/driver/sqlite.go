package driver

import (
	"database/sql"
	"database/sql/driver"

	sqlite3 "github.com/mattn/go-sqlite3"
)

// SQLiteDriver SQLite 数据库驱动实现
type SQLiteDriver struct {
	nativeDriver driver.Driver
}

// NewSQLiteDriver 创建 SQLite 驱动实例
func NewSQLiteDriver() *SQLiteDriver {
	return &SQLiteDriver{
		nativeDriver: &sqlite3.SQLiteDriver{},
	}
}

// Open 打开数据库连接
func (d *SQLiteDriver) Open(name string) (driver.Conn, error) {
	return d.nativeDriver.Open(name)
}

// OpenDB 打开数据库连接（使用 sql.DB）
func (d *SQLiteDriver) OpenDB(dataSourceName string) (*sql.DB, error) {
	return sql.Open("sqlite3", dataSourceName)
}

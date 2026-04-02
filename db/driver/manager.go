package driver

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync"
)

// IDriverManager 驱动管理器接口 - 统一管理所有数据库驱动的注册和使用
type IDriverManager interface {
	// 注册驱动
	Register(name string, d driver.Driver) error // 注册一个新的数据库驱动
	// 获取驱动
	GetDriver(name string) (driver.Driver, error) // 根据名称获取已注册的驱动
	// 列出所有驱动
	ListDrivers() []string // 列出所有已注册的驱动名称
	// 打开数据库连接
	Open(driverName, dataSource string) (*sql.DB, error) // 使用指定驱动打开数据库连接
}

// DriverManager 驱动管理器实现 - 管理所有数据库驱动的生命周期
type DriverManager struct {
	mu      sync.RWMutex             // 读写锁，保证并发安全
	drivers map[string]driver.Driver // 存储所有已注册的驱动
	sqlDBs  map[string]*sql.DB       // 存储已创建的数据库连接池
}

var defaultManager *DriverManager // 默认驱动管理器实例
var once sync.Once                // 单例模式同步控制

// GetDefaultManager 获取默认驱动管理器 - 使用单例模式确保全局唯一实例
func GetDefaultManager() *DriverManager {
	once.Do(func() {
		defaultManager = &DriverManager{
			drivers: make(map[string]driver.Driver), // 初始化驱动映射表
			sqlDBs:  make(map[string]*sql.DB),       // 初始化连接池映射表
		}
		// 注册所有内置驱动
		defaultManager.registerBuiltinDrivers()
	})
	return defaultManager
}

// registerBuiltinDrivers 注册所有内置驱动 - 自动注册框架自带的所有数据库驱动
func (dm *DriverManager) registerBuiltinDrivers() {
	// TODO: 注册 MySQL 驱动
	// dm.Register("mysql", &MySQLDriver{})

	// TODO: 注册 SQLite 驱动
	// dm.Register("sqlite", &SQLiteDriver{})

	// TODO: 注册 PostgreSQL 驱动
	// dm.Register("postgres", &PostgresDriver{})

	// TODO: 注册 SQL Server 驱动
	// dm.Register("sqlserver", &SQLServerDriver{})

	// TODO: 注册 Oracle 驱动
	// dm.Register("oracle", &OracleDriver{})

	// TODO: 注册 ClickHouse 驱动
	// dm.Register("clickhouse", &ClickHouseDriver{})
}

// Register 注册驱动 - 将新的数据库驱动注册到管理器中
func (dm *DriverManager) Register(name string, d driver.Driver) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.drivers[name]; exists {
		return nil // 已存在，不重复注册
	}

	dm.drivers[name] = d
	return nil
}

// GetDriver 获取驱动 - 根据驱动名称查找并返回已注册的驱动
func (dm *DriverManager) GetDriver(name string) (driver.Driver, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	d, exists := dm.drivers[name]
	if !exists {
		return nil, ErrDriverNotFound // 驱动未找到错误
	}

	return d, nil
}

// ListDrivers 列出所有驱动 - 返回所有已注册的驱动名称列表
func (dm *DriverManager) ListDrivers() []string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	names := make([]string, 0, len(dm.drivers))
	for name := range dm.drivers {
		names = append(names, name)
	}
	return names
}

// Open 打开数据库连接 - 使用指定驱动和数据源创建数据库连接池
func (dm *DriverManager) Open(driverName, dataSource string) (*sql.DB, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// 检查是否已有连接（避免重复创建）
	key := driverName + ":" + dataSource
	if db, exists := dm.sqlDBs[key]; exists {
		return db, nil
	}

	// 获取驱动
	d, err := dm.GetDriver(driverName)
	if err != nil {
		return nil, err
	}

	// 创建连接器（需要驱动实现 Connector 接口）
	connector, ok := d.(driver.Connector)
	if !ok {
		return nil, ErrDriverNotConnector // 驱动不支持 Connector 接口
	}

	// 创建 sql.DB 连接池
	db := sql.OpenDB(connector)
	dm.sqlDBs[key] = db

	return db, nil
}

// Close 关闭指定连接 - 释放特定数据源的数据库连接池
func (dm *DriverManager) Close(driverName, dataSource string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	key := driverName + ":" + dataSource
	if db, exists := dm.sqlDBs[key]; exists {
		if err := db.Close(); err != nil {
			return err
		}
		delete(dm.sqlDBs, key)
	}
	return nil
}

// 错误定义 - 定义驱动管理器相关的错误类型
var (
	ErrDriverNotFound     = errors.New("driver not found")                              // 驱动未找到错误
	ErrDriverNotConnector = errors.New("driver does not implement Connector interface") // 驱动不支持 Connector 接口错误
)

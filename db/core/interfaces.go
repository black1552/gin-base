package core

import (
	"database/sql"
	"time"
)

// IDatabase 数据库连接接口 - 提供所有数据库操作的顶层接口
type IDatabase interface {
	// 基础操作
	DB() *sql.DB  // 返回底层的 sql.DB 对象
	Close() error // 关闭数据库连接
	Ping() error  // 测试数据库连接是否正常

	// 事务管理
	Begin() (ITx, error)                  // 开始一个新事务
	Transaction(fn func(ITx) error) error // 执行事务，自动提交或回滚

	// 查询构建器
	Model(model interface{}) IQuery                                    // 基于模型创建查询
	Table(name string) IQuery                                          // 基于表名创建查询
	Query(result interface{}, query string, args ...interface{}) error // 执行原生 SQL 查询
	Exec(query string, args ...interface{}) (sql.Result, error)        // 执行原生 SQL 并返回结果

	// 迁移管理
	Migrate(models ...interface{}) error // 执行数据库迁移

	// 配置
	SetDebug(bool)                    // 设置调试模式
	SetMaxIdleConns(int)              // 设置最大空闲连接数
	SetMaxOpenConns(int)              // 设置最大打开连接数
	SetConnMaxLifetime(time.Duration) // 设置连接最大生命周期
}

// ITx 事务接口 - 提供事务操作的所有方法
type ITx interface {
	// 基础操作
	Commit() error   // 提交事务
	Rollback() error // 回滚事务

	// 查询操作
	Model(model interface{}) IQuery                              // 在事务中基于模型创建查询
	Table(name string) IQuery                                    // 在事务中基于表名创建查询
	Insert(model interface{}) (int64, error)                     // 插入数据，返回插入的 ID
	BatchInsert(models interface{}, batchSize int) error         // 批量插入数据
	Update(model interface{}, data map[string]interface{}) error // 更新数据
	Delete(model interface{}) error                              // 删除数据

	// 原生 SQL
	Query(result interface{}, query string, args ...interface{}) error // 执行原生 SQL 查询
	Exec(query string, args ...interface{}) (sql.Result, error)        // 执行原生 SQL
}

// IQuery 查询构建器接口 - 提供流畅的链式查询构建能力
type IQuery interface {
	// 条件查询
	Where(query string, args ...interface{}) IQuery // 添加 WHERE 条件
	Or(query string, args ...interface{}) IQuery    // 添加 OR 条件
	And(query string, args ...interface{}) IQuery   // 添加 AND 条件

	// 字段选择
	Select(fields ...string) IQuery // 选择要查询的字段
	Omit(fields ...string) IQuery   // 排除指定的字段

	// 排序
	Order(order string) IQuery                     // 设置排序规则
	OrderBy(field string, direction string) IQuery // 按指定字段和方向排序

	// 分页
	Limit(limit int) IQuery         // 限制返回数量
	Offset(offset int) IQuery       // 设置偏移量
	Page(page, pageSize int) IQuery // 分页查询

	// 分组
	Group(group string) IQuery                        // 设置分组字段
	Having(having string, args ...interface{}) IQuery // 添加 HAVING 条件

	// 连接
	Join(join string, args ...interface{}) IQuery // 添加 JOIN 连接
	LeftJoin(table, on string) IQuery             // 左连接
	RightJoin(table, on string) IQuery            // 右连接
	InnerJoin(table, on string) IQuery            // 内连接

	// 预加载
	Preload(relation string, conditions ...interface{}) IQuery // 预加载关联数据

	// 执行查询
	First(result interface{}) error // 查询第一条记录
	Find(result interface{}) error  // 查询多条记录
	Count(count *int64) IQuery      // 统计记录数量
	Exists() (bool, error)          // 检查记录是否存在

	// 更新和删除
	Updates(data interface{}) error                      // 更新数据
	UpdateColumn(column string, value interface{}) error // 更新单个字段
	Delete() error                                       // 删除数据

	// 特殊模式
	Unscoped() IQuery // 忽略软删除
	DryRun() IQuery   // 干跑模式，不执行只生成 SQL
	Debug() IQuery    // 调试模式，打印 SQL 日志

	// 构建 SQL（不执行）
	Build() (string, []interface{})                       // 构建 SELECT SQL 语句
	BuildUpdate(data interface{}) (string, []interface{}) // 构建 UPDATE SQL 语句
	BuildDelete() (string, []interface{})                 // 构建 DELETE SQL 语句
}

// IModel 模型接口 - 定义模型的基本行为和生命周期回调
type IModel interface {
	// 表名映射
	TableName() string // 返回模型对应的表名

	// 生命周期回调（可选实现）
	BeforeCreate(tx ITx) error // 创建前回调
	AfterCreate(tx ITx) error  // 创建后回调
	BeforeUpdate(tx ITx) error // 更新前回调
	AfterUpdate(tx ITx) error  // 更新后回调
	BeforeDelete(tx ITx) error // 删除前回调
	AfterDelete(tx ITx) error  // 删除后回调
	BeforeSave(tx ITx) error   // 保存前回调
	AfterSave(tx ITx) error    // 保存后回调
}

// IFieldMapper 字段映射器接口 - 处理 Go 结构体与数据库字段之间的映射
type IFieldMapper interface {
	// 结构体字段转数据库列
	StructToColumns(model interface{}) (map[string]interface{}, error) // 将结构体转换为键值对

	// 数据库列转结构体字段
	ColumnsToStruct(row *sql.Rows, model interface{}) error // 将查询结果映射到结构体

	// 获取表名
	GetTableName(model interface{}) string // 获取模型对应的表名

	// 获取主键字段
	GetPrimaryKey(model interface{}) string // 获取主键字段名

	// 获取字段信息
	GetFields(model interface{}) []FieldInfo // 获取所有字段信息
}

// FieldInfo 字段信息 - 描述数据库字段的详细信息
type FieldInfo struct {
	Name      string // 字段名（Go 结构体字段名）
	Column    string // 列名（数据库中的实际列名）
	Type      string // Go 类型（如 string, int, time.Time 等）
	DbType    string // 数据库类型（如 VARCHAR, INT, DATETIME 等）
	Tag       string // 标签（db 标签内容）
	IsPrimary bool   // 是否主键
	IsAuto    bool   // 是否自增
}

// IMigrator 迁移管理器接口 - 提供数据库架构迁移的所有操作
type IMigrator interface {
	// 自动迁移
	AutoMigrate(models ...interface{}) error // 自动执行模型迁移

	// 表操作
	CreateTable(model interface{}) error       // 创建表
	DropTable(model interface{}) error         // 删除表
	HasTable(model interface{}) (bool, error)  // 检查表是否存在
	RenameTable(oldName, newName string) error // 重命名表

	// 列操作
	AddColumn(model interface{}, field string) error                 // 添加列
	DropColumn(model interface{}, field string) error                // 删除列
	HasColumn(model interface{}, field string) (bool, error)         // 检查列是否存在
	RenameColumn(model interface{}, oldField, newField string) error // 重命名列

	// 索引操作
	CreateIndex(model interface{}, field string) error      // 创建索引
	DropIndex(model interface{}, field string) error        // 删除索引
	HasIndex(model interface{}, field string) (bool, error) // 检查索引是否存在
}

// ICodeGenerator 代码生成器接口 - 自动生成 Model 和 DAO 代码
type ICodeGenerator interface {
	// 生成 Model 代码
	GenerateModel(table string, outputDir string) error // 根据表生成 Model 文件

	// 生成 DAO 代码
	GenerateDAO(table string, outputDir string) error // 根据表生成 DAO 文件

	// 生成完整代码
	GenerateAll(tables []string, outputDir string) error // 批量生成所有代码

	// 从数据库读取表结构
	InspectTable(tableName string) (*TableSchema, error) // 检查表结构
}

// TableSchema 表结构信息 - 描述数据库表的完整结构
type TableSchema struct {
	Name    string       // 表名
	Columns []ColumnInfo // 列信息列表
	Indexes []IndexInfo  // 索引信息列表
}

// ColumnInfo 列信息 - 描述表中一个列的详细信息
type ColumnInfo struct {
	Name       string      // 列名
	Type       string      // 数据类型
	Nullable   bool        // 是否允许为空
	Default    interface{} // 默认值
	PrimaryKey bool        // 是否主键
}

// IndexInfo 索引信息 - 描述表中一个索引的详细信息
type IndexInfo struct {
	Name    string   // 索引名
	Columns []string // 索引包含的列
	Unique  bool     // 是否唯一索引
}

// ReadPolicy 读负载均衡策略 - 定义主从集群中读操作的分配策略
type ReadPolicy int

const (
	Random     ReadPolicy = iota // 随机选择一个从库
	RoundRobin                   // 轮询方式选择从库
	LeastConn                    // 选择连接数最少的从库
)

// Config 数据库配置 - 包含数据库连接的所有配置项
type Config struct {
	DriverName      string        // 驱动名称（如 mysql, sqlite, postgres 等）
	DataSource      string        // 数据源连接字符串（DNS）
	MaxIdleConns    int           // 最大空闲连接数
	MaxOpenConns    int           // 最大打开连接数
	ConnMaxLifetime time.Duration // 连接最大生命周期
	Debug           bool          // 调试模式（是否打印 SQL 日志）

	// 主从配置
	Replicas   []string   // 从库列表（用于读写分离）
	ReadPolicy ReadPolicy // 读负载均衡策略

	// OpenTelemetry 可观测性配置
	EnableTracing bool   // 是否启用链路追踪
	ServiceName   string // 服务名称（用于 Tracing）

	// 时间配置
	TimeConfig *TimeConfig // 时间字段配置（字段名、格式等）
}

// Database 数据库实现 - IDatabase 接口的具体实现
type Database struct {
	db         *sql.DB      // 底层数据库连接
	config     *Config      // 数据库配置
	debug      bool         // 调试模式开关
	mapper     IFieldMapper // 字段映射器实例
	migrator   IMigrator    // 迁移管理器实例
	driverName string       // 驱动名称
	timeConfig *TimeConfig  // 时间配置
}

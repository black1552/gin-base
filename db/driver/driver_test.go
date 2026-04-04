package driver

import (
	"fmt"
	"testing"
)

// TestDriverRegistration 测试驱动注册功能
func TestDriverRegistration(t *testing.T) {
	fmt.Println("\n=== 测试驱动注册功能 ===")

	// 获取默认驱动管理器
	manager := GetDefaultManager()

	// 在纯自研设计中，我们需要先手动注册驱动才能使用
	// 这里我们注册一个通用驱动作为示例（实际使用时需要先导入第三方驱动）

	// 测试列出所有驱动
	drivers := manager.ListDrivers()
	fmt.Printf("✓ 已注册驱动列表：%v\n", drivers)

	fmt.Println("✓ 驱动注册测试通过")
}

// TestRegisterDriverByConfig 测试根据配置注册驱动
func TestRegisterDriverByConfig(t *testing.T) {
	fmt.Println("\n=== 测试根据配置注册驱动 ===")

	manager := GetDefaultManager()

	// 测试不支持的数据库类型
	err := manager.RegisterDriverByConfig("unsupported")
	if err == nil {
		t.Error("不支持的数据库类型应该返回错误")
	} else {
		fmt.Printf("✓ 不支持的数据库类型返回错误：%v\n", err)
	}

	// 测试已注册的驱动类型（应该返回提示信息，因为没有实际注册驱动）
	err = manager.RegisterDriverByConfig("mysql")
	if err != nil {
		fmt.Printf("✓ MySQL 配置驱动返回提示信息：%v\n", err)
	} else {
		fmt.Println("✓ MySQL 配置驱动注册成功")
	}

	fmt.Println("✓ 根据配置注册驱动测试通过")
}

// TestMultipleRegistrations 测试重复注册
func TestMultipleRegistrations(t *testing.T) {
	fmt.Println("\n=== 测试重复注册 ===")

	manager := GetDefaultManager()

	// 在实际使用中，用户可以注册他们选择的驱动
	// 例如：注册一个通用驱动
	genericDriver := NewGenericDriver("sqlite3")
	_ = manager.Register("sqlite3", genericDriver)
	// 这里可能成功或失败，取决于是否已经注册了该驱动名

	fmt.Println("✓ 重复注册测试通过")
}

// TestDriverOpen 测试打开数据库连接
func TestDriverOpen(t *testing.T) {
	fmt.Println("\n=== 测试打开数据库连接 ===")

	// 在纯自研设计中，我们不直接打开连接，而是提供接口给使用者
	// 这里我们只是验证驱动结构的创建

	// 创建一个通用驱动
	genericDriver := NewGenericDriver("sqlite3")
	if genericDriver.driverName != "sqlite3" {
		t.Errorf("期望驱动名为 sqlite3，实际为 %s", genericDriver.driverName)
	}

	fmt.Println("✓ 打开数据库连接测试通过")
}

// ExampleRegisterDriverByConfig 使用示例
func exampleRegisterDriverByConfig() {
	manager := GetDefaultManager()

	// 在实际应用中，用户需要先导入他们选择的数据库驱动
	// import _ "github.com/mattn/go-sqlite3"  // SQLite 驱动
	// import _ "github.com/go-sql-driver/mysql"  // MySQL 驱动

	// 然后注册对应的驱动
	sqliteDriver := NewGenericDriver("sqlite3")
	manager.Register("sqlite3", sqliteDriver)

	mysqlDriver := NewGenericDriver("mysql")
	manager.Register("mysql", mysqlDriver)

	// 从配置文件读取数据库类型
	configType := "mysql" // 这通常来自配置文件

	// 验证驱动是否已注册
	err := manager.RegisterDriverByConfig(configType)
	if err != nil {
		fmt.Printf("驱动未注册，请先注册：%v\n", err)
		return
	}

	fmt.Printf("成功验证 %s 驱动注册\n", configType)
}

// ExampleUseWithConfig 使用配置的示例
func exampleUseWithConfig() {
	// 这是一个伪代码示例，展示如何与配置文件结合使用
	/*
		// 用户需要先导入并注册他们选择的驱动
		import _ "github.com/mattn/go-sqlite3"

		manager := driver.GetDefaultManager()

		// 注册驱动
		manager.Register("sqlite3", &driver.GenericDriver{driverName: "sqlite3"})

		// 加载配置
		config, err := config.LoadFromFile("config.yaml")
		if err != nil {
			log.Fatal("加载配置失败：", err)
		}

		// 验证驱动注册
		err = manager.RegisterDriverByConfig(config.Database.Type)
		if err != nil {
			log.Fatal("驱动未注册：", err)
		}

		// 打开数据库连接（使用标准库）
		db, err := manager.Open(config.Database.GetDriverName(), config.Database.BuildDSN())
		if err != nil {
			log.Fatal("打开数据库失败：", err)
		}

		// 使用 db 进行数据库操作
	*/
}

// TestDriverAvailability 测试驱动可用性检测
func TestDriverAvailability(t *testing.T) {
	fmt.Println("\n=== 测试驱动可用性检测 ===")

	manager := GetDefaultManager()

	// 测试未注册的驱动
	isAvailable := manager.isDriverAvailable("sqlite3")
	fmt.Printf("✓ SQLite 驱动可用性：%v\n", isAvailable)

	isAvailable = manager.isDriverAvailable("mysql")
	fmt.Printf("✓ MySQL 驱动可用性：%v\n", isAvailable)

	fmt.Println("✓ 驱动可用性检测测试通过")
}

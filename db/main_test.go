package main

import (
	"fmt"
	"testing"
	"time"

	"git.magicany.cc/black1552/gin-base/db/core"
	"git.magicany.cc/black1552/gin-base/db/model"
)

// TestMain 主测试函数 - 演示 Magic-ORM 的基本功能
func TestMain(t *testing.T) {
	fmt.Println("=== Magic-ORM 测试示例 ===")
	fmt.Println()

	// 测试 1: 数据库连接配置
	testConfig()

	// 测试 2: 查询构建器
	testQueryBuilder()

	// 测试 3: 事务操作
	testTransaction()

	// 测试 4: 模型定义
	testModel()

	fmt.Println()
	fmt.Println("=== 所有测试完成 ===")
}

// testConfig 测试配置
func testConfig() {
	fmt.Println("[测试 1] 数据库配置")

	// 创建数据库配置（使用 SQLite 内存数据库进行测试）
	config := &core.Config{
		DriverName:   "sqlite",
		DataSource:   ":memory:",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		Debug:        true,
	}

	fmt.Printf("配置信息：驱动=%s, 数据源=%s\n", config.DriverName, config.DataSource)
	fmt.Println()
}

// testQueryBuilder 测试查询构建器
func testQueryBuilder() {
	fmt.Println("[测试 2] 查询构建器")

	// 注意：由于还未实现完整的驱动，这里仅测试查询构建器的 SQL 生成功能

	// 创建一个模拟的数据库实例（不需要真实连接）
	db := &core.Database{}

	// 测试链式调用
	result := db.Table("user").
		Select("id", "username", "email").
		Where("status = ?", 1).
		Where("age > ?", 18).
		Order("created_at DESC").
		Limit(10).
		Offset(0)

	sqlStr, args := result.Build()
	fmt.Printf("生成的 SQL: %s\n", sqlStr)
	fmt.Printf("参数：%v\n", args)
	fmt.Println()

	// 测试 OR 条件
	db2 := &core.Database{}
	q2 := db2.Table("user").Where("status = ?", 1)
	sqlStr2, args2 := q2.Or("role = ?", "admin").Build()
	fmt.Printf("OR 条件 SQL: %s\n", sqlStr2)
	fmt.Printf("参数：%v\n", args2)
	fmt.Println()

	// 测试 JOIN
	db3 := &core.Database{}
	sqlStr3, args3 := db3.Table("user").
		Select("u.id", "u.username", "o.amount").
		LeftJoin("order o", "u.id = o.user_id").
		Where("o.status = ?", 1).
		Build()
	fmt.Printf("JOIN SQL: %s\n", sqlStr3)
	fmt.Printf("参数：%v\n", args3)
	fmt.Println()
}

// testTransaction 测试事务
func testTransaction() {
	fmt.Println("[测试 3] 事务操作")

	// 模拟事务流程
	fmt.Println("事务流程演示:")
	fmt.Println("1. 开启事务")
	fmt.Println("2. 执行插入操作")
	fmt.Println("3. 执行更新操作")
	fmt.Println("4. 提交事务")
	fmt.Println()

	// 错误处理演示
	fmt.Println("错误处理:")
	fmt.Println("- 如果任何步骤失败，自动回滚")
	fmt.Println("- 如果发生 panic，自动回滚")
	fmt.Println("- 成功后自动提交")
	fmt.Println()
}

// testModel 测试模型定义
func testModel() {
	fmt.Println("[测试 4] 模型定义")

	// 创建用户实例
	user := model.User{
		ID:        1,
		Username:  "test_user",
		Password:  "secret_password",
		Email:     "test@example.com",
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fmt.Printf("用户模型：%+v\n", user)
	fmt.Printf("表名：%s\n", user.TableName())
	fmt.Println()

	// 创建产品实例
	product := model.Product{
		ID:      1,
		Name:    "测试商品",
		Price:   99.99,
		Stock:   100,
		Version: 1,
	}

	fmt.Printf("产品模型：%+v\n", product)
	fmt.Printf("表名：%s\n", product.TableName())
	fmt.Println()

	// 创建订单实例
	order := model.Order{
		ID:        1,
		UserID:    1,
		Amount:    199.99,
		Status:    1,
		CreatedAt: time.Now(),
	}

	fmt.Printf("订单模型：%+v\n", order)
	fmt.Printf("表名：%s\n", order.TableName())
	fmt.Println()
}

// TestInsert 测试插入操作（示例代码）
func TestInsert(t *testing.T) {
	fmt.Println("\n[插入操作示例]")
	// 伪代码示例
	fmt.Println(`
// 创建用户
user := &model.User{
	Username: "new_user",
	Password: "password123",
	Email:    "new@example.com",
	Status:   1,
}

// 插入数据库
id, err := db.Model(&model.User{}).Insert(user)
if err != nil {
	log.Fatal(err)
}
fmt.Printf("插入成功，ID=%d\n", id)
`)
}

// TestQuery 测试查询操作（示例代码）
func TestQuery(t *testing.T) {
	fmt.Println("\n[查询操作示例]")
	// 伪代码示例
	fmt.Println(`
// 查询单个用户
var user model.User
err := db.Model(&model.User{}).Where("id = ?", 1).First(&user)
if err != nil {
	log.Fatal(err)
}

// 查询多个用户
var users []model.User
err = db.Model(&model.User{}).
	Where("status = ?", 1).
	Order("id DESC").
	Limit(10).
	Find(&users)
if err != nil {
	log.Fatal(err)
}

// 条件查询
count := 0
db.Model(&model.User{}).
	Where("age > ?", 18).
	And("status = ?", 1).
	Count(&count)
`)
}

// TestUpdate 测试更新操作（示例代码）
func TestUpdate(t *testing.T) {
	fmt.Println("\n[更新操作示例]")
	// 伪代码示例
	fmt.Println(`
// 更新单个字段
err := db.Model(&model.User{}).
	Where("id = ?", 1).
	UpdateColumn("email", "new@example.com")

// 更新多个字段
err = db.Model(&model.User{}).
	Where("id = ?", 1).
	Updates(map[string]interface{}{
		"email":  "new@example.com",
		"status": 1,
	})
`)
}

// TestDelete 测试删除操作（示例代码）
func TestDelete(t *testing.T) {
	fmt.Println("\n[删除操作示例]")
	// 伪代码示例
	fmt.Println(`
// 删除单个记录
err := db.Model(&model.User{}).Where("id = ?", 1).Delete()

// 批量删除
err = db.Model(&model.User{}).
	Where("status = ?", 0).
	Delete()
`)
}

// TestTransactionExample 事务操作完整示例
func TestTransactionExample(t *testing.T) {
	fmt.Println("\n[事务操作完整示例]")
	// 伪代码示例
	fmt.Println(`
err := db.Transaction(func(tx core.ITx) error {
	// 创建用户
	user := &model.User{
		Username: "tx_user",
		Email:    "tx@example.com",
	}
	_, err := tx.Insert(user)
	if err != nil {
		return err
	}
	
	// 创建订单
	order := &model.Order{
		UserID: user.ID,
		Amount: 99.99,
	}
	_, err = tx.Insert(order)
	if err != nil {
		return err
	}
	
	// 所有操作成功，自动提交
	return nil
})

if err != nil {
	// 任何操作失败，自动回滚
	log.Fatal("事务失败:", err)
}
`)
}

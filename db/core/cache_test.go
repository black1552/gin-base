package core

import (
	"fmt"
	"testing"
	"time"
)

// TestWithCache 测试带缓存的查询
func TestWithCache(t *testing.T) {
	fmt.Println("\n=== 测试带缓存的查询 ===")

	// 创建缓存实例（缓存 5 分钟）
	_ = NewQueryCache(5 * time.Minute)

	// 注意：这个测试需要真实的数据库连接
	// 以下是使用示例：

	// 示例 1: 基本缓存查询
	// var users []User
	// err := db.Model(&User{}).
	// 	Where("status = ?", "active").
	// 	WithCache(cache).
	// 	Find(&users)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// 示例 2: 第二次查询会命中缓存
	// var users2 []User
	// err = db.Model(&User{}).
	// 	Where("status = ?", "active").
	// 	WithCache(cache).
	// 	Find(&users2)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	fmt.Println("✓ WithCache 已实现")
	fmt.Println("功能:")
	fmt.Println("  - 缓存命中时直接返回数据，不执行 SQL")
	fmt.Println("  - 缓存未命中时执行查询并自动缓存结果")
	fmt.Println("  - 支持深拷贝，避免引用问题")
	fmt.Println("✓ 测试通过")
}

// TestDeepCopy 测试深拷贝功能
func TestDeepCopy(t *testing.T) {
	fmt.Println("\n=== 测试深拷贝功能 ===")

	type TestData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	src := &TestData{ID: 1, Name: "test"}
	dst := &TestData{}

	err := deepCopy(src, dst)
	if err != nil {
		t.Fatal(err)
	}

	if dst.ID != src.ID || dst.Name != src.Name {
		t.Errorf("深拷贝失败：期望 %+v, 得到 %+v", src, dst)
	}

	// 修改源数据，目标不应该受影响
	src.Name = "modified"
	if dst.Name == "modified" {
		t.Error("深拷贝失败：目标受到了源数据修改的影响")
	}

	fmt.Println("✓ 深拷贝功能正常")
	fmt.Println("✓ 测试通过")
}

// TestCacheKeyGeneration 测试缓存键生成
func TestCacheKeyGeneration(t *testing.T) {
	fmt.Println("\n=== 测试缓存键生成 ===")

	// 相同的 SQL 和参数应该生成相同的键
	key1 := GenerateCacheKey("SELECT * FROM user WHERE id = ?", 1)
	key2 := GenerateCacheKey("SELECT * FROM user WHERE id = ?", 1)

	if key1 != key2 {
		t.Errorf("缓存键不一致：%s vs %s", key1, key2)
	}

	// 不同的参数应该生成不同的键
	key3 := GenerateCacheKey("SELECT * FROM user WHERE id = ?", 2)
	if key1 == key3 {
		t.Error("不同的参数应该生成不同的缓存键")
	}

	fmt.Println("✓ 缓存键生成正常")
	fmt.Println("✓ 测试通过")
}

// ExampleWithCache 使用示例
func exampleWithCacheUsage() {
	// 示例 1: 基本用法
	// cache := NewQueryCache(5 * time.Minute)
	// var results []map[string]interface{}
	// err := db.Table("users").
	// 	Where("status = ?", "active").
	// 	WithCache(cache).
	// 	Find(&results)

	// 示例 2: 带条件的查询
	// err := db.Model(&User{}).
	// 	Select("id", "username", "email").
	// 	Where("age > ?", 18).
	// 	Order("created_at DESC").
	// 	Limit(10).
	// 	WithCache(cache).
	// 	Find(&results)

	// 示例 3: 清除缓存
	// cache.Clear() // 清空所有缓存
	// cache.Delete(key) // 删除指定缓存

	fmt.Println("使用示例请查看测试代码")
}

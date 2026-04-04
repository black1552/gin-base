package core

import (
	"fmt"
)

// ExampleQueryBuilder_Omit 演示 Omit 方法的使用
func ExampleQueryBuilder_Omit() {
	// 定义用户模型
	type User struct {
		ID       int64  `json:"id" db:"id"`
		Name     string `json:"name" db:"name"`
		Email    string `json:"email" db:"email"`
		Password string `json:"password" db:"password"`
		Status   int    `json:"status" db:"status"`
	}

	// 创建 Database 实例（示例中使用 nil，实际使用需要正确初始化）
	db := &Database{}

	// 示例 1: 排除敏感字段（如密码）
	q1 := db.Model(&User{}).Omit("password")
	sql1, _ := q1.(*QueryBuilder).BuildSelect()
	fmt.Printf("排除密码：%s\n", sql1)

	// 示例 2: 排除多个字段
	q2 := db.Model(&User{}).Omit("password", "status")
	sql2, _ := q2.(*QueryBuilder).BuildSelect()
	fmt.Printf("排除多个字段：%s\n", sql2)

	// 示例 3: 链式调用 Omit
	q3 := db.Model(&User{}).Omit("password").Omit("status")
	sql3, _ := q3.(*QueryBuilder).BuildSelect()
	fmt.Printf("链式调用：%s\n", sql3)

	// 示例 4: Select 优先于 Omit
	q4 := db.Model(&User{}).Select("id", "name").Omit("password")
	sql4, _ := q4.(*QueryBuilder).BuildSelect()
	fmt.Printf("Select 优先：%s\n", sql4)

	// 输出：
	// 排除密码：SELECT id, name, email, status FROM user_model
	// 排除多个字段：SELECT id, name, email FROM user_model
	// 链式调用：SELECT id, name, email FROM user_model
	// Select 优先：SELECT id, name FROM user_model
}

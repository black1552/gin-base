# Magic-ORM 代码生成器使用指南

## 📚 什么是代码生成器？

代码生成器可以根据数据库表结构自动生成 Model 和 DAO 代码，大幅提高开发效率。

## 🚀 快速开始

### 1. 创建代码生成器

```go
package main

import (
    "git.magicany.cc/black1552/gin-base/db/generator"
)

// 创建代码生成器实例
cg := generator.NewCodeGenerator("./generated")
```

### 2. 定义列信息

```go
columns := []generator.ColumnInfo{
    {
        ColumnName: "id",           // 数据库列名
        FieldName:  "ID",           // Go 字段名（驼峰）
        FieldType:  "int64",        // Go 字段类型
        JSONName:   "id",           // JSON 标签名
        IsPrimary:  true,           // 是否主键
        IsNullable: false,          // 是否可为空
    },
    {
        ColumnName: "username",
        FieldName:  "Username",
        FieldType:  "string",
        JSONName:   "username",
        IsPrimary:  false,
        IsNullable: false,
    },
    {
        ColumnName: "email",
        FieldName:  "Email",
        FieldType:  "string",
        JSONName:   "email",
        IsPrimary:  false,
        IsNullable: true,
    },
    {
        ColumnName: "created_at",
        FieldName:  "CreatedAt",
        FieldType:  "time.Time",
        JSONName:   "created_at",
    },
}
```

### 3. 生成代码

#### 方式一：一键生成（推荐）

```go
// 同时生成 Model 和 DAO
err := cg.GenerateAll("user", columns)
if err != nil {
    panic(err)
}
```

#### 方式二：分别生成

```go
// 只生成 Model
err := cg.GenerateModel("user", columns)

// 只生成 DAO
err := cg.GenerateDAO("user", "User")
```

## 📁 生成的文件结构

```
generated/
├── user.go              # User Model
├── user_dao.go          # User DAO
├── product.go           # Product Model
└── product_dao.go       # Product DAO
```

## 💡 使用生成的代码

### 导入包

```go
import (
    "context"
    "git.magicany.cc/black1552/gin-base/db/core"
    "git.magicany.cc/black1552/gin-base/db/model"
    "your-project/generated"
)
```

### 初始化数据库

```go
db, err := core.AutoConnect(false)
```

### 创建 DAO 实例

```go
userDAO := generated.NewUserDAO(db)
```

### CRUD 操作

```go
// 创建用户
user := &model.User{
    Username: "john",
    Email:    "john@example.com",
}
err = userDAO.Create(context.Background(), user)

// 查询用户
user, err := userDAO.GetByID(context.Background(), 1)

// 更新用户
user.Email = "new@example.com"
err = userDAO.Update(context.Background(), user)

// 删除用户
err = userDAO.Delete(context.Background(), 1)

// 分页查询
users, err := userDAO.FindByPage(context.Background(), 1, 10)
```

## 🔧 API 参考

### NewCodeGenerator

```go
func NewCodeGenerator(outputDir string) *CodeGenerator
```

创建代码生成器实例。

**参数:**
- `outputDir` - 输出目录路径

### GenerateModel

```go
func (cg *CodeGenerator) GenerateModel(tableName string, columns []ColumnInfo) error
```

生成 Model 代码。

**参数:**
- `tableName` - 数据库表名
- `columns` - 列信息数组

### GenerateDAO

```go
func (cg *CodeGenerator) GenerateDAO(tableName string, modelName string) error
```

生成 DAO 代码。

**参数:**
- `tableName` - 数据库表名
- `modelName` - Model 名称（驼峰）

### GenerateAll

```go
func (cg *CodeGenerator) GenerateAll(tableName string, columns []ColumnInfo) error
```

一键生成 Model + DAO（推荐）。

**参数:**
- `tableName` - 数据库表名
- `columns` - 列信息数组

## 📋 ColumnInfo 结构

```go
type ColumnInfo struct {
    ColumnName string  // 数据库列名（下划线风格）
    FieldName  string  // Go 字段名（驼峰风格）
    FieldType  string  // Go 数据类型
    JSONName   string  // JSON 标签名
    IsPrimary  bool    // 是否主键
    IsNullable bool    // 是否可为空
}
```

## 🗺️ 类型映射表

| 数据库类型 | Go 类型 |
|-----------|---------|
| INT/BIGINT | int64 |
| VARCHAR/TEXT | string |
| DATETIME | time.Time |
| TIMESTAMP | time.Time |
| BOOLEAN | bool |
| FLOAT/DOUBLE | float64 |
| DECIMAL | string 或 float64 |

## 🎯 最佳实践

### 1. 从数据库读取真实表结构

```go
// 示例：从 MySQL INFORMATION_SCHEMA 获取列信息
query := `
    SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, COLUMN_KEY
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = 'your_database'
    AND TABLE_NAME = 'your_table'
`
```

### 2. 批量生成所有表

```go
tables := []string{"users", "products", "orders"}

for _, table := range tables {
    columns := getColumnsFromDB(table) // 从数据库获取列信息
    err := cg.GenerateAll(table, columns)
    if err != nil {
        log.Printf("生成 %s 失败：%v", table, err)
    }
}
```

### 3. 自定义模板

修改 `generator.go` 中的模板字符串，添加自定义方法：

```go
tmpl := `package model

// {{.ModelName}} {{.TableName}} 模型
type {{.ModelName}} struct {
{{range .Columns}}
    {{.FieldName}} {{.FieldType}} ` + "`" + `json:"{{.JSONName}}" db:"{{.ColumnName}}"` + "`" + `
{{end}}
}

// 自定义方法
func (m *{{.ModelName}}) Validate() error {
    // 验证逻辑
    return nil
}
`
```

### 4. 代码审查

- ✅ 检查生成的字段类型是否正确
- ✅ 验证主键和索引设置
- ✅ 添加业务逻辑方法
- ✅ 补充注释和文档

### 5. 版本控制

```bash
# 将生成的代码纳入 Git 管理
git add generated/
git commit -m "feat: 生成用户和产品模块代码"
```

## ⚠️ 注意事项

1. **不要频繁覆盖**: 手动修改的代码可能会被覆盖
2. **代码审查**: 生成的代码需要人工审查
3. **类型映射**: 特殊类型可能需要手动调整
4. **关联关系**: 复杂的模型关联需要手动实现
5. **验证逻辑**: 业务验证逻辑需要手动添加

## 🎉 总结

✅ **优势:**
- 大幅提高开发效率
- 代码规范统一
- 减少重复劳动
- 快速搭建项目骨架

✅ **适用场景:**
- 新项目快速启动
- 数据库表结构变更
- 批量生成基础代码
- 原型开发

✅ **推荐用法:**
- 使用 `GenerateAll` 一键生成
- 从真实数据库读取列信息
- 定期重新生成保持同步
- 配合版本控制管理代码

开始使用代码生成器，提升你的开发效率吧！🚀

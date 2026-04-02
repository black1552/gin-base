# Magic-ORM 代码生成器 - 命令行工具

## 🚀 快速开始

### 1. 构建命令行工具

**Windows:**
```bash
build.bat
```

**Linux/Mac:**
```bash
chmod +x build.sh
./build.sh
```

或者手动构建：
```bash
cd db
go build -o ../bin/gendb ./cmd/gendb
```

### 2. 使用方法

#### 基础用法

```bash
# 生成单个表
gendb user

# 生成多个表
gendb user product order

# 指定输出目录
gendb -o ./models user product
```

#### 高级用法

```bash
# 自定义列定义
gendb user id:int64:primary username:string email:string created_at:time.Time

# 混合使用（自动推断 + 自定义）
gendb -o ./generated user username:string email:string product name:string price:float64

# 查看版本
gendb -v

# 查看帮助
gendb -h
```

## 📋 功能特性

✅ **自动生成**: 根据表名自动推断常用字段
✅ **批量生成**: 一次生成多个表的代码
✅ **自定义列**: 支持手动指定列定义
✅ **灵活输出**: 可指定输出目录
✅ **智能推断**: 自动识别常见表结构

## 🎯 支持的类型

| 类型别名 | Go 类型 |
|---------|---------|
| int, integer, bigint | int64 |
| string, text, varchar | string |
| time, datetime | time.Time |
| bool, boolean | bool |
| float, double | float64 |
| decimal | string |

## 📝 列定义格式

```
字段名：类型 [:primary] [:nullable]
```

示例：
- `id:int64:primary` - 主键 ID
- `username:string` - 用户名字段
- `email:string:nullable` - 可为空的邮箱字段
- `created_at:time.Time` - 创建时间字段

## 🔧 预设表结构

工具内置了常见表的默认结构：

### user / users
- id (主键)
- username
- email (可空)
- password
- status
- created_at
- updated_at

### product / products
- id (主键)
- name
- price
- stock
- description (可空)
- created_at

### order / orders
- id (主键)
- order_no
- user_id
- total_amount
- status
- created_at

## 💡 使用示例

### 示例 1: 快速生成用户模块

```bash
gendb user
```

生成文件：
- `generated/user.go` - User Model
- `generated/user_dao.go` - User DAO

### 示例 2: 生成电商模块

```bash
gendb -o ./shop user product order
```

生成文件：
- `shop/user.go`
- `shop/user_dao.go`
- `shop/product.go`
- `shop/product_dao.go`
- `shop/order.go`
- `shop/order_dao.go`

### 示例 3: 完全自定义

```bash
gendb article \
  id:int64:primary \
  title:string \
  content:string:nullable \
  author_id:int64 \
  view_count:int \
  published:bool \
  created_at:time.Time
```

## 📁 生成的代码结构

### Model (user.go)

```go
package model

import "time"

// User user 表模型
type User struct {
    ID        int64     `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TableName 表名
func (User) TableName() string {
    return "user"
}
```

### DAO (user_dao.go)

```go
package dao

import (
    "context"
    "git.magicany.cc/black1552/gin-base/db/core"
    "git.magicany.cc/black1552/gin-base/db/model"
)

// UserDAO user 表数据访问对象
type UserDAO struct {
    db *core.Database
}

// NewUserDAO 创建 UserDAO 实例
func NewUserDAO(db *core.Database) *UserDAO {
    return &UserDAO{db: db}
}

// Create 创建记录
func (dao *UserDAO) Create(ctx context.Context, model *model.User) error {
    _, err := dao.db.Model(model).Insert(model)
    return err
}

// GetByID 根据 ID 查询
func (dao *UserDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
    var result model.User
    err := dao.db.Model(&model.User{}).Where("id = ?", id).First(&result)
    if err != nil {
        return nil, err
    }
    return &result, nil
}

// ... 更多 CRUD 方法
```

## 🛠️ 安装到 PATH

### Windows

1. 将 `bin` 目录添加到系统环境变量 PATH
2. 或者复制 `gendb.exe` 到任意 PATH 中的目录

```powershell
# 临时添加到当前会话
$env:PATH += ";$(pwd)\bin"

# 永久添加（需要管理员权限）
[Environment]::SetEnvironmentVariable(
    "Path", 
    $env:Path + ";$(pwd)\bin", 
    [EnvironmentVariableTarget]::Machine
)
```

### Linux/Mac

```bash
# 临时添加到当前会话
export PATH=$PATH:$(pwd)/bin

# 永久添加（添加到 ~/.bashrc 或 ~/.zshrc）
echo 'export PATH=$PATH:$(pwd)/bin' >> ~/.bashrc
source ~/.bashrc

# 或者复制到系统目录
sudo cp bin/gendb /usr/local/bin/
```

## ⚙️ 选项说明

| 选项 | 简写 | 说明 | 默认值 |
|------|------|------|--------|
| `-version` | `-v` | 显示版本号 | - |
| `-help` | `-h` | 显示帮助信息 | - |
| `-o` | - | 输出目录 | `./generated` |

## 🎨 最佳实践

### 1. 从数据库读取真实结构

```bash
# 先用 SQL 导出表结构
mysql -u root -p -e "DESCRIBE your_database.users;"

# 然后根据输出调整列定义
```

### 2. 批量生成项目所有表

```bash
# 一次性生成所有表
gendb user product order category tag article comment
```

### 3. 版本控制

```bash
# 将生成的代码纳入 Git 管理
git add generated/
git commit -m "feat: 生成基础 Model 和 DAO 代码"
```

### 4. 自定义扩展

生成的代码可以作为基础，手动添加：
- 业务逻辑方法
- 验证逻辑
- 关联查询
- 索引优化

## ⚠️ 注意事项

1. **生成的代码需审查**: 自动生成的代码可能不完全符合业务需求
2. **不要频繁覆盖**: 手动修改的代码可能会被覆盖
3. **类型映射**: 特殊类型可能需要手动调整
4. **关联关系**: 复杂的模型关联需手动实现

## 🐛 故障排除

### 问题 1: 找不到命令

```bash
# 确保已构建并添加到 PATH
gendb: command not found

# 解决：
./bin/gendb -h  # 使用相对路径
```

### 问题 2: 生成失败

```bash
# 检查输出目录是否有写权限
# 检查表名是否合法
# 使用 -h 查看正确的语法
```

### 问题 3: 类型不匹配

```bash
# 手动指定正确的类型
gendb user price:float64 instead of price:int
```

## 📞 获取帮助

```bash
# 查看完整帮助
gendb -h

# 查看版本
gendb -v
```

## 🎉 开始使用

```bash
# 最简单的用法
gendb user

# 立即体验！
```

---

**Magic-ORM Code Generator** - 让代码生成如此简单！🚀

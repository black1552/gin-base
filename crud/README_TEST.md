# CRUD 测试用例说明

## 测试文件

### 1. `curd_func_test.go` - 不依赖数据库的测试
测试所有不依赖数据库的函数，包括：
- `TestBuildWhere` - 测试 BuildWhere 方法及其辅助函数
- `TestBuildMap` - 测试 BuildMap 方法
- `TestClearField` - 测试 ClearField 方法
- `TestBuildWhereAndOr` - 测试 BuildWhereAndOr 构建器
- `TestPaginateStruct` - 测试 Paginate 结构体
- `TestPageInfo` - 测试 pageInfo 常量

**运行方式：**
```bash
go test -v ./crud -run "TestBuild|TestClear|TestPaginate|TestPage"
```

### 2. `curd_test.go` - 依赖数据库的测试
测试所有需要数据库的 CRUD 操作方法，包括：
- `TestCrud_BuildWhere` - 完整的 BuildWhere 测试
- `TestCrud_ClearFieldPage` - 分页查询测试
- `TestCrud_ClearFieldList` - 列表查询测试
- `TestCrud_ClearFieldOne` - 单条查询测试
- `TestCrud_Value` - 字段值查询测试
- `TestCrud_DeletePri` - 按主键删除测试
- `TestCrud_DeleteWhere` - 按条件删除测试
- `TestCrud_Sum` - 求和测试
- `TestCrud_ArrayField` - 字段数组查询测试
- `TestCrud_FindPri` - 按主键查询测试
- `TestCrud_First` - 查询第一条测试
- `TestCrud_Exists` - 存在性检查测试
- `TestCrud_All` - 查询所有测试
- `TestCrud_Count` - 统计测试
- `TestCrud_Save` - 保存测试
- `TestCrud_Update` - 更新测试
- `TestCrud_UpdatePri` - 按主键更新测试
- `TestCrud_Paginate` - 分页查询测试
- `TestHelperFunctions` - 辅助函数测试

**运行方式：**
```bash
# 需要 MySQL 服务
go test -v ./crud -run TestCrud
```

**注意：** 数据库测试需要 MySQL 服务，连接信息：
- Host: 127.0.0.1:3306
- User: root
- Password: root
- Database: test

如果没有 MySQL 服务，这些测试会自动跳过。

## 运行所有测试

```bash
# 运行所有测试
go test -v ./crud

# 运行特定测试
go test -v ./crud -run TestBuildWhere

# 显示覆盖率
go test -v ./crud -cover
```

## 测试覆盖的函数列表

### 核心方法
- ✅ `BuildWhere` - 构建查询条件 map
- ✅ `BuildMap` - 构建变更条件 map
- ✅ `BuildWhereAndOr` - AND/OR 查询条件构建器
- ✅ `BuildWhereGORM` - GORM 原生语法构建器
- ✅ `ClearField` - 清理请求参数
- ✅ `ClearFieldPage` - 清理参数 + 分页查询
- ✅ `ClearFieldList` - 清理参数 + 列表查询
- ✅ `ClearFieldOne` - 清理参数 + 单条查询

### 查询方法
- ✅ `Value` - 查询单个字段值
- ✅ `FindPri` - 按主键查询单条记录
- ✅ `First` - 按条件查询第一条记录
- ✅ `Exists` - 判断记录是否存在
- ✅ `All` - 查询所有符合条件的记录
- ✅ `Count` - 统计记录总数
- ✅ `ArrayField` - 查询指定字段数组
- ✅ `Sum` - 字段求和

### 操作方法
- ✅ `Save` - 新增/更新记录
- ✅ `Update` - 按条件更新记录
- ✅ `UpdatePri` - 按主键更新记录
- ✅ `DeletePri` - 按主键删除
- ✅ `DeleteWhere` - 按条件删除
- ✅ `Paginate` - 分页查询

### 辅助函数
- ✅ `convToMap` - 类型转换为 map
- ✅ `isEmpty` - 判断值是否为空
- ✅ `strInArray` - 判断字符串是否在数组中
- ✅ `caseConvert` - 字段名风格转换

## 测试示例

### BuildWhereAndOr 使用示例
```go
func TestBuildWhereAndOr(t *testing.T) {
    var crud Crud[interface{}]
    
    // 混合使用 AND 和 OR
    where := crud.BuildWhereAndOr().
        AND(map[string]any{"status": 1}).
        OR(
            map[string]any{"age >": 25},
            map[string]any{"vip": true},
        ).
        Build()
    
    if where == nil {
        t.Error("Expected where to not be nil")
    }
}
```

### BuildWhereGORM 使用示例
```go
func TestBuildWhereGORM(t *testing.T) {
    var crud Crud[TestModel]
    
    // 使用 GORM 原生语法
    var results []*TestModel
    err := crud.BuildWhereGORM("status = ?", 1).
        Where("age > ?", 20).
        Or("vip = ?", true).
        Find(&results)
    
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}
```

## 注意事项

1. 数据库测试需要 MySQL 服务，如果不可用会自动跳过
2. 所有测试都是独立的，不会相互影响
3. 每个测试都会创建自己的测试数据
4. 测试完成后会自动清理资源

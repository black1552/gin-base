# GF-Source 自维护代码

这个目录包含了从 GoFrame (GF) 框架复制并修改的 DAO 生成工具源码，已将所有 `gdb` 相关引用替换为项目自定义的 `database` 包。

## 📁 目录结构

```
gf-source/
├── internal/                    # 内部工具包
│   ├── consts/                 # 常量定义（从 GF internal/consts 复制）
│   └── utility/                # 工具函数
│       ├── mlog/              # 日志工具
│       └── utils/             # 通用工具
├── templates/                  # 代码生成模板
│   ├── consts_gen_dao_template_dao.go
│   ├── consts_gen_dao_template_do.go
│   ├── consts_gen_dao_template_entity.go
│   └── consts_gen_dao_template_table.go
├── gendao*.go                  # DAO 生成核心逻辑
├── go.mod                      # 子模块依赖
└── go.sum
```

## 🔧 主要修改

### 1. 数据库接口替换
- ✅ `github.com/gogf/gf/v2/database/gdb` → `github.com/black1552/gin-base/database`
- ✅ `gdb.DB` → `database.DB`
- ✅ `gdb.ConfigNode` → `database.ConfigNode`
- ✅ `gdb.AddConfigNode()` → `database.AddConfigNode()`
- ✅ `gdb.Instance()` → `database.Instance()`
- ✅ `g.DB()` → `database.Database()`

### 2. 内部包路径修改
- ✅ `github.com/gogf/gf/cmd/gf/v2/internal/consts` → `github.com/black1552/gin-base/cmd/gf-source/internal/consts`
- ✅ `github.com/gogf/gf/cmd/gf/v2/internal/utility/mlog` → `github.com/black1552/gin-base/cmd/gf-source/internal/utility/mlog`
- ✅ `github.com/gogf/gf/cmd/gf/v2/internal/utility/utils` → `github.com/black1552/gin-base/cmd/gf-source/internal/utility/utils`

### 3. 包名统一
- ✅ 所有 gendao 相关文件统一为 `package gendao`
- ✅ 模板文件移至 `templates/` 子目录（避免包名冲突）

## 📦 编译

```bash
cd cmd/gf-source
go build .
```

## ⚠️ 注意事项

1. **这是参考代码**：这个目录主要用于参考 GF 的 DAO 生成逻辑，实际使用的是 `cmd/gin-dao-gen/main.go`
2. **不要直接导入**：这个子模块有独立的 go.mod，不应该被主项目直接导入
3. **保持同步**：如果 GF 更新了 DAO 生成逻辑，需要手动同步这些文件并重新应用修改

## 🔄 更新流程

如果需要从 GF 更新代码：

1. 从 `D:\web-object\gf\cmd\gf\internal\cmd\gendao` 复制最新文件
2. 运行批量替换脚本将 `gdb` 改为 `database`
3. 运行批量替换脚本将 internal 路径改为本项目路径
4. 测试编译确保没有错误

## 📝 相关文件

- 主程序：`cmd/gin-dao-gen/main.go`
- 数据库包：`database/`
- 配置文件：`config/config.toml`

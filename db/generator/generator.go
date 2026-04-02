package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// CodeGenerator 代码生成器 - 自动生成 Model 和 DAO 代码
type CodeGenerator struct {
	outputDir string // 输出目录
}

// NewCodeGenerator 创建代码生成器实例
func NewCodeGenerator(outputDir string) *CodeGenerator {
	return &CodeGenerator{
		outputDir: outputDir,
	}
}

// GenerateModel 生成 Model 代码
func (cg *CodeGenerator) GenerateModel(tableName string, columns []ColumnInfo) error {
	// 生成文件名
	fileName := strings.ToLower(tableName) + ".go"
	filePath := filepath.Join(cg.outputDir, fileName)

	// 创建输出目录
	if err := os.MkdirAll(cg.outputDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	// 生成代码
	code := cg.generateModelCode(tableName, columns)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Printf("[Model] Generated: %s\n", filePath)
	return nil
}

// GenerateDAO 生成 DAO 代码
func (cg *CodeGenerator) GenerateDAO(tableName string, modelName string) error {
	fileName := strings.ToLower(tableName) + "_dao.go"
	// DAO 文件输出到 dao 子目录
	daoDir := filepath.Join(cg.outputDir, "dao")
	filePath := filepath.Join(daoDir, fileName)

	// 创建输出目录（包括 dao 子目录）
	if err := os.MkdirAll(daoDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败：%w", err)
	}

	// 生成代码
	code := cg.generateDAOCode(tableName, modelName)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(code), 0644); err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Printf("[DAO] Generated: %s\n", filePath)
	return nil
}

// GenerateAll 生成完整代码（Model + DAO）
func (cg *CodeGenerator) GenerateAll(tableName string, columns []ColumnInfo) error {
	modelName := cg.toCamelCase(tableName)

	// 生成 Model
	if err := cg.GenerateModel(tableName, columns); err != nil {
		return err
	}

	// 生成 DAO
	if err := cg.GenerateDAO(tableName, modelName); err != nil {
		return err
	}

	return nil
}

// generateModelCode 生成 Model 代码
func (cg *CodeGenerator) generateModelCode(tableName string, columns []ColumnInfo) string {
	// 检查是否有时间字段
	hasTime := false
	for _, col := range columns {
		if col.FieldType == "time.Time" {
			hasTime = true
			break
		}
	}

	importStmt := ""
	if hasTime {
		importStmt = `import "time"`
	} else {
		importStmt = ``
	}

	tmpl := `package model

` + importStmt + `
// {{.ModelName}} {{.TableName}} 模型
type {{.ModelName}} struct {
{{range .Columns}}	{{.FieldName}} {{.FieldType}} ` + "`" + `json:"{{.JSONName}}" db:"{{.ColumnName}}"` + "`" + `
{{end}}
}

// TableName 表名
func ({{.ModelName}}) TableName() string {
	return "{{.TableName}}"
}
`

	data := struct {
		ModelName string
		TableName string
		Columns   []ColumnInfo
	}{
		ModelName: cg.toCamelCase(tableName),
		TableName: tableName,
		Columns:   columns,
	}

	return cg.executeTemplate(tmpl, data)
}

// generateDAOCode 生成 DAO 代码 - 简化版本，只定义结构体并继承 Database
func (cg *CodeGenerator) generateDAOCode(tableName string, modelName string) string {
	tmpl := `package dao

import (
	"git.magicany.cc/black1552/gin-base/db/core"
	"git.magicany.cc/black1552/gin-base/db/model"
)

// {{.ModelName}}DAO {{.TableName}} 数据访问对象
// 嵌入 core.DAO，自动获得所有 CRUD 方法
type {{.ModelName}}DAO struct {
	*core.DAO
}

// New{{.ModelName}}DAO 创建 {{.ModelName}}DAO 实例
func New{{.ModelName}}DAO(db *core.Database) *{{.ModelName}}DAO {
	return &{{.ModelName}}DAO{
		DAO: core.NewDAOWithModel(db, &model.{{.ModelName}}{}),
	}
}
`

	data := struct {
		ModelName string
		TableName string
	}{
		ModelName: cg.toCamelCase(tableName),
		TableName: tableName,
	}

	return cg.executeTemplate(tmpl, data)
}

// executeTemplate 执行模板
func (cg *CodeGenerator) executeTemplate(tmpl string, data interface{}) string {
	t := template.Must(template.New("code").Parse(tmpl))

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("// 模板执行错误：%v", err)
	}

	return buf.String()
}

// toCamelCase 转换为驼峰命名
func (cg *CodeGenerator) toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	result := ""

	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(string(part[0])) + part[1:]
		}
	}

	return result
}

// ColumnInfo 列信息
type ColumnInfo struct {
	ColumnName string // 列名
	FieldName  string // 字段名（驼峰）
	FieldType  string // 字段类型
	JSONName   string // JSON 标签名
	IsPrimary  bool   // 是否主键
	IsNullable bool   // 是否可为空
}

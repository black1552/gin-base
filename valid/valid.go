package valid

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

// -------------------------- 全局变量（翻译器+验证器） --------------------------
var (
	trans    ut.Translator       // 全局中文翻译器
	validate *validator.Validate // 全局验证器实例
	found    bool
)

// -------------------------- 初始化全局验证器+中文翻译器（关键：与Gin关联） --------------------------
func init() {
	// 1. 初始化中文本地化
	zhLocale := zh.New()
	uni := ut.New(zhLocale, zhLocale)

	// 2. 获取中文翻译器
	var err error
	trans, found = uni.GetTranslator("zh")
	if !found {
		panic("获取中文翻译器失败")
	}

	// 3. 关键：获取 Gin 内置的 validator 实例（而非手动创建 new(validator.Validate)）
	// 保证 Gin 绑定参数时使用的是我们注册了翻译的这个验证器
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validate = v
	} else {
		panic("获取 Gin 内置验证器失败")
	}

	// 4. 注册中文翻译（核心：将验证器与中文翻译器绑定）
	err = zh_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic("注册中文翻译失败：" + err.Error())
	}

	// 5. 注册字段名映射：用 json 标签替换结构体字段名（解决报错中的 Title -> title 问题）
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 提取 json 标签，忽略 ,omitempty 等附加选项
		jsonTag := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if jsonTag == "-" {
			return ""
		}
		return jsonTag
	})
}

// ValidateError -------------------------- 工具函数：解析第一条验证错误（中文翻译） --------------------------
func ValidateError(err error) string {
	// 1. 断言错误类型为 validator.ValidationErrors
	var validateErrors validator.ValidationErrors
	ok := errors.As(err, &validateErrors)
	if !ok {
		// 非参数验证错误，直接返回
		return "参数格式错误：" + err.Error()
	}

	// 2. 只返回第一条中文翻译错误
	return validateErrors[0].Translate(trans)
}

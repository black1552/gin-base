package valid

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni      *ut.UniversalTranslator // 全局翻译器实例
	validate *validator.Validate     // 全局验证器实例
	trans    ut.Translator           // 全局中文翻译器
	foun     bool
)

func init() {
	// 1. 初始化中文本地化实例
	zhLocale := zh.New()

	// 2. 初始化通用翻译器，加载中文本地化
	uni = ut.New(zhLocale, zhLocale)

	// 3. 获取中文翻译器实例
	var err error
	trans, foun = uni.GetTranslator("zh")
	if !foun {
		panic("获取中文翻译器失败")
	}

	// 4. 获取 Gin 内置的 validator 验证器实例
	validate = validator.New()

	// 5. 注册 validator 中文翻译（替换默认英文错误提示）
	err = zh_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		panic("注册中文翻译失败：" + err.Error())
	}

	// 6. 核心优化：注册字段名映射函数，用结构体的 json 标签作为返回的字段名
	// 例如：结构体字段 Username -> json 标签 username -> 错误提示中显示 username（更友好）
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 提取 json 标签（忽略 ,omitempty 等附加选项）
		jsonTag := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// 若 json 标签为 "-"（表示忽略该字段），返回空字符串
		if jsonTag == "-" {
			return ""
		}
		return jsonTag
	})
}
func GetFirstValidateError(err error) string {
	// 1. 断言错误类型是否为 validator 验证错误
	var validateErrors validator.ValidationErrors
	ok := errors.As(err, &validateErrors)
	if !ok {
		// 非参数验证错误（如绑定格式错误），直接返回错误信息
		return "参数格式错误：" + err.Error()
	}

	// 2. 仅返回第一条验证错误（核心：取切片索引 0 的元素）
	firstErr := validateErrors[0]

	// 3. 翻译第一条错误为中文，返回给前端
	return firstErr.Translate(trans)
}

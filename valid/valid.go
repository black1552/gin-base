package valid

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

// ValidToStruct 验证参数并返回结构体
func ValidToStruct[T any](c *gin.Context) (object *T) {
	obj := new(T)
	if err := c.BindJSON(obj); err != nil {
		panic(err)
	}
	if err := g.Validator().Data(obj).Run(c); err != nil {
		panic(gerror.Current(err).Error())
	}
	return obj
}

// ValidToStructAndMap 验证参数并返回map
func ValidToStructAndMap[T any](c *gin.Context) (stru *T, object map[string]any) {
	obj := new(T)
	if err := c.BindJSON(obj); err != nil {
		panic(err)
	}
	if err := g.Validator().Data(obj).Run(c); err != nil {
		panic(gerror.Current(err).Error())
	}
	return obj, gconv.Map(obj)
}

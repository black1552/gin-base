package valid

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// ValidAndStruct 验证参数并返回结构体
func ValidAndStruct[T any](c *gin.Context) (object *T) {
	object = new(T)
	if err := c.BindJSON(object); err != nil {
		panic(err)
	}
	if err := g.Validator().Data(object).Run(c); err != nil {
		panic(gerror.Current(err).Error())
	}
	return
}

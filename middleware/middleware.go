package middleware

import (
	"net/http"

	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/response"
	"github.com/gin-gonic/gin"
)

// ErrorHandler 全局异常处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				res := response.Error(c).SetCode(http.StatusInternalServerError)
				switch e := err.(type) {
				case string:
					res = res.SetMsg(e)
				case error:
					res = res.SetMsg(e.Error())
				default:
					res = res.SetMsg("服务器内部异常，请稍后重试")
				}
				res.End()
				log.Error("发生panic=》", "path：", c.Request.URL.Path, ",method:", c.Request.Method, ",", err)
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

package server

import (
	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/middleware"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	g := gin.New()
	g.Use(middleware.ErrorHandler())
	g.Use(middleware.CORSMiddleware())
	g.Run(config.GetConfigValue("server.addr", "127.0.0.1:8080").String())
	return g
}

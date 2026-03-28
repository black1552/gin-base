package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.magicany.cc/black1552/gin-base/config"
	"git.magicany.cc/black1552/gin-base/log"
	"git.magicany.cc/black1552/gin-base/middleware"
	"github.com/gin-gonic/gin"
)

// New 创建一个gin实例
// 默认使用全局异常处理中间件
// 默认使用跨域中间件
// @Return *gin.Engine 可以使用创建路由
func New() *gin.Engine {
	gin.SetMode(config.GetConfigValue("server.mode", "debug").String())
	g := gin.New()
	g.Use(middleware.ErrorHandler())
	g.Use(middleware.CORSMiddleware())
	return g
}

// Run 启动服务
// @Param *gin.Engine 路由实例
// 设置监听挂壁
func Run(g *gin.Engine) {
	addr := config.GetConfigValue("server.addr", ":8080").String()
	s := &http.Server{
		Addr:              addr,
		Handler:           g,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    20 * 1024 * 1024,
	}
	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// 检查是否是端口被占用的错误
			if strings.Contains(err.Error(), "address already in use") ||
				strings.Contains(err.Error(), "Only one usage of each socket address") {
				log.Error(fmt.Sprintf("服务器启动失败：%s 端口已被占用，请检查是否有其他程序正在使用该端口", addr))
			} else {
				log.Error("服务器启动失败:", err)
			}
			os.Exit(1) // 启动失败则退出程序
		}

		log.Info("服务器启动成功....")
		if strings.Contains(s.Addr, "127.0.0.1") || strings.Contains(s.Addr, "0.0.0.0") || strings.Contains(s.Addr, "locahost") {
			log.Info("请使用打开:", fmt.Sprintf("http://%s\n", s.Addr))
		} else {
			log.Info("请使用打开:", fmt.Sprintf("http://localhost%s\n", s.Addr))
		}
	}()
	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在关闭服务器...")

	// 设置5秒的超时时间用于关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		log.Error("服务器强制关闭")
	}
	log.Info("服务器已退出")
}

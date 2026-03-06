package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"git.magicany.cc/black1552/gin-base/log"
	"github.com/gin-gonic/gin"
)

// BuildRequest 创建 HTTP 反向代理 Handler
// 支持 HTTP 和 WebSocket 协议
// @Param host 目标服务器地址，例如 "http://127.0.0.1:8081"
func BuildRequest(host string) gin.HandlerFunc {
	return func(c *gin.Context) {
		targetURL, err := url.Parse(host)
		if err != nil {
			log.Error("parse target host error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid target host"})
			return
		}

		// 创建反向代理
		proxy := httputil.NewSingleHostReverseProxy(targetURL)

		// 自定义 Director 修改请求，保留原始路径和查询参数
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// 保留原始请求的路径和查询参数
			req.URL.Path = c.Request.URL.Path
			req.URL.RawQuery = c.Request.URL.RawQuery
			req.Host = targetURL.Host
		}

		// 处理代理错误
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("proxy error:", err)
			http.Error(w, "proxy error: "+err.Error(), http.StatusBadGateway)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

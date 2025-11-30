package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lf-web-tools/gin-web-server/middleware"
	"github.com/lf-web-tools/gin-web-server/routes"
)

func main() {
	// 创建一个默认的gin路由引擎
	r := gin.Default()

	// 设置静态文件目录
	r.Static("/static", "./static")

	// PWA 相关路由
	r.GET("/manifest.json", func(c *gin.Context) {
		c.File("./static/manifest.json")
	})

	r.GET("/sw.js", func(c *gin.Context) {
		// Service Worker 需要特殊的缓存头
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.File("./static/sw.js")
	})

	// 设置HTML模板目录
	r.LoadHTMLGlob("templates/*")

	// 定义根路由
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/static/index.html")
	})

	// 设置API路由
	routes.SetupAPIRoutes(r)

	// 设置WebSocket路由
	routes.SetupWebSocketRoutes(r)

	// 设置页面路由
	routes.SetupPageRoutes(r)

	// 设置CORS代理路由
	middleware.RegisterCorsProxyRoutes(r)

	// 设置端口扫描路由
	middleware.RegisterPortScanRoutes(r)

	// 启动服务器
	r.Run(":8081") // 监听并在0.0.0.0:8081上启动服务
}

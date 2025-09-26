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

	// 启动服务器
	r.Run(":8080") // 监听并在0.0.0.0:8080上启动服务
}

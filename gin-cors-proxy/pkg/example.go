package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/lf-web-tools/gin-cors-proxy/middleware"
)

// SetupExampleServer 展示如何将CORS代理集成到现有Gin应用中的示例
func SetupExampleServer() *gin.Engine {
	r := gin.Default()
	
	// 注册CORS代理路由
	middleware.RegisterCorsProxyRoutes(r)
	
	// 添加其他路由
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to Gin CORS Proxy",
			"usage": "POST to /cors-proxy with {\"curlParam\": \"curl -X GET https://example.com\"}",
		})
	})
	
	return r
}

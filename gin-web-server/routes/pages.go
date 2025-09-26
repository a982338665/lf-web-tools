package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupPageRoutes 设置页面相关的路由
func SetupPageRoutes(r *gin.Engine) {
	// 示例模板页面
	r.GET("/hello", func(c *gin.Context) {
		c.HTML(http.StatusOK, "hello.html", gin.H{
			"message": "欢迎使用Gin Web服务器",
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		})
	})
}

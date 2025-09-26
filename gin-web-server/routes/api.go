package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupAPIRoutes 设置API相关的路由
func SetupAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 获取服务器时间
		api.GET("/time", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"time": time.Now().Format(time.RFC3339),
			})
		})

		// 获取服务器信息
		api.GET("/info", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"name":    "LF Web Tools",
				"version": "1.0.0",
				"status":  "running",
			})
		})
	}
}

package routes

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// QRCodeRequest 二维码生成请求
type QRCodeRequest struct {
	Text       string `json:"text" binding:"required"`
	Size       int    `json:"size"`
	ErrorLevel string `json:"errorLevel"`
}

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

		// 生成二维码API
		api.POST("/generate-qrcode", func(c *gin.Context) {
			var req QRCodeRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "无效的请求参数: " + err.Error(),
				})
				return
			}

			// 设置默认值
			if req.Size == 0 {
				req.Size = 300
			}
			if req.ErrorLevel == "" {
				req.ErrorLevel = "M"
			}

			// 验证参数
			if req.Size < 100 || req.Size > 1000 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "尺寸必须在100-1000之间",
				})
				return
			}

			// 转换容错级别
			var recoveryLevel qrcode.RecoveryLevel
			switch req.ErrorLevel {
			case "L":
				recoveryLevel = qrcode.Low
			case "M":
				recoveryLevel = qrcode.Medium
			case "Q":
				recoveryLevel = qrcode.High
			case "H":
				recoveryLevel = qrcode.Highest
			default:
				recoveryLevel = qrcode.Medium
			}

			// 生成二维码
			qrCode, err := qrcode.New(req.Text, recoveryLevel)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "生成二维码失败: " + err.Error(),
				})
				return
			}

			// 生成PNG图片数据
			pngBytes, err := qrCode.PNG(req.Size)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "生成PNG数据失败: " + err.Error(),
				})
			}

			// 转换为base64
			base64String := base64.StdEncoding.EncodeToString(pngBytes)
			dataURL := "data:image/png;base64," + base64String

			// 返回结果
			c.JSON(http.StatusOK, gin.H{
				"success":    true,
				"dataUrl":    dataURL,
				"size":       req.Size,
				"errorLevel": req.ErrorLevel,
				"content":    req.Text,
				"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
			})
		})

		// 生成二维码图片（直接返回PNG）
		api.GET("/qrcode", func(c *gin.Context) {
			text := c.Query("text")
			if text == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "缺少text参数",
				})
				return
			}

			sizeStr := c.DefaultQuery("size", "300")
			size, err := strconv.Atoi(sizeStr)
			if err != nil || size < 100 || size > 1000 {
				size = 300
			}

			errorLevel := c.DefaultQuery("level", "M")
			var recoveryLevel qrcode.RecoveryLevel
			switch errorLevel {
			case "L":
				recoveryLevel = qrcode.Low
			case "M":
				recoveryLevel = qrcode.Medium
			case "Q":
				recoveryLevel = qrcode.High
			case "H":
				recoveryLevel = qrcode.Highest
			default:
				recoveryLevel = qrcode.Medium
			}

			// 生成二维码PNG数据
			pngBytes, err := qrcode.Encode(text, recoveryLevel, size)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "生成二维码失败: " + err.Error(),
				})
				return
			}

			// 设置响应头并返回PNG数据
			c.Header("Content-Type", "image/png")
			c.Header("Cache-Control", "public, max-age=3600")
			c.Data(http.StatusOK, "image/png", pngBytes)
		})
	}
}

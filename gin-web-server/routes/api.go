package routes

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
)

// QRCodeRequest 二维码生成请求
type QRCodeRequest struct {
	Text       string `json:"text" binding:"required"`
	Size       int    `json:"size"`
	ErrorLevel string `json:"errorLevel"`
	// 颜色配置
	ForegroundColor string `json:"foregroundColor"` // 码颜色
	BackgroundColor string `json:"backgroundColor"` // 背景颜色
	// Logo配置
	LogoData string  `json:"logoData"` // Logo的base64数据
	LogoSize float64 `json:"logoSize"` // Logo大小比例 (0.1-0.3)
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
			if req.ForegroundColor == "" {
				req.ForegroundColor = "#000000"
			}
			if req.BackgroundColor == "" {
				req.BackgroundColor = "#FFFFFF"
			}
			if req.LogoSize == 0 {
				req.LogoSize = 0.2
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

			// 设置二维码颜色
			if req.ForegroundColor != "" && req.BackgroundColor != "" {
				fgColor, err := parseHexColor(req.ForegroundColor)
				if err == nil {
					bgColor, err := parseHexColor(req.BackgroundColor)
					if err == nil {
						qrCode.ForegroundColor = fgColor
						qrCode.BackgroundColor = bgColor
					}
				}
			}

			// 生成PNG图片数据
			pngBytes, err := qrCode.PNG(req.Size)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "生成PNG数据失败: " + err.Error(),
				})
				return
			}

			// 如果有Logo，添加Logo到二维码中心
			if req.LogoData != "" {
				pngBytes, err = addLogoToQRCode(pngBytes, req.LogoData, req.Size, req.LogoSize)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "添加Logo失败: " + err.Error(),
					})
					return
				}
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

// parseHexColor 解析十六进制颜色字符串
func parseHexColor(hexColor string) (color.RGBA, error) {
	// 移除#号
	hex := strings.TrimPrefix(hexColor, "#")

	// 解析RGB值
	if len(hex) != 6 {
		return color.RGBA{}, nil
	}

	var r, g, b uint8
	_, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	r = uint8(parseUint(hex[0:2], 16))

	_, err = strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	g = uint8(parseUint(hex[2:4], 16))

	_, err = strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	b = uint8(parseUint(hex[4:6], 16))

	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func parseUint(s string, base int) uint64 {
	val, _ := strconv.ParseUint(s, base, 8)
	return val
}

// addLogoToQRCode 在二维码中心添加Logo
func addLogoToQRCode(qrPNG []byte, logoBase64 string, qrSize int, logoSizeRatio float64) ([]byte, error) {
	// 解码QR码图片
	qrImg, err := png.Decode(bytes.NewReader(qrPNG))
	if err != nil {
		return nil, err
	}

	// 解码Logo base64数据
	logoData := strings.Split(logoBase64, ",")
	if len(logoData) > 1 {
		logoBase64 = logoData[1] // 移除data:image/...;base64,前缀
	}

	logoBytes, err := base64.StdEncoding.DecodeString(logoBase64)
	if err != nil {
		return nil, err
	}

	// 自动检测图片格式并解码 - 支持所有常见格式
	logoImg, _, err := image.Decode(bytes.NewReader(logoBytes))
	if err != nil {
		return nil, err
	}

	// 创建新的图像
	bounds := qrImg.Bounds()
	newImg := image.NewRGBA(bounds)
	draw.Draw(newImg, bounds, qrImg, bounds.Min, draw.Src)

	// 计算Logo大小和位置
	logoSize := int(float64(qrSize) * logoSizeRatio)
	logoX := (qrSize - logoSize) / 2
	logoY := (qrSize - logoSize) / 2

	// 创建缩放后的Logo
	logoRect := image.Rect(0, 0, logoSize, logoSize)
	scaledLogo := image.NewRGBA(logoRect)

	// 简单的最近邻缩放
	logoBounds := logoImg.Bounds()
	scaleX := float64(logoBounds.Dx()) / float64(logoSize)
	scaleY := float64(logoBounds.Dy()) / float64(logoSize)

	for y := 0; y < logoSize; y++ {
		for x := 0; x < logoSize; x++ {
			srcX := int(float64(x)*scaleX) + logoBounds.Min.X
			srcY := int(float64(y)*scaleY) + logoBounds.Min.Y
			scaledLogo.Set(x, y, logoImg.At(srcX, srcY))
		}
	}

	// 将缩放后的Logo绘制到QR码中心
	logoPos := image.Rect(logoX, logoY, logoX+logoSize, logoY+logoSize)
	draw.Draw(newImg, logoPos, scaledLogo, image.Point{0, 0}, draw.Over)

	// 编码为PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, newImg)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// getBackgroundColor 获取背景色
func getBackgroundColor(img image.Image) color.RGBA {
	// 简单假设：取左上角的颜色作为背景色
	r, g, b, a := img.At(0, 0).RGBA()
	return color.RGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

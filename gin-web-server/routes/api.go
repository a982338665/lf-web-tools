package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"os"
	"os/exec"
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

// IDPhotoRequest 证件照生成请求
type IDPhotoRequest struct {
	ImageData       string  `json:"imageData" binding:"required"` // base64图片数据
	BackgroundColor string  `json:"backgroundColor"`              // 背景颜色，默认白色
	Width           int     `json:"width"`                        // 输出宽度
	Height          int     `json:"height"`                       // 输出高度
	Quality         float64 `json:"quality"`                      // 图片质量 0.1-1.0
	Format          string  `json:"format"`                       // 输出格式 jpeg/png
	BackgroundMode  string  `json:"backgroundMode"`               // 背景模式: auto(自动抠图), replace(简单替换)
}

// AIServiceRequest AI服务请求结构
type AIServiceRequest struct {
	ImageData       string  `json:"imageData"`
	BackgroundColor string  `json:"backgroundColor"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	Quality         float64 `json:"quality"`
	Format          string  `json:"format"`
	UseGPU          bool    `json:"useGPU"`
}

// AIServiceResponse AI服务响应结构
type AIServiceResponse struct {
	Success   bool   `json:"success"`
	ImageData string `json:"imageData"`
	FileSize  int64  `json:"fileSize"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	UsedGPU   bool   `json:"usedGPU"`
	Error     string `json:"error,omitempty"`
}

// IDPhotoResponse 证件照生成响应
type IDPhotoResponse struct {
	Success   bool   `json:"success"`
	DataURL   string `json:"dataUrl"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Size      int64  `json:"size"`      // 文件大小(字节)
	Timestamp string `json:"timestamp"` // 生成时间
	Message   string `json:"message,omitempty"`
}

// StandardSize 标准证件照尺寸
type StandardSize struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// StandardSizesResponse 标准尺寸列表响应
type StandardSizesResponse struct {
	Success bool           `json:"success"`
	Sizes   []StandardSize `json:"sizes"`
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

		// 获取标准证件照尺寸
		api.GET("/id-photo/sizes", func(c *gin.Context) {
			sizes := []StandardSize{
				{Name: "一寸", Width: 295, Height: 413},
				{Name: "二寸", Width: 413, Height: 579},
				{Name: "小二寸", Width: 390, Height: 567},
				{Name: "小一寸", Width: 260, Height: 378},
				{Name: "护照", Width: 358, Height: 441},
				{Name: "大一寸", Width: 480, Height: 640},
				{Name: "驾驶证", Width: 260, Height: 378},
				{Name: "社保卡", Width: 358, Height: 441},
			}

			c.JSON(http.StatusOK, StandardSizesResponse{
				Success: true,
				Sizes:   sizes,
			})
		})

		// 生成证件照API
		api.POST("/id-photo/generate", func(c *gin.Context) {
			var req IDPhotoRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, IDPhotoResponse{
					Success: false,
					Message: "无效的请求参数: " + err.Error(),
				})
				return
			}

			// 设置默认值
			if req.BackgroundColor == "" {
				req.BackgroundColor = "#FFFFFF"
			}
			if req.Width == 0 {
				req.Width = 295
			}
			if req.Height == 0 {
				req.Height = 413
			}
			if req.Quality == 0 {
				req.Quality = 0.8
			}
			if req.Format == "" {
				req.Format = "jpeg"
			}
			if req.BackgroundMode == "" {
				req.BackgroundMode = "auto"
			}

			// 验证参数
			if req.Width < 100 || req.Width > 2000 {
				c.JSON(http.StatusBadRequest, IDPhotoResponse{
					Success: false,
					Message: "宽度必须在100-2000之间",
				})
				return
			}

			if req.Height < 100 || req.Height > 2000 {
				c.JSON(http.StatusBadRequest, IDPhotoResponse{
					Success: false,
					Message: "高度必须在100-2000之间",
				})
				return
			}

			if req.Quality < 0.1 || req.Quality > 1.0 {
				c.JSON(http.StatusBadRequest, IDPhotoResponse{
					Success: false,
					Message: "质量参数必须在0.1-1.0之间",
				})
				return
			}

			// 处理证件照
			dataURL, size, err := processIDPhoto(req)
			if err != nil {
				c.JSON(http.StatusInternalServerError, IDPhotoResponse{
					Success: false,
					Message: "生成证件照失败: " + err.Error(),
				})
				return
			}

			// 返回结果
			c.JSON(http.StatusOK, IDPhotoResponse{
				Success:   true,
				DataURL:   dataURL,
				Width:     req.Width,
				Height:    req.Height,
				Format:    req.Format,
				Size:      size,
				Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			})
		})

		// 批量生成证件照API
		api.POST("/id-photo/batch", func(c *gin.Context) {
			var requests []IDPhotoRequest
			if err := c.ShouldBindJSON(&requests); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "无效的请求参数: " + err.Error(),
				})
				return
			}

			if len(requests) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "请求列表不能为空",
				})
				return
			}

			if len(requests) > 10 {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "批量处理最多支持10张图片",
				})
				return
			}

			var results []IDPhotoResponse
			var errors []string

			for i, req := range requests {
				// 设置默认值
				if req.BackgroundColor == "" {
					req.BackgroundColor = "#FFFFFF"
				}
				if req.Width == 0 {
					req.Width = 295
				}
				if req.Height == 0 {
					req.Height = 413
				}
				if req.Quality == 0 {
					req.Quality = 0.8
				}
				if req.Format == "" {
					req.Format = "jpeg"
				}

				// 处理证件照
				dataURL, size, err := processIDPhoto(req)
				if err != nil {
					errors = append(errors, fmt.Sprintf("第%d张图片处理失败: %s", i+1, err.Error()))
					results = append(results, IDPhotoResponse{
						Success: false,
						Message: err.Error(),
					})
				} else {
					results = append(results, IDPhotoResponse{
						Success:   true,
						DataURL:   dataURL,
						Width:     req.Width,
						Height:    req.Height,
						Format:    req.Format,
						Size:      size,
						Timestamp: time.Now().Format("2006-01-02 15:04:05"),
					})
				}
			}

			response := gin.H{
				"success":       len(errors) == 0,
				"results":       results,
				"total":         len(requests),
				"success_count": len(requests) - len(errors),
				"error_count":   len(errors),
				"timestamp":     time.Now().Format("2006-01-02 15:04:05"),
			}

			if len(errors) > 0 {
				response["errors"] = errors
			}

			c.JSON(http.StatusOK, response)
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

// processIDPhoto 处理证件照 - 使用AI服务
func processIDPhoto(req IDPhotoRequest) (string, int64, error) {
	// 优先使用AI服务
	return processIDPhotoAI(req)
}

// processIDPhotoAI 使用AI服务处理
func processIDPhotoAI(req IDPhotoRequest) (string, int64, error) {
	// 检查是否有Python环境和AI服务脚本
	pythonPath := "python"
	scriptPath := "./ai_service.py"

	// 尝试python3
	if _, err := exec.LookPath("python3"); err == nil {
		pythonPath = "python3"
	} else if _, err := exec.LookPath("python"); err != nil {
		// 回退到原始算法
		return processIDPhotoLegacy(req)
	}

	// 检查脚本是否存在
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return processIDPhotoLegacy(req)
	}

	// 准备AI服务请求
	aiReq := AIServiceRequest{
		ImageData:       req.ImageData,
		BackgroundColor: req.BackgroundColor,
		Width:           req.Width,
		Height:          req.Height,
		Quality:         req.Quality,
		Format:          req.Format,
		UseGPU:          true, // 默认尝试GPU
	}

	// 序列化请求
	reqJSON, err := json.Marshal(aiReq)
	if err != nil {
		fmt.Printf("AI请求序列化失败: %v\n", err)
		return processIDPhotoLegacy(req)
	}

	// 调用AI服务
	cmd := exec.Command(pythonPath, scriptPath)
	cmd.Stdin = strings.NewReader(string(reqJSON))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		fmt.Printf("AI服务调用失败: %s\n错误输出: %s\n", err.Error(), stderr.String())
		// AI服务失败，回退到原始算法
		return processIDPhotoLegacy(req)
	}

	// 解析AI服务响应
	var aiResp AIServiceResponse
	if err := json.Unmarshal(stdout.Bytes(), &aiResp); err != nil {
		fmt.Printf("AI响应解析失败: %v\n响应内容: %s\n", err, stdout.String())
		return processIDPhotoLegacy(req)
	}

	if !aiResp.Success {
		fmt.Printf("AI处理失败: %s\n", aiResp.Error)
		return processIDPhotoLegacy(req)
	}

	fmt.Printf("✅ AI处理成功! GPU: %v, 尺寸: %dx%d, 大小: %d bytes\n",
		aiResp.UsedGPU, aiResp.Width, aiResp.Height, aiResp.FileSize)

	return aiResp.ImageData, aiResp.FileSize, nil
}

// processIDPhotoLegacy 原始处理方法（回退）
func processIDPhotoLegacy(req IDPhotoRequest) (string, int64, error) {
	// 解码base64图片数据
	imageData := req.ImageData
	if strings.Contains(imageData, ",") {
		parts := strings.Split(imageData, ",")
		if len(parts) > 1 {
			imageData = parts[1] // 移除data:image/...;base64,前缀
		}
	}

	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		return "", 0, fmt.Errorf("无法解码图片数据: %v", err)
	}

	// 解码图片
	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return "", 0, fmt.Errorf("无法解码图片: %v", err)
	}

	// 解析背景颜色
	bgColor, err := parseHexColor(req.BackgroundColor)
	if err != nil {
		return "", 0, fmt.Errorf("无效的背景颜色: %v", err)
	}

	var outputImg *image.RGBA

	if req.BackgroundMode == "auto" {
		// 自动抠图模式
		outputImg = processImageWithBackgroundRemoval(img, bgColor, req.Width, req.Height)
	} else {
		// 简单背景替换模式
		outputImg = processImageSimple(img, bgColor, req.Width, req.Height)
	}

	// 编码输出图片
	var buf bytes.Buffer
	var size int64

	switch strings.ToLower(req.Format) {
	case "png":
		err = png.Encode(&buf, outputImg)
		if err != nil {
			return "", 0, fmt.Errorf("PNG编码失败: %v", err)
		}
		size = int64(buf.Len())
		dataURL := fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
		return dataURL, size, nil

	default: // jpeg
		err = jpeg.Encode(&buf, outputImg, &jpeg.Options{Quality: int(req.Quality * 100)})
		if err != nil {
			return "", 0, fmt.Errorf("JPEG编码失败: %v", err)
		}
		size = int64(buf.Len())
		dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", base64.StdEncoding.EncodeToString(buf.Bytes()))
		return dataURL, size, nil
	}
}

// processImageSimple 简单背景替换
func processImageSimple(img image.Image, bgColor color.RGBA, width, height int) *image.RGBA {
	// 创建输出图像
	outputImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			outputImg.Set(x, y, bgColor)
		}
	}

	// 计算原图的缩放和位置以适应目标尺寸
	srcBounds := img.Bounds()
	srcWidth := float64(srcBounds.Dx())
	srcHeight := float64(srcBounds.Dy())
	dstWidth := float64(width)
	dstHeight := float64(height)

	// 计算缩放比例，保持宽高比
	scaleX := dstWidth / srcWidth
	scaleY := dstHeight / srcHeight
	scale := math.Min(scaleX, scaleY)

	// 计算缩放后的尺寸
	scaledWidth := int(srcWidth * scale)
	scaledHeight := int(srcHeight * scale)

	// 计算居中位置
	offsetX := (width - scaledWidth) / 2
	offsetY := (height - scaledHeight) / 2

	// 绘制缩放后的图片
	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			// 计算在原图中的对应位置
			srcX := int(float64(x)/scale) + srcBounds.Min.X
			srcY := int(float64(y)/scale) + srcBounds.Min.Y

			if srcX < srcBounds.Max.X && srcY < srcBounds.Max.Y {
				outputImg.Set(x+offsetX, y+offsetY, img.At(srcX, srcY))
			}
		}
	}

	return outputImg
}

// processImageWithBackgroundRemoval 改进的自动抠图+背景替换
func processImageWithBackgroundRemoval(img image.Image, bgColor color.RGBA, width, height int) *image.RGBA {
	// 创建输出图像
	outputImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// 填充背景色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			outputImg.Set(x, y, bgColor)
		}
	}

	// 计算缩放比例
	srcBounds := img.Bounds()
	srcWidth := float64(srcBounds.Dx())
	srcHeight := float64(srcBounds.Dy())
	dstWidth := float64(width)
	dstHeight := float64(height)

	scale := math.Min(dstWidth/srcWidth, dstHeight/srcHeight)
	scaledWidth := int(srcWidth * scale)
	scaledHeight := int(srcHeight * scale)

	offsetX := (width - scaledWidth) / 2
	offsetY := (height - scaledHeight) / 2

	// 改进的背景检测：基于区域增长的抠图算法
	backgroundMap := make([][]bool, scaledHeight)
	for i := range backgroundMap {
		backgroundMap[i] = make([]bool, scaledWidth)
	}

	// 第一步：从边缘开始种子填充
	queue := make([]struct{ x, y int }, 0)
	visited := make([][]bool, scaledHeight)
	for i := range visited {
		visited[i] = make([]bool, scaledWidth)
	}

	edgeThickness := int(math.Max(2, math.Min(float64(scaledWidth), float64(scaledHeight))*0.02))

	// 添加边缘种子点
	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			if x < edgeThickness || x >= scaledWidth-edgeThickness ||
				y < edgeThickness || y >= scaledHeight-edgeThickness {
				backgroundMap[y][x] = true
				queue = append(queue, struct{ x, y int }{x, y})
			}
		}
	}

	// 第二步：区域增长算法
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if visited[current.y][current.x] {
			continue
		}
		visited[current.y][current.x] = true

		// 获取当前像素颜色
		srcX := int(float64(current.x)/scale) + srcBounds.Min.X
		srcY := int(float64(current.y)/scale) + srcBounds.Min.Y

		if srcX >= srcBounds.Max.X || srcY >= srcBounds.Max.Y {
			continue
		}

		pixel1 := img.At(srcX, srcY)
		r1, g1, b1, _ := pixel1.RGBA()
		r1_8 := uint8(r1 >> 8)
		g1_8 := uint8(g1 >> 8)
		b1_8 := uint8(b1 >> 8)

		// 检查8个方向的邻居
		directions := [][]int{{-1, -1}, {-1, 0}, {-1, 1}, {0, -1}, {0, 1}, {1, -1}, {1, 0}, {1, 1}}
		for _, dir := range directions {
			nx := current.x + dir[0]
			ny := current.y + dir[1]

			if nx >= 0 && nx < scaledWidth && ny >= 0 && ny < scaledHeight &&
				!visited[ny][nx] && !backgroundMap[ny][nx] {

				// 获取邻居像素颜色
				nSrcX := int(float64(nx)/scale) + srcBounds.Min.X
				nSrcY := int(float64(ny)/scale) + srcBounds.Min.Y

				if nSrcX < srcBounds.Max.X && nSrcY < srcBounds.Max.Y {
					pixel2 := img.At(nSrcX, nSrcY)
					r2, g2, b2, _ := pixel2.RGBA()
					r2_8 := uint8(r2 >> 8)
					g2_8 := uint8(g2 >> 8)
					b2_8 := uint8(b2 >> 8)

					// 改进的颜色相似度计算
					colorDistance := math.Sqrt(float64((r2_8-r1_8)*(r2_8-r1_8) +
						(g2_8-g1_8)*(g2_8-g1_8) +
						(b2_8-b1_8)*(b2_8-b1_8)))

					// 动态阈值：基于局部颜色方差
					threshold := calculateDynamicThresholdGo(img, nSrcX, nSrcY, srcBounds)

					if colorDistance < threshold {
						backgroundMap[ny][nx] = true
						queue = append(queue, struct{ x, y int }{nx, ny})
					}
				}
			}
		}
	}

	// 第三步：形态学操作去除噪点
	backgroundMap = morphologyOperationGo(backgroundMap, scaledWidth, scaledHeight)

	// 第四步：绘制结果，添加边缘抗锯齿
	for y := 0; y < scaledHeight; y++ {
		for x := 0; x < scaledWidth; x++ {
			srcX := int(float64(x)/scale) + srcBounds.Min.X
			srcY := int(float64(y)/scale) + srcBounds.Min.Y

			if srcX >= srcBounds.Max.X || srcY >= srcBounds.Max.Y {
				continue
			}

			if !backgroundMap[y][x] {
				// 前景像素，检查是否为边缘并应用抗锯齿
				pixel := img.At(srcX, srcY)
				r, g, b, a := pixel.RGBA()
				r8 := uint8(r >> 8)
				g8 := uint8(g >> 8)
				b8 := uint8(b >> 8)
				a8 := uint8(a >> 8)

				// 边缘抗锯齿处理
				alpha := calculateEdgeAlphaGo(backgroundMap, x, y, scaledWidth, scaledHeight)
				if alpha < 1.0 {
					// 混合前景和背景色
					finalR := uint8(float64(r8)*alpha + float64(bgColor.R)*(1-alpha))
					finalG := uint8(float64(g8)*alpha + float64(bgColor.G)*(1-alpha))
					finalB := uint8(float64(b8)*alpha + float64(bgColor.B)*(1-alpha))
					outputImg.Set(x+offsetX, y+offsetY, color.RGBA{finalR, finalG, finalB, a8})
				} else {
					outputImg.Set(x+offsetX, y+offsetY, color.RGBA{r8, g8, b8, a8})
				}
			}
		}
	}

	return outputImg
}

// calculateDynamicThresholdGo 计算动态阈值
func calculateDynamicThresholdGo(img image.Image, x, y int, bounds image.Rectangle) float64 {
	radius := 3
	var totalVariance float64
	count := 0

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nx := x + dx
			ny := y + dy
			if nx >= bounds.Min.X && nx < bounds.Max.X && ny >= bounds.Min.Y && ny < bounds.Max.Y {
				pixel := img.At(nx, ny)
				r, g, b, _ := pixel.RGBA()
				brightness := float64(uint8(r>>8)+uint8(g>>8)+uint8(b>>8)) / 3
				totalVariance += math.Abs(brightness - 128)
				count++
			}
		}
	}

	avgVariance := totalVariance / float64(count)
	return math.Max(15, math.Min(60, avgVariance*0.6))
}

// morphologyOperationGo 形态学操作
func morphologyOperationGo(binaryMap [][]bool, width, height int) [][]bool {
	// 腐蚀操作
	eroded := make([][]bool, height)
	for i := range eroded {
		eroded[i] = make([]bool, width)
		copy(eroded[i], binaryMap[i])
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			if binaryMap[y][x] {
				allBackground := true
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if !binaryMap[y+dy][x+dx] {
							allBackground = false
							break
						}
					}
					if !allBackground {
						break
					}
				}
				eroded[y][x] = allBackground
			}
		}
	}

	// 膨胀操作
	dilated := make([][]bool, height)
	for i := range dilated {
		dilated[i] = make([]bool, width)
		copy(dilated[i], eroded[i])
	}

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			if !eroded[y][x] {
				hasBackground := false
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if eroded[y+dy][x+dx] {
							hasBackground = true
							break
						}
					}
					if hasBackground {
						break
					}
				}
				dilated[y][x] = hasBackground
			}
		}
	}

	return dilated
}

// calculateEdgeAlphaGo 计算边缘透明度
func calculateEdgeAlphaGo(backgroundMap [][]bool, x, y, width, height int) float64 {
	backgroundCount := 0
	totalCount := 0

	// 检查3x3邻域
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx := x + dx
			ny := y + dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if backgroundMap[ny][nx] {
					backgroundCount++
				}
				totalCount++
			}
		}
	}

	if totalCount == 0 {
		return 1.0
	}

	// 如果周围有背景像素，则降低不透明度实现平滑过渡
	ratio := float64(backgroundCount) / float64(totalCount)
	if ratio > 0 {
		return 1.0 - (ratio * 0.8) // 保留一些不透明度
	}
	return 1.0
}

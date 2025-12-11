package routes

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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

type captchaItem struct {
	code      string
	expiresAt time.Time
}

type authToken struct {
	username  string
	expiresAt time.Time
}

type userRecord struct {
	Username     string `json:"username"`
	PasswordHash string `json:"passwordHash"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	UUID         string `json:"uuid"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

var (
	captchaStore = struct {
		sync.Mutex
		data map[string]captchaItem
	}{
		data: make(map[string]captchaItem),
	}
	userStore = struct {
		sync.Mutex
		data map[string]userRecord
	}{
		data: map[string]userRecord{
			"admin": {
				Username:     "admin",
				PasswordHash: hashPassword("admin123"),
				Email:        "",
				Phone:        "",
				UUID:         "",
				CreatedAt:    "",
				UpdatedAt:    "",
			},
		},
	}
	tokenStore = struct {
		sync.Mutex
		data map[string]authToken
	}{
		data: make(map[string]authToken),
	}
	tokenTTL      = 24 * time.Hour
	captchaTTL    = 5 * time.Minute
	userDataFile  = "data/users.json"
	loadUsersOnce sync.Once
)

// SetupAPIRoutes 设置API相关的路由
func SetupAPIRoutes(r *gin.Engine) {
	loadUsersOnce.Do(loadUsersFromFile)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.GET("/captcha", func(c *gin.Context) {
				captchaID, imageData, err := generateCaptcha()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "生成验证码失败",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"captchaId":   captchaID,
					"captchaData": imageData,
					"expiresIn":   int(captchaTTL.Seconds()),
				})
			})

			auth.POST("/register", func(c *gin.Context) {
				var req struct {
					Username    string `json:"username"`
					Password    string `json:"password"`
					Email       string `json:"email"`
					Phone       string `json:"phone"`
					CaptchaID   string `json:"captchaId"`
					CaptchaCode string `json:"captchaCode"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
					return
				}

				if !validateCaptcha(req.CaptchaID, req.CaptchaCode) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
					return
				}

				if len(req.Username) < 3 || len(req.Password) < 6 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "用户名或密码长度不符合要求"})
					return
				}

				if req.Email == "" || req.Phone == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱和手机号不能为空"})
					return
				}

				userStore.Lock()
				defer userStore.Unlock()
				if _, exists := userStore.data[req.Username]; exists {
					c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
					return
				}

				now := time.Now().Format("2006-01-02 15:04:05")
				record := userRecord{
					Username:     req.Username,
					PasswordHash: hashPassword(req.Password),
					Email:        req.Email,
					Phone:        req.Phone,
					UUID:         generateUUID(),
					CreatedAt:    now,
					UpdatedAt:    now,
				}

				userStore.data[req.Username] = record
				if err := persistUsersLocked(); err != nil {
					delete(userStore.data, req.Username)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "保存用户失败"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "注册成功"})
			})

			auth.POST("/login", func(c *gin.Context) {
				var req struct {
					Username    string `json:"username"`
					Password    string `json:"password"`
					CaptchaID   string `json:"captchaId"`
					CaptchaCode string `json:"captchaCode"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
					return
				}

				if !validateCaptcha(req.CaptchaID, req.CaptchaCode) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
					return
				}

				userStore.Lock()
				record, exists := userStore.data[req.Username]
				userStore.Unlock()
				if !exists || !verifyPassword(req.Password, record.PasswordHash) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
					return
				}

				token := generateToken()
				saveToken(token, req.Username)

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"token":   token,
					"user":    req.Username,
					"expires": time.Now().Add(tokenTTL).Format(time.RFC3339),
				})
			})

			auth.GET("/profile", func(c *gin.Context) {
				token := extractToken(c)
				username, ok := getTokenOwner(token)
				if !ok {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"success":  true,
					"username": username,
				})
			})

			auth.POST("/logout", func(c *gin.Context) {
				token := extractToken(c)
				if token == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "未提供令牌"})
					return
				}
				deleteToken(token)
				c.JSON(http.StatusOK, gin.H{"success": true})
			})

			auth.POST("/change-password", func(c *gin.Context) {
				token := extractToken(c)
				username, ok := getTokenOwner(token)
				if !ok {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
					return
				}

				var req struct {
					OldPassword string `json:"oldPassword"`
					NewPassword string `json:"newPassword"`
					CaptchaID   string `json:"captchaId"`
					CaptchaCode string `json:"captchaCode"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
					return
				}

				if !validateCaptcha(req.CaptchaID, req.CaptchaCode) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
					return
				}

				if len(req.NewPassword) < 6 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "新密码长度不能少于6位"})
					return
				}

				userStore.Lock()
				defer userStore.Unlock()

				record, exists := userStore.data[username]
				if !exists || !verifyPassword(req.OldPassword, record.PasswordHash) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "原密码错误"})
					return
				}

				record.PasswordHash = hashPassword(req.NewPassword)
				record.UpdatedAt = time.Now().Format("2006-01-02 15:04:05")
				userStore.data[username] = record
				if err := persistUsersLocked(); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "保存新密码失败"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "密码修改成功，请重新登录",
				})
			})
		}

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

func ensureUserFile() error {
	dir := filepath.Dir(userDataFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func loadUsersFromFile() {
	if err := ensureUserFile(); err != nil {
		return
	}

	userStore.Lock()
	defer userStore.Unlock()

	data, err := os.ReadFile(userDataFile)
	if err != nil {
		// 初始化写入默认用户
		_ = persistUsersLocked()
		return
	}

	var users map[string]userRecord
	if err := json.Unmarshal(data, &users); err != nil || len(users) == 0 {
		_ = persistUsersLocked()
		return
	}

	userStore.data = users
}

func persistUsersLocked() error {
	snapshot := make(map[string]userRecord, len(userStore.data))
	for k, v := range userStore.data {
		snapshot[k] = v
	}

	bytesData, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	if err := ensureUserFile(); err != nil {
		return err
	}

	tmpFile := userDataFile + ".tmp"
	if err := os.WriteFile(tmpFile, bytesData, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFile, userDataFile)
}

func generateCaptcha() (string, string, error) {
	code, err := randomString(4)
	if err != nil {
		return "", "", err
	}

	captchaID, err := randomString(16)
	if err != nil {
		return "", "", err
	}

	svg := fmt.Sprintf(`
<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40">
	<defs>
		<linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
			<stop offset="0%%" stop-color="#7f5af0"/>
			<stop offset="100%%" stop-color="#2cb67d"/>
		</linearGradient>
	</defs>
	<rect width="120" height="40" rx="6" fill="url(#g)"/>
	<text x="50%%" y="55%%" font-family="Arial, sans-serif" font-size="20" fill="#fff" text-anchor="middle" letter-spacing="4" font-weight="700">%s</text>
	<line x1="10" y1="10" x2="110" y2="12" stroke="#0ea5e9" stroke-width="2" opacity="0.6"/>
	<line x1="15" y1="30" x2="105" y2="26" stroke="#ef4444" stroke-width="2" opacity="0.6"/>
</svg>`, code)

	imageData := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))

	captchaStore.Lock()
	captchaStore.data[captchaID] = captchaItem{
		code:      strings.ToUpper(code),
		expiresAt: time.Now().Add(captchaTTL),
	}
	captchaStore.Unlock()

	return captchaID, imageData, nil
}

func randomString(length int) (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	result := make([]byte, length)
	_, err := rand.Read(result)
	if err != nil {
		return "", err
	}
	for i := 0; i < length; i++ {
		result[i] = chars[int(result[i])%len(chars)]
	}
	return string(result), nil
}

func validateCaptcha(id, code string) bool {
	if id == "" || code == "" {
		return false
	}

	captchaStore.Lock()
	defer captchaStore.Unlock()

	item, exists := captchaStore.data[id]
	if !exists {
		return false
	}
	delete(captchaStore.data, id)

	if time.Now().After(item.expiresAt) {
		return false
	}

	return strings.EqualFold(item.code, code)
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func verifyPassword(raw, hashed string) bool {
	return hashPassword(raw) == hashed
}

func generateToken() string {
	token, err := randomString(32)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return token
}

func generateUUID() string {
	u, err := randomString(32)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return u
}

func saveToken(token, username string) {
	tokenStore.Lock()
	defer tokenStore.Unlock()

	tokenStore.data[token] = authToken{
		username:  username,
		expiresAt: time.Now().Add(tokenTTL),
	}
}

func deleteToken(token string) {
	tokenStore.Lock()
	defer tokenStore.Unlock()
	delete(tokenStore.data, token)
}

func getTokenOwner(token string) (string, bool) {
	if token == "" {
		return "", false
	}

	tokenStore.Lock()
	defer tokenStore.Unlock()

	item, exists := tokenStore.data[token]
	if !exists {
		return "", false
	}
	if time.Now().After(item.expiresAt) {
		delete(tokenStore.data, token)
		return "", false
	}
	return item.username, true
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	token := c.GetHeader("X-Auth-Token")
	if token != "" {
		return token
	}
	return ""
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

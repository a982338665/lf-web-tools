package middleware

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CurlRequest 定义请求体结构
type CurlRequest struct {
	CurlParam string `json:"curlParam" binding:"required"`
}

// CurlResponse 定义响应体结构
type CurlResponse struct {
	ExecutionTime   string            `json:"executionTime"`
	StatusCode      int               `json:"statusCode"`
	ResponseBody    string            `json:"responseBody"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	Error           string            `json:"error,omitempty"`
}

// CorsProxyMiddleware 返回一个处理CORS代理请求的中间件
func CorsProxyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许所有来源的CORS请求
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// 继续处理请求
		c.Next()
	}
}

// HandleCurlProxy 处理curl代理请求
func HandleCurlProxy(c *gin.Context) {
	var request CurlRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 解析curl命令并执行HTTP请求
	startTime := time.Now()
	statusCode, responseBody, responseHeaders, err := executeCurlAsHTTP(request.CurlParam)
	executionTime := time.Since(startTime).String()

	response := CurlResponse{
		ExecutionTime:   executionTime,
		StatusCode:      statusCode,
		ResponseBody:    responseBody,
		ResponseHeaders: responseHeaders,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(http.StatusOK, response)
}

// CurlCommand 表示解析后的curl命令
type CurlCommand struct {
	Method      string
	URL         string
	Headers     map[string]string
	Data        string
	Insecure    bool
	FormData    map[string]string
	QueryParams map[string]string
}

// executeCurlAsHTTP 将curl命令解析为HTTP请求并执行
func executeCurlAsHTTP(curlCmd string) (int, string, map[string]string, error) {
	// 确保命令以curl开头
	if !strings.HasPrefix(curlCmd, "curl ") {
		return 0, "", nil, fmt.Errorf("command must start with 'curl'")
	}

	// 将多行命令合并为一行（去掉行尾的反斜杠和换行符）
	curlCmd = strings.ReplaceAll(curlCmd, "\\\n", " ")
	curlCmd = strings.ReplaceAll(curlCmd, "\\\r\n", " ")

	// 解析curl命令
	cmd, err := parseCurlCommand(curlCmd)
	if err != nil {
		return 0, "", nil, err
	}

	// 创建HTTP客户端
	transport := &http.Transport{}
	
	// 如果指定了insecure选项，则跳过TLS证书验证
	if cmd.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	// 准备请求体
	var reqBody io.Reader
	if cmd.Data != "" {
		reqBody = strings.NewReader(cmd.Data)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(cmd.Method, cmd.URL, reqBody)
	if err != nil {
		return 0, "", nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 添加请求头
	for key, value := range cmd.Headers {
		req.Header.Set(key, value)
	}

	// 执行请求
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", nil, fmt.Errorf("执行HTTP请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", headerToMap(resp.Header), fmt.Errorf("读取响应体失败: %v", err)
	}

	return resp.StatusCode, string(bodyBytes), headerToMap(resp.Header), nil
}

// parseCurlCommand 解析curl命令
func parseCurlCommand(curlCmd string) (*CurlCommand, error) {
	cmd := &CurlCommand{
		Method:      "GET",
		Headers:     make(map[string]string),
		FormData:    make(map[string]string),
		QueryParams: make(map[string]string),
	}

	// 提取URL
	urlPattern := regexp.MustCompile(`curl\s+['"]([^'"]+)['"]`)
	urlMatches := urlPattern.FindStringSubmatch(curlCmd)
	if len(urlMatches) < 2 {
		// 尝试不带引号的URL
		urlPattern = regexp.MustCompile(`curl\s+(\S+)`)
		urlMatches = urlPattern.FindStringSubmatch(curlCmd)
		if len(urlMatches) < 2 {
			return nil, fmt.Errorf("无法解析URL")
		}
	}
	cmd.URL = urlMatches[1]

	// 提取请求方法
	methodPattern := regexp.MustCompile(`-X\s+(\S+)`)
	methodMatches := methodPattern.FindStringSubmatch(curlCmd)
	if len(methodMatches) >= 2 {
		cmd.Method = methodMatches[1]
	}

	// 提取请求头
	headerPattern := regexp.MustCompile(`-H\s+['"]([^'"]+)['"]`)
	headerMatches := headerPattern.FindAllStringSubmatch(curlCmd, -1)
	for _, match := range headerMatches {
		if len(match) >= 2 {
			headerParts := strings.SplitN(match[1], ":", 2)
			if len(headerParts) == 2 {
				cmd.Headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
			}
		}
	}

	// 提取请求体数据 - 支持多种数据格式
	// 1. --data-raw
	dataRawPattern := regexp.MustCompile(`--data-raw\s+['"]([^'"]+)['"]`)
	dataRawMatches := dataRawPattern.FindStringSubmatch(curlCmd)
	if len(dataRawMatches) >= 2 {
		cmd.Data = dataRawMatches[1]
	}
	
	// 2. --data 或 -d
	if cmd.Data == "" {
		dataPattern := regexp.MustCompile(`(?:--data|-d)\s+['"]([^'"]+)['"]`)
		dataMatches := dataPattern.FindStringSubmatch(curlCmd)
		if len(dataMatches) >= 2 {
			cmd.Data = dataMatches[1]
		}
	}
	
	// 3. --json
	if cmd.Data == "" {
		jsonPattern := regexp.MustCompile(`--json\s+['"]([^'"]+)['"]`)
		jsonMatches := jsonPattern.FindStringSubmatch(curlCmd)
		if len(jsonMatches) >= 2 {
			cmd.Data = jsonMatches[1]
			cmd.Headers["Content-Type"] = "application/json"
		}
	}
	
	// 如果有数据，设置适当的Content-Type和Method
	if cmd.Data != "" {
		// 如果没有明确指定Content-Type，根据数据格式推断
		if _, exists := cmd.Headers["Content-Type"]; !exists {
			if strings.HasPrefix(cmd.Data, "{") && strings.HasSuffix(cmd.Data, "}") {
				cmd.Headers["Content-Type"] = "application/json"
			} else {
				cmd.Headers["Content-Type"] = "application/x-www-form-urlencoded"
			}
		}
		// 如果是POST请求但没有指定方法，则设置为POST
		if cmd.Method == "GET" {
			cmd.Method = "POST"
		}
	}

	// 检查是否有--insecure选项
	cmd.Insecure = strings.Contains(curlCmd, "--insecure") || strings.Contains(curlCmd, "-k")

	return cmd, nil
}

// headerToMap 将HTTP头转换为map
func headerToMap(header http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range header {
		if len(values) > 0 {
			result[key] = strings.Join(values, ", ")
		}
	}
	return result
}

// RegisterCorsProxyRoutes 注册CORS代理路由到Gin引擎
func RegisterCorsProxyRoutes(r *gin.Engine) {
	r.Use(CorsProxyMiddleware())
	r.POST("/cors-proxy", HandleCurlProxy)
}

// StartStandalone 启动独立的CORS代理服务器
func StartStandalone(port string) {
	if port == "" {
		port = "8081" // 默认端口
	}

	r := gin.Default()
	RegisterCorsProxyRoutes(r)
	
	fmt.Printf("CORS Proxy server starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

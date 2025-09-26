package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
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
	ExitCode        int               `json:"exitCode"`
	Result          string            `json:"result"`
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

	// 执行curl命令并计时
	startTime := time.Now()
	result, exitCode, responseBody, responseHeaders, err := executeCurl(request.CurlParam)
	executionTime := time.Since(startTime).String()

	response := CurlResponse{
		ExecutionTime:   executionTime,
		ExitCode:        exitCode,
		Result:          result,
		ResponseBody:    responseBody,
		ResponseHeaders: responseHeaders,
	}

	if err != nil {
		response.Error = err.Error()
	}

	c.JSON(http.StatusOK, response)
}

// executeCurl 执行curl命令并返回结果
func executeCurl(curlCmd string) (string, int, string, map[string]string, error) {
	// 确保命令以curl开头
	if !strings.HasPrefix(curlCmd, "curl ") {
		return "", -1, "", nil, fmt.Errorf("command must start with 'curl'")
	}

	// 添加-i参数获取响应头信息，如果没有的话
	if !strings.Contains(curlCmd, " -i ") && !strings.Contains(curlCmd, " --include ") {
		curlCmd = strings.Replace(curlCmd, "curl ", "curl -i ", 1)
	}

	// 根据操作系统确定命令执行方式
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", curlCmd)
	} else {
		cmd = exec.Command("sh", "-c", curlCmd)
	}

	// 捕获标准输出和错误输出
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	// 获取输出
	result := stdout.String()
	if result == "" && stderr.String() != "" {
		result = stderr.String()
	}

	// 分离响应头和响应体
	responseBody, responseHeaders := parseResponse(result)

	return result, exitCode, responseBody, responseHeaders, err
}

// parseResponse 解析响应，分离头部和主体
func parseResponse(response string) (string, map[string]string) {
	headers := make(map[string]string)
	
	// 查找头部和主体的分隔点（空行）
	parts := strings.SplitN(response, "\r\n\r\n", 2)
	if len(parts) < 2 {
		// 尝试不同的换行符
		parts = strings.SplitN(response, "\n\n", 2)
		if len(parts) < 2 {
			// 如果没有找到分隔符，则认为整个响应都是主体
			return response, headers
		}
	}

	// 解析头部
	headerLines := strings.Split(parts[0], "\n")
	for i, line := range headerLines {
		// 跳过第一行（HTTP状态行）
		if i == 0 {
			continue
		}
		
		// 解析头部字段
		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])
			headers[key] = value
		}
	}

	// 返回主体和解析的头部
	return parts[1], headers
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

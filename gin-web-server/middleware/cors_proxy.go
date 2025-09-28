package middleware

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
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
	clientIP := c.ClientIP()
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	fmt.Printf("[CORS-PROXY] [%s] 收到请求 - 客户端IP: %s, 请求方法: %s, 请求路径: %s\n",
		requestID, clientIP, c.Request.Method, c.Request.URL.Path)

	var request CurlRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Printf("[CORS-PROXY] [%s] 请求体解析失败: %v\n", requestID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 打印CURL命令（截断过长的命令）
	curlCmd := request.CurlParam
	if len(curlCmd) > 100 {
		fmt.Printf("[CORS-PROXY] [%s] CURL命令: %s...(已截断，总长度: %d)\n",
			requestID, curlCmd[:100], len(curlCmd))
	} else {
		fmt.Printf("[CORS-PROXY] [%s] CURL命令: %s\n", requestID, curlCmd)
	}

	// 解析curl命令并执行HTTP请求
	startTime := time.Now()
	fmt.Printf("[CORS-PROXY] [%s] 开始执行请求...\n", requestID)

	statusCode, responseBody, responseHeaders, err := executeCurlAsHTTP(request.CurlParam)
	executionTime := time.Since(startTime)

	// 打印响应信息
	if err != nil {
		fmt.Printf("[CORS-PROXY] [%s] 请求失败: %v, 耗时: %v\n", requestID, err, executionTime)
	} else {
		// 截断响应体以避免日志过长
		bodyPreview := responseBody
		if len(bodyPreview) > 200 {
			bodyPreview = bodyPreview[:200] + "...(已截断)"
		}

		fmt.Printf("[CORS-PROXY] [%s] 请求成功: 状态码=%d, 响应头数量=%d, 响应体长度=%d, 耗时=%v\n",
			requestID, statusCode, len(responseHeaders), len(responseBody), executionTime)
		fmt.Printf("[CORS-PROXY] [%s] 响应体预览: %s\n", requestID, bodyPreview)
	}

	response := CurlResponse{
		ExecutionTime:   executionTime.String(),
		StatusCode:      statusCode,
		ResponseBody:    responseBody,
		ResponseHeaders: responseHeaders,
	}

	if err != nil {
		response.Error = err.Error()
	}

	fmt.Printf("[CORS-PROXY] [%s] 请求处理完成, 总耗时: %v\n", requestID, executionTime)
	c.JSON(http.StatusOK, response)
}

// CurlCommand 表示解析后的curl命令
type CurlCommand struct {
	Method          string
	URL             string
	Headers         map[string]string
	Data            string
	Insecure        bool
	FormData        map[string]string
	QueryParams     map[string]string
	Timeout         int               // 连接超时时间（秒）
	FollowRedirects bool              // 是否跟随重定向
	Auth            string            // 基本认证信息
	Cookies         map[string]string // Cookie信息
}

// executeCurlAsHTTP 将curl命令解析为HTTP请求并执行
func executeCurlAsHTTP(curlCmd string) (int, string, map[string]string, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())

	// 确保命令以curl开头
	if !strings.HasPrefix(curlCmd, "curl ") {
		return 0, "", nil, fmt.Errorf("command must start with 'curl'")
	}

	// 将多行命令合并为一行（去掉行尾的反斜杠和换行符）
	curlCmd = strings.ReplaceAll(curlCmd, "\\\n", " ")
	curlCmd = strings.ReplaceAll(curlCmd, "\\\r\n", " ")

	fmt.Printf("[CORS-PROXY] [%s] 开始解析CURL命令...\n", requestID)

	// 解析curl命令
	cmd, err := parseCurlCommand(curlCmd)
	if err != nil {
		fmt.Printf("[CORS-PROXY] [%s] 解析CURL命令失败: %v\n", requestID, err)
		return 0, "", nil, err
	}

	fmt.Printf("[CORS-PROXY] [%s] 解析结果: 方法=%s, URL=%s, 数据长度=%d, 头部数量=%d, Insecure=%v\n",
		requestID, cmd.Method, cmd.URL, len(cmd.Data), len(cmd.Headers), cmd.Insecure)

	// 创建HTTP客户端
	transport := &http.Transport{
		// 设置更宽松的连接超时
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 如果指定了insecure选项，则跳过TLS证书验证
	if cmd.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		fmt.Printf("[CORS-PROXY] [%s] 启用不安全模式，跳过TLS证书验证\n", requestID)
	}

	// 检查环境变量中是否设置了HTTP代理
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	if httpProxy != "" || httpsProxy != "" {
		proxyURL, err := url.Parse(httpProxy)
		if err == nil && httpProxy != "" {
			transport.Proxy = http.ProxyURL(proxyURL)
			fmt.Printf("[CORS-PROXY] [%s] 使用HTTP代理: %s\n", requestID, httpProxy)
		}

		if httpsProxy != "" && httpProxy != httpsProxy {
			proxyURL, err = url.Parse(httpsProxy)
			if err == nil {
				transport.Proxy = http.ProxyURL(proxyURL)
				fmt.Printf("[CORS-PROXY] [%s] 使用HTTPS代理: %s\n", requestID, httpsProxy)
			}
		}
	}

	// 设置客户端选项
	client := &http.Client{
		Timeout:   time.Duration(cmd.Timeout) * time.Second,
		Transport: transport,
	}

	// 设置重定向策略
	if cmd.FollowRedirects {
		fmt.Printf("[CORS-PROXY] [%s] 启用跟随重定向\n", requestID)
	} else {
		// 不跟随重定向
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// 准备请求体
	var reqBody io.Reader
	if cmd.Data != "" {
		reqBody = strings.NewReader(cmd.Data)
		fmt.Printf("[CORS-PROXY] [%s] 设置请求体数据 (%d字节)\n", requestID, len(cmd.Data))
	}

	// 创建HTTP请求
	fmt.Printf("[CORS-PROXY] [%s] 创建HTTP请求: %s %s\n", requestID, cmd.Method, cmd.URL)
	req, err := http.NewRequest(cmd.Method, cmd.URL, reqBody)
	if err != nil {
		fmt.Printf("[CORS-PROXY] [%s] 创建HTTP请求失败: %v\n", requestID, err)
		return 0, "", nil, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 添加请求头
	for key, value := range cmd.Headers {
		req.Header.Set(key, value)
		fmt.Printf("[CORS-PROXY] [%s] 添加请求头: %s: %s\n", requestID, key, value)
	}

	// 如果没有设置User-Agent，添加默认User-Agent
	if _, exists := cmd.Headers["User-Agent"]; !exists {
		defaultUserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
		req.Header.Set("User-Agent", defaultUserAgent)
		fmt.Printf("[CORS-PROXY] [%s] 添加默认User-Agent: %s\n", requestID, defaultUserAgent)
	}

	// 执行请求
	fmt.Printf("[CORS-PROXY] [%s] 发送HTTP请求: %s %s\n", requestID, req.Method, req.URL.String())

	// 打印所有请求头，便于调试
	fmt.Printf("[CORS-PROXY] [%s] 请求头详情:\n", requestID)
	for k, v := range req.Header {
		fmt.Printf("[CORS-PROXY] [%s]   %s: %s\n", requestID, k, strings.Join(v, ", "))
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	requestDuration := time.Since(startTime)

	if err != nil {
		fmt.Printf("[CORS-PROXY] [%s] 执行HTTP请求失败: %v, 耗时: %v\n", requestID, err, requestDuration)
		// 检查是否是DNS解析错误
		if strings.Contains(err.Error(), "lookup") && strings.Contains(err.Error(), "no such host") {
			return 0, "", nil, fmt.Errorf("DNS解析失败，无法找到主机: %v", err)
		}
		// 检查是否是连接超时
		if strings.Contains(err.Error(), "timeout") {
			return 0, "", nil, fmt.Errorf("请求超时: %v", err)
		}
		// 检查是否是TLS错误
		if strings.Contains(err.Error(), "tls") || strings.Contains(err.Error(), "certificate") {
			return 0, "", nil, fmt.Errorf("TLS/SSL错误: %v，请尝试添加--insecure选项", err)
		}
		return 0, "", nil, fmt.Errorf("执行HTTP请求失败: %v", err)
	}

	fmt.Printf("[CORS-PROXY] [%s] 收到响应: 状态码=%d, 耗时=%v\n", requestID, resp.StatusCode, requestDuration)

	// 打印响应头，便于调试
	fmt.Printf("[CORS-PROXY] [%s] 响应头详情:\n", requestID)
	for k, v := range resp.Header {
		fmt.Printf("[CORS-PROXY] [%s]   %s: %s\n", requestID, k, strings.Join(v, ", "))
	}

	defer resp.Body.Close()

	// 读取响应体
	fmt.Printf("[CORS-PROXY] [%s] 读取响应体...\n", requestID)
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[CORS-PROXY] [%s] 读取响应体失败: %v\n", requestID, err)
		return resp.StatusCode, "", headerToMap(resp.Header), fmt.Errorf("读取响应体失败: %v", err)
	}

	responseHeaders := headerToMap(resp.Header)
	fmt.Printf("[CORS-PROXY] [%s] 响应处理完成: 状态码=%d, 响应体大小=%d字节, 响应头数量=%d\n",
		requestID, resp.StatusCode, len(bodyBytes), len(responseHeaders))

	return resp.StatusCode, string(bodyBytes), responseHeaders, nil
}

// parseCurlCommand 解析curl命令
func parseCurlCommand(curlCmd string) (*CurlCommand, error) {
	cmd := &CurlCommand{
		Method:          "GET",
		Headers:         make(map[string]string),
		FormData:        make(map[string]string),
		QueryParams:     make(map[string]string),
		Cookies:         make(map[string]string),
		FollowRedirects: false,
		Timeout:         30, // 默认30秒超时
	}

	// 先提取请求方法，避免将-X误认为是URL的一部分
	methodPattern := regexp.MustCompile(`-X\s+(\S+)`)
	methodMatches := methodPattern.FindStringSubmatch(curlCmd)
	if len(methodMatches) >= 2 {
		cmd.Method = methodMatches[1]
	}

	// 提取URL (在处理完方法后)
	// 尝试直接匹配http或https URL
	urlPattern := regexp.MustCompile(`(https?://[^\s'"]+)`)
	urlMatches := urlPattern.FindStringSubmatch(curlCmd)

	if len(urlMatches) >= 1 {
		cmd.URL = urlMatches[1]
	} else {
		// 如果没有找到http开头的URL，尝试其他方式
		quotedUrlPattern := regexp.MustCompile(`curl\s+['"]([^'"]+)['"]`)
		quotedUrlMatches := quotedUrlPattern.FindStringSubmatch(curlCmd)

		if len(quotedUrlMatches) >= 2 {
			cmd.URL = quotedUrlMatches[1]
		} else {
			// 分析命令行参数，找到不是选项的参数作为URL
			parts := strings.Fields(curlCmd)
			for i, part := range parts {
				// 跳过curl本身和所有选项及其参数
				if part == "curl" || strings.HasPrefix(part, "-") {
					continue
				}
				// 跳过选项的参数
				if i > 0 && (parts[i-1] == "-X" || parts[i-1] == "-H" ||
					parts[i-1] == "-d" || parts[i-1] == "--data" ||
					parts[i-1] == "--data-raw") {
					continue
				}
				// 找到的第一个非选项参数可能是URL
				cmd.URL = part
				break
			}

			// 如果仍然没有找到URL
			if cmd.URL == "" {
				return nil, fmt.Errorf("无法解析URL")
			}
		}
	}

	// 提取请求头 - 支持多种引号格式
	// 单引号
	headerPattern1 := regexp.MustCompile(`-H\s+'([^']+)'`)
	headerMatches1 := headerPattern1.FindAllStringSubmatch(curlCmd, -1)
	for _, match := range headerMatches1 {
		if len(match) >= 2 {
			headerParts := strings.SplitN(match[1], ":", 2)
			if len(headerParts) == 2 {
				cmd.Headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
			}
		}
	}

	// 双引号
	headerPattern2 := regexp.MustCompile(`-H\s+"([^"]+)"`)
	headerMatches2 := headerPattern2.FindAllStringSubmatch(curlCmd, -1)
	for _, match := range headerMatches2 {
		if len(match) >= 2 {
			headerParts := strings.SplitN(match[1], ":", 2)
			if len(headerParts) == 2 {
				cmd.Headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
			}
		}
	}

	// 提取请求体数据 - 支持多种数据格式和引号
	// 1. --data-raw (单引号)
	dataRawPattern1 := regexp.MustCompile(`--data-raw\s+'([^']*(?:\\'[^']*)*)'`)
	dataRawMatches1 := dataRawPattern1.FindStringSubmatch(curlCmd)
	if len(dataRawMatches1) >= 2 {
		cmd.Data = dataRawMatches1[1]
		// 处理转义的单引号
		cmd.Data = strings.ReplaceAll(cmd.Data, "\\'", "'")
	}

	// --data-raw (双引号)
	if cmd.Data == "" {
		dataRawPattern2 := regexp.MustCompile(`--data-raw\s+"([^"]*(?:\\"[^"]*)*)"`)
		dataRawMatches2 := dataRawPattern2.FindStringSubmatch(curlCmd)
		if len(dataRawMatches2) >= 2 {
			cmd.Data = dataRawMatches2[1]
			// 处理转义的双引号
			cmd.Data = strings.ReplaceAll(cmd.Data, "\\\"", "\"")
		}
	}

	// 2. --data 或 -d (单引号)
	if cmd.Data == "" {
		dataPattern1 := regexp.MustCompile(`(?:--data|-d)\s+'([^']*(?:\\'[^']*)*)'`)
		dataMatches1 := dataPattern1.FindStringSubmatch(curlCmd)
		if len(dataMatches1) >= 2 {
			cmd.Data = dataMatches1[1]
			cmd.Data = strings.ReplaceAll(cmd.Data, "\\'", "'")
		}
	}

	// --data 或 -d (双引号)
	if cmd.Data == "" {
		dataPattern2 := regexp.MustCompile(`(?:--data|-d)\s+"([^"]*(?:\\"[^"]*)*)"`)
		dataMatches2 := dataPattern2.FindStringSubmatch(curlCmd)
		if len(dataMatches2) >= 2 {
			cmd.Data = dataMatches2[1]
			cmd.Data = strings.ReplaceAll(cmd.Data, "\\\"", "\"")
		}
	}

	// 3. --json (单引号)
	if cmd.Data == "" {
		jsonPattern1 := regexp.MustCompile(`--json\s+'([^']*(?:\\'[^']*)*)'`)
		jsonMatches1 := jsonPattern1.FindStringSubmatch(curlCmd)
		if len(jsonMatches1) >= 2 {
			cmd.Data = jsonMatches1[1]
			cmd.Data = strings.ReplaceAll(cmd.Data, "\\'", "'")
			cmd.Headers["Content-Type"] = "application/json"
		}
	}

	// --json (双引号)
	if cmd.Data == "" {
		jsonPattern2 := regexp.MustCompile(`--json\s+"([^"]*(?:\\"[^"]*)*)"`)
		jsonMatches2 := jsonPattern2.FindStringSubmatch(curlCmd)
		if len(jsonMatches2) >= 2 {
			cmd.Data = jsonMatches2[1]
			cmd.Data = strings.ReplaceAll(cmd.Data, "\\\"", "\"")
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

	// 检查是否有超时选项
	timeoutPattern := regexp.MustCompile(`--connect-timeout\s+(\d+)`)
	timeoutMatches := timeoutPattern.FindStringSubmatch(curlCmd)
	if len(timeoutMatches) >= 2 {
		if timeout, err := strconv.Atoi(timeoutMatches[1]); err == nil {
			cmd.Timeout = timeout
		}
	}

	// 检查是否有跟随重定向选项
	cmd.FollowRedirects = strings.Contains(curlCmd, "--location") || strings.Contains(curlCmd, "-L")

	// 检查是否有用户代理选项
	userAgentPattern := regexp.MustCompile(`-A\s+['"]([^'"]+)['"]`)
	userAgentMatches := userAgentPattern.FindStringSubmatch(curlCmd)
	if len(userAgentMatches) >= 2 {
		cmd.Headers["User-Agent"] = userAgentMatches[1]
	}

	// 检查是否有基本认证选项
	authPattern := regexp.MustCompile(`-u\s+['"]?([^'"]+)['"]?`)
	authMatches := authPattern.FindStringSubmatch(curlCmd)
	if len(authMatches) >= 2 {
		auth := authMatches[1]
		cmd.Auth = auth
		// 如果包含:，则是用户名:密码格式
		if strings.Contains(auth, ":") {
			parts := strings.SplitN(auth, ":", 2)
			if len(parts) == 2 {
				// 将用户名:密码编码为Base64
				authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
				cmd.Headers["Authorization"] = authHeader
			}
		}
	}

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
	r.POST("/cors-proxy", HandleCurlProxy)
}

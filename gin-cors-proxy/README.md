# Gin CORS Proxy

这是一个基于Gin框架的CORS代理服务，可以作为独立服务运行，也可以作为组件集成到现有的Gin应用中。

## 功能特点

- 接收curl命令并解析为原生HTTP请求，解决前端跨域问题
- 返回详细的执行信息，包括执行时间、响应体、响应头等
- 支持作为独立服务运行
- 支持作为Gin组件集成到现有应用
- 不依赖系统curl命令，完全使用Go原生HTTP客户端

## 使用方法

### 作为独立服务运行

1. 直接运行启动脚本：

```bash
./start.bat  # Windows
```

或者

```bash
go run main.go  # 默认端口8081
go run main.go -port=8082  # 自定义端口
```

2. 向服务发送请求：

```
POST http://localhost:8081/cors-proxy
Content-Type: application/json

{
  "curlParam": "curl -X GET https://httpbin.org/get"
}
```

### 作为Gin组件集成

在你的Gin应用中，导入middleware包并注册路由：

```go
import "github.com/lf-web-tools/gin-cors-proxy/middleware"

func main() {
    r := gin.Default()
    
    // 注册CORS代理路由
    middleware.RegisterCorsProxyRoutes(r)
    
    // 你的其他路由...
    
    r.Run(":8080")
}
```

## API 说明

### 请求

- 路径: `/cors-proxy`
- 方法: `POST`
- 内容类型: `application/json`
- 请求体:

```json
{
  "curlParam": "curl -X GET https://httpbin.org/get"
}
```

### 响应

```json
{
  "executionTime": "235.0394ms",
  "statusCode": 200,
  "responseBody": "{\n  \"args\": {}, \n  \"headers\": {...}, \n  \"origin\": \"...\", \n  \"url\": \"https://httpbin.org/get\"\n}\n",
  "responseHeaders": {
    "Content-Type": "application/json",
    "Date": "...",
    "Content-Length": "..."
  }
}
```

## 支持的curl选项

- `-X, --request`: 指定HTTP请求方法（GET、POST、PUT、DELETE等）
- `-H, --header`: 设置请求头
- `-d, --data`: 设置请求体数据
- `--data-raw`: 设置原始请求体数据
- `--json`: 设置JSON格式的请求体数据
- `-k, --insecure`: 允许不安全的SSL连接

## 注意事项

- 不需要安装curl命令行工具，完全使用Go原生HTTP客户端
- 为安全起见，建议在内部网络或受信任的环境中使用
- 默认情况下，服务允许所有来源的CORS请求
- 当前版本仅支持基本的curl命令解析，不支持所有curl选项

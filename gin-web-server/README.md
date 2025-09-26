# Gin Web Server

这是一个基于Gin框架的Web服务器，提供了静态文件服务、API接口、WebSocket支持和HTML模板渲染功能。

## 功能

- 静态文件服务：访问 `/static/` 路径下的文件
- API接口：
  - GET `/api/time` - 获取服务器当前时间
  - GET `/api/info` - 获取服务器信息
- WebSocket支持：
  - `/ws` - WebSocket连接点
- 模板渲染：
  - GET `/hello` - 显示欢迎页面

## 如何运行

### 方法1：使用批处理脚本

直接运行 `start.bat` 文件。

### 方法2：命令行运行

```bash
# 进入项目目录
cd gin-web-server

# 整理依赖
go mod tidy

# 运行服务器
go run main.go
```

## 访问地址

服务器启动后，可以通过以下地址访问：

- 主页：http://localhost:8080/
- 静态文件：http://localhost:8080/static/
- API时间接口：http://localhost:8080/api/time
- API信息接口：http://localhost:8080/api/info
- WebSocket测试：http://localhost:8080/static/socket_test.html
- 欢迎页面：http://localhost:8080/hello

## 目录结构

- `main.go` - 主程序入口
- `routes/` - 路由定义
  - `api.go` - API路由
  - `websocket.go` - WebSocket路由
  - `pages.go` - 页面路由
- `static/` - 静态文件目录
- `templates/` - HTML模板目录

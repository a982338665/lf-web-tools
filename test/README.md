# 端口检测测试服务器

这是一个简单的Node.js测试服务器，用于测试网络信息检测器的本地端口检测功能。

## 服务器配置

- **WebSocket服务器**: 端口 3000
- **HTTP服务器1**: 端口 8080  
- **HTTP服务器2**: 端口 8000

## 快速启动

```bash
# 1. 进入test目录
cd test

# 2. 安装依赖
npm install

# 3. 启动服务器
npm start
```

## 服务器功能

### WebSocket服务器 (端口3000)
- 地址: `ws://localhost:3000`
- 功能: 接收连接，发送心跳，回显消息
- 测试: 可以用浏览器的WebSocket连接测试

### HTTP服务器 (端口8080)
- 地址: `http://localhost:8080`
- 接口:
  - `GET /` - 首页
  - `GET /api/status` - 服务器状态
  - `GET /api/time` - 当前时间
  - `GET /favicon.ico` - 网站图标

### HTTP服务器 (端口8000)  
- 地址: `http://localhost:8000`
- 接口:
  - `GET /` - 首页
  - `GET /api/info` - 服务器信息
  - `GET /api/health` - 健康检查
  - `GET /favicon.ico` - 网站图标

## 使用说明

1. 启动服务器后，打开网络信息检测器页面
2. 查看端口3000、8000、8080是否被正确检测为开放状态
3. 按F12查看控制台日志，观察检测过程
4. 按Ctrl+C停止服务器

## 预期测试结果

启动服务器后，在网络信息检测器中应该看到:
- 端口 3000 (Node.js Dev/TCP): 🟢 WebSocket检测到开放
- 端口 8000 (Python HTTP/TCP): 🟢 HTTP检测到开放  
- 端口 8080 (HTTP-Alt/TCP): 🟢 HTTP检测到开放

## 额外工具

### 1. WebSocket连接测试
- 文件: `websocket-test.html`
- 功能: 直接在浏览器中测试WebSocket连接
- 用法: 用浏览器打开该文件

### 2. 端口检测调试工具  
- 文件: `debug-ports.html`
- 功能: 专门调试3306、5000等问题端口
- 用法: 用浏览器打开该文件

### 3. 真正的Telnet工具（Node.js）
- 文件: `real-telnet.js` 
- 功能: 使用真正的TCP连接进行端口检测
- 用法: `node real-telnet.js localhost 22,80,443,3306,5000`

### 4. Telnet能力对比页面
- 文件: `telnet-comparison.html`
- 功能: 对比浏览器JS和Node.js的telnet能力
- 用法: 用浏览器打开该文件

## 关于前端JS的Telnet能力

**简答：前端JS没有直接的telnet方法**

由于浏览器安全限制：
- 无法直接建立TCP连接
- 无法进行原始套接字操作  
- 受同源策略限制

**解决方案：**
- 🌐 **浏览器环境**: 使用WebSocket、fetch等模拟telnet行为
- 🟢 **Node.js环境**: 使用net模块进行真正的TCP连接

详细对比请查看：`telnet-comparison.html`

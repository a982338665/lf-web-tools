# PWA 功能说明

## 概述

此Web应用现已支持PWA（Progressive Web App）功能，可以被Google Chrome浏览器安装到桌面，并支持自动更新。

## 主要功能

### 1. 桌面应用安装
- 用户访问网站时，Chrome会自动显示"安装应用"提示
- 支持从地址栏的安装图标进行安装
- 安装后应用将出现在桌面和应用菜单中
- 支持独立窗口模式，类似原生应用

### 2. 自动更新机制
- Service Worker会自动缓存应用资源
- 定期检查更新（每30分钟）
- 发现新版本时会显示更新通知
- 用户可选择立即更新或稍后更新
- 支持手动更新检查（Ctrl+U快捷键）

### 3. 离线功能
- 应用资源被缓存，支持离线访问
- 离线时显示相应状态提示
- 网络恢复时自动同步

### 4. 额外特性
- 主题色配置
- 应用快捷方式
- 推送通知支持（预留）
- 响应式设计

## 文件结构

```
gin-web-server/
├── static/
│   ├── manifest.json          # PWA 配置文件
│   ├── sw.js                  # Service Worker 脚本
│   ├── index.html            # 主页面（已添加PWA支持）
│   └── icons/                # 应用图标目录
│       ├── icon.svg          # 源图标文件
│       ├── icon-72x72.png    # 各尺寸PNG图标
│       ├── icon-96x96.png
│       ├── icon-128x128.png
│       ├── icon-144x144.png
│       ├── icon-152x152.png
│       ├── icon-192x192.png
│       ├── icon-384x384.png
│       └── icon-512x512.png
├── generate-icons.js         # 图标生成脚本
└── main.go                  # Go服务器（已添加PWA路由支持）
```

## 部署说明

### 1. 生成图标文件

**方式一：自动生成（推荐）**
```bash
# 安装canvas依赖
npm install canvas

# 运行图标生成脚本
node generate-icons.js
```

**方式二：手动创建**
如果自动生成失败，请手动创建以下尺寸的PNG图标文件：
- 72x72, 96x96, 128x128, 144x144, 152x152
- 192x192, 384x384, 512x512

图标应放置在 `static/icons/` 目录下，命名格式为 `icon-{size}x{size}.png`

### 2. 启动服务

```bash
cd gin-web-server
go run main.go
```

服务器启动后，访问 `http://localhost:8080` 即可体验PWA功能。

### 3. HTTPS部署（生产环境）

PWA功能在生产环境需要HTTPS支持。可以使用以下方式：

**方式一：使用nginx反向代理**
```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/your/cert.pem;
    ssl_certificate_key /path/to/your/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**方式二：修改Go服务器支持TLS**
```go
// 在main.go中替换最后一行
r.RunTLS(":8443", "cert.pem", "key.pem")
```

## 更新版本

要发布新版本并触发自动更新：

1. 修改 `static/sw.js` 中的 `CACHE_VERSION` 常量
2. 重启服务器
3. 用户下次访问时会收到更新提示

示例：
```javascript
const CACHE_VERSION = 'v1.0.1';  // 更新版本号
```

## 测试PWA功能

### 1. 安装提示测试
- 在Chrome中打开开发者工具
- 转到Application > Manifest标签页
- 检查manifest文件是否正确加载
- 点击"Add to homescreen"进行测试

### 2. Service Worker测试
- 在开发者工具的Application > Service Workers中查看注册状态
- 可以手动触发更新和缓存清理

### 3. 离线功能测试
- 在开发者工具的Network标签页中勾选"Offline"
- 刷新页面验证离线访问功能

## 常见问题

### Q: 为什么没有显示安装提示？
A: 确保满足以下条件：
- 使用HTTPS或localhost访问
- manifest.json文件正确配置
- 有有效的Service Worker
- 至少有一个192x192和512x512的图标

### Q: 如何强制显示安装提示？
A: 在开发者工具Console中执行：
```javascript
window.dispatchEvent(new Event('beforeinstallprompt'));
```

### Q: 更新通知不显示怎么办？
A: 检查：
- Service Worker是否正确注册
- 版本号是否已更新
- 浏览器控制台是否有错误信息

## 技术细节

- **PWA核心技术**：Web App Manifest + Service Worker
- **缓存策略**：缓存优先 + 网络更新
- **更新策略**：后台更新 + 用户确认
- **兼容性**：Chrome 67+, Edge 79+, Firefox 44+, Safari 11.1+

## 相关资源

- [PWA官方文档](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps)
- [Web App Manifest](https://developer.mozilla.org/en-US/docs/Web/Manifest)
- [Service Worker API](https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API)



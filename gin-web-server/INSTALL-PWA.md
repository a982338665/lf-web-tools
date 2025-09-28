# PWA 安装和测试指南

## 快速开始

您的Web应用现在已经支持PWA功能！以下是完整的设置和测试步骤：

### 1. 启动服务器

```bash
cd gin-web-server
go run main.go
```

服务器将在 `http://localhost:8080` 启动

### 2. 生成应用图标（必需）

由于PWA需要图标文件，请通过以下方式之一生成图标：

**方法A: 使用浏览器生成（推荐）**
1. 访问 `http://localhost:8080/static/icons/create-placeholder-icons.html`
2. 点击"生成图标"按钮
3. 将下载的所有PNG文件放入 `gin-web-server/static/icons/` 目录

**方法B: 使用现成的图标生成器**
1. 访问 `http://localhost:8080/static/create-basic-icons.html`
2. 点击"生成所有图标"然后"下载全部"
3. 将下载的文件放入 `gin-web-server/static/icons/` 目录

**方法C: 手动创建**
创建以下尺寸的PNG图标文件并放入 `static/icons/` 目录：
- `icon-72x72.png`
- `icon-96x96.png`
- `icon-128x128.png`
- `icon-144x144.png`
- `icon-152x152.png`
- `icon-192x192.png`
- `icon-384x384.png`
- `icon-512x512.png`

### 3. 测试PWA功能

#### 在Chrome浏览器中测试安装功能：

1. **访问应用**
   - 打开Chrome浏览器
   - 访问 `http://localhost:8080`

2. **检查PWA状态**
   - 按F12打开开发者工具
   - 转到"Application"标签页
   - 点击左侧的"Manifest"查看配置
   - 点击左侧的"Service Workers"查看注册状态

3. **安装应用**
   - 方式1：等待5秒后会自动显示安装提示
   - 方式2：点击地址栏右侧的安装图标（+号或下载图标）
   - 方式3：Chrome菜单 > 更多工具 > 创建快捷方式 > 勾选"在窗口中打开"

4. **验证安装**
   - 安装后应用会以独立窗口打开
   - 在桌面和开始菜单中可以找到应用快捷方式
   - 应用窗口不显示浏览器地址栏和工具栏

#### 测试自动更新功能：

1. **触发更新**
   - 修改 `static/sw.js` 中的 `CACHE_VERSION`（例如改为 'v1.0.1'）
   - 重启Go服务器

2. **验证更新通知**
   - 刷新或重新打开应用
   - 应该会在右上角显示"发现新版本"通知
   - 点击"立即更新"按钮进行更新

3. **手动检查更新**
   - 使用快捷键 `Ctrl+U`
   - 或等待30分钟自动检查

#### 测试离线功能：

1. **模拟离线**
   - 在开发者工具的"Network"标签页中勾选"Offline"
   - 或断开网络连接

2. **验证离线访问**
   - 刷新页面，应用仍可正常加载
   - 头部状态会显示"离线模式"
   - 会显示离线提示消息

### 4. 生产环境部署

对于生产环境，需要HTTPS支持：

#### 方法A: 使用nginx反向代理

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### 方法B: 直接在Go中启用TLS

修改 `main.go` 最后一行：
```go
r.RunTLS(":8443", "cert.pem", "key.pem")
```

### 5. 自定义配置

#### 修改应用信息
编辑 `static/manifest.json`:
```json
{
  "name": "你的应用名称",
  "short_name": "短名称",
  "description": "应用描述",
  "theme_color": "#your-color",
  "background_color": "#your-color"
}
```

#### 更新Service Worker版本
修改 `static/sw.js` 中的版本号：
```javascript
const CACHE_VERSION = 'v1.0.1';
```

### 6. 故障排除

#### PWA不显示安装提示？
- 确保使用HTTPS或localhost
- 检查manifest.json是否正确加载
- 确保Service Worker成功注册
- 确认图标文件存在且可访问

#### 更新不生效？
- 确认Service Worker版本号已更新
- 清除浏览器缓存
- 检查浏览器控制台错误信息

#### 离线功能无效？
- 检查Service Worker是否正确安装
- 确认资源已被缓存（开发者工具 > Application > Storage）

### 7. 兼容性

- **Chrome 67+**: 完全支持
- **Edge 79+**: 完全支持  
- **Firefox 44+**: 基础支持
- **Safari 11.1+**: 部分支持

### 8. 功能特性

✅ **已实现的功能**
- 桌面应用安装
- 自动缓存和更新
- 离线访问支持
- 安装/更新通知UI
- 独立窗口模式
- 应用图标和主题
- 快捷方式支持
- 在线/离线状态检测

🔄 **可扩展的功能**
- 推送通知
- 后台同步
- 地理位置API
- 摄像头/麦克风权限
- 文件系统访问

---

现在您的Web应用已经具备了完整的PWA功能，用户可以像使用原生应用一样安装和使用它！



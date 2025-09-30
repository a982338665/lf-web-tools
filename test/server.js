const http = require('http');
const WebSocket = require('ws');

console.log('🚀 启动测试服务器...\n');

// ========== WebSocket 服务器 (端口 3000) ==========
const wss = new WebSocket.Server({ port: 3000 });

wss.on('connection', function connection(ws, req) {
    console.log('📡 WebSocket客户端连接成功 (端口3000)');

    // 发送欢迎消息
    ws.send(JSON.stringify({
        type: 'welcome',
        message: '欢迎连接WebSocket服务器!',
        port: 3000,
        timestamp: new Date().toISOString()
    }));

    // 心跳机制,5s,改为500ms测试
    const heartbeat = setInterval(() => {
        if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
                type: 'ping',
                message: 'heartbeat',
                timestamp: new Date().toISOString()
            }));
        }
    }, 500);

    // 处理接收到的消息
    ws.on('message', function incoming(message) {
        console.log('📨 WebSocket收到消息:', message.toString());

        // 回显消息
        ws.send(JSON.stringify({
            type: 'echo',
            originalMessage: message.toString(),
            timestamp: new Date().toISOString()
        }));
    });

    // 连接关闭时清理
    ws.on('close', function close() {
        console.log('🔌 WebSocket客户端断开连接');
        clearInterval(heartbeat);
    });

    ws.on('error', function error(err) {
        console.error('❌ WebSocket错误:', err);
        clearInterval(heartbeat);
    });
});

console.log('✅ WebSocket服务器启动成功 - 端口: 3000');
console.log('   连接地址: ws://localhost:3000\n');

// ========== HTTP 服务器 1 (端口 8080) ==========
const server8080 = http.createServer((req, res) => {
    const url = req.url;
    const method = req.method;

    console.log(`📤 HTTP 8080 请求: ${method} ${url}`);

    // 设置CORS头
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

    if (method === 'OPTIONS') {
        res.writeHead(200);
        res.end();
        return;
    }

    if (url === '/' || url === '/index.html') {
        res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
        res.end(`
            <html>
            <head><title>测试服务器 - 端口8080</title></head>
            <body>
                <h1>🌐 HTTP服务器 - 端口8080</h1>
                <p>服务器运行正常！</p>
                <p>当前时间: ${new Date().toLocaleString()}</p>
                <p>请求路径: ${url}</p>
                <p>请求方法: ${method}</p>
                <hr>
                <p>可用接口:</p>
                <ul>
                    <li><a href="/api/status">GET /api/status</a> - 服务器状态</li>
                    <li><a href="/api/time">GET /api/time</a> - 当前时间</li>
                    <li><a href="/favicon.ico">GET /favicon.ico</a> - 网站图标</li>
                </ul>
            </body>
            </html>
        `);
    } else if (url === '/api/status') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
            status: 'ok',
            port: 8080,
            message: 'HTTP服务器8080运行正常',
            timestamp: new Date().toISOString(),
            uptime: process.uptime()
        }));
    } else if (url === '/api/time') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
            time: new Date().toISOString(),
            timestamp: Date.now(),
            port: 8080
        }));
    } else if (url === '/favicon.ico') {
        res.writeHead(200, { 'Content-Type': 'text/plain' });
        res.end('favicon');
    } else {
        res.writeHead(404, { 'Content-Type': 'text/html; charset=utf-8' });
        res.end(`
            <html>
            <body>
                <h1>404 - 页面未找到</h1>
                <p>请求路径: ${url}</p>
                <p><a href="/">返回首页</a></p>
            </body>
            </html>
        `);
    }
});

server8080.listen(8080, () => {
    console.log('✅ HTTP服务器启动成功 - 端口: 8080');
    console.log('   访问地址: http://localhost:8080\n');
});

// ========== HTTP 服务器 2 (端口 8000) ==========
const server8000 = http.createServer((req, res) => {
    const url = req.url;
    const method = req.method;

    console.log(`📤 HTTP 8000 请求: ${method} ${url}`);

    // 设置CORS头
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type, Authorization');

    if (method === 'OPTIONS') {
        res.writeHead(200);
        res.end();
        return;
    }

    if (url === '/' || url === '/index.html') {
        res.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8' });
        res.end(`
            <html>
            <head><title>测试服务器 - 端口8000</title></head>
            <body>
                <h1>🌐 HTTP服务器 - 端口8000</h1>
                <p>服务器运行正常！</p>
                <p>当前时间: ${new Date().toLocaleString()}</p>
                <p>请求路径: ${url}</p>
                <p>请求方法: ${method}</p>
                <hr>
                <p>可用接口:</p>
                <ul>
                    <li><a href="/api/info">GET /api/info</a> - 服务器信息</li>
                    <li><a href="/api/health">GET /api/health</a> - 健康检查</li>
                    <li><a href="/favicon.ico">GET /favicon.ico</a> - 网站图标</li>
                </ul>
            </body>
            </html>
        `);
    } else if (url === '/api/info') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
            server: 'Test HTTP Server',
            port: 8000,
            message: 'HTTP服务器8000运行正常',
            timestamp: new Date().toISOString(),
            nodeVersion: process.version,
            platform: process.platform
        }));
    } else if (url === '/api/health') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({
            status: 'healthy',
            port: 8000,
            uptime: process.uptime(),
            memory: process.memoryUsage(),
            timestamp: new Date().toISOString()
        }));
    } else if (url === '/favicon.ico') {
        res.writeHead(200, { 'Content-Type': 'text/plain' });
        res.end('favicon');
    } else {
        res.writeHead(404, { 'Content-Type': 'text/html; charset=utf-8' });
        res.end(`
            <html>
            <body>
                <h1>404 - 页面未找到</h1>
                <p>请求路径: ${url}</p>
                <p><a href="/">返回首页</a></p>
            </body>
            </html>
        `);
    }
});

server8000.listen(8000, () => {
    console.log('✅ HTTP服务器启动成功 - 端口: 8000');
    console.log('   访问地址: http://localhost:8000\n');
});

// ========== 错误处理 ==========
process.on('uncaughtException', (err) => {
    console.error('❌ 未捕获的异常:', err);
});

process.on('unhandledRejection', (reason, promise) => {
    console.error('❌ 未处理的Promise拒绝:', reason);
});

// ========== 优雅关闭 ==========
process.on('SIGINT', () => {
    console.log('\n🛑 收到终止信号，正在关闭服务器...');

    wss.close(() => {
        console.log('✅ WebSocket服务器已关闭');
    });

    server8080.close(() => {
        console.log('✅ HTTP服务器8080已关闭');
    });

    server8000.close(() => {
        console.log('✅ HTTP服务器8000已关闭');
        process.exit(0);
    });
});

console.log('🎯 所有服务器启动完成!');
console.log('📋 服务器列表:');
console.log('   • WebSocket: ws://localhost:3000');
console.log('   • HTTP 8080: http://localhost:8080');
console.log('   • HTTP 8000: http://localhost:8000');
console.log('\n💡 按 Ctrl+C 停止服务器');
console.log('📊 服务器日志:');
console.log('=====================================');







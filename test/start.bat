@echo off
echo 🚀 启动端口检测测试服务器...
echo.
echo 服务器配置:
echo   • WebSocket: ws://localhost:3000
echo   • HTTP 8080: http://localhost:8080  
echo   • HTTP 8000: http://localhost:8000
echo.
echo 💡 启动后请在另一个窗口打开网络信息检测器进行测试
echo    地址: http://localhost:8000/network_info_detector.html
echo.
echo 🛑 按 Ctrl+C 停止服务器
echo =====================================
echo.

node server.js

pause


@echo off
chcp 65001 >nul

REM 构建脚本，用于编译Windows和Linux版本的可执行文件

echo 开始构建gin-web-server...

REM 确保依赖已安装
echo 检查并安装依赖...
go mod tidy

REM 创建输出目录
mkdir build 2>nul

REM 构建Windows版本
echo 构建Windows版本...
go build -o build\app-windows-amd64.exe main.go

REM 构建Linux版本
echo 构建Linux版本...
set GOOS=linux
set GOARCH=amd64
go build -o build\app-linux-amd64 main.go
set GOOS=
set GOARCH=

REM 复制静态文件和模板到构建目录
echo 复制静态资源和模板文件...
xcopy /E /I /Y static build\static >nul
xcopy /E /I /Y templates build\templates >nul

REM 显示构建结果
echo 构建完成！
echo 输出文件：
dir build\

REM 保持窗口打开，按任意键继续
pause

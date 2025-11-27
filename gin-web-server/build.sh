#!/bin/bash

# 构建脚本，用于编译Windows和Linux版本的可执行文件

echo "开始构建gin-web-server..."

# 确保依赖已安装
echo "检查并安装依赖..."
go mod tidy

# 创建输出目录
mkdir -p build

# 构建Windows版本
echo "构建Windows版本..."
GOOS=windows GOARCH=amd64 go build -o build/app-windows-amd64.exe main.go

# 构建Linux版本
echo "构建Linux版本..."
GOOS=linux GOARCH=amd64 go build -o build/app-linux-amd64 main.go

# 复制静态文件和模板到构建目录
echo "复制静态资源和模板文件..."
cp -r static build/
cp -r templates build/

# 显示构建结果
echo "构建完成！"
echo "输出文件："
ls -la build/

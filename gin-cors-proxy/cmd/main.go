package main

import (
	"flag"
	"fmt"

	"github.com/lf-web-tools/gin-cors-proxy/middleware"
)

func main() {
	// 解析命令行参数
	port := flag.String("port", "8081", "Port to run the server on")
	flag.Parse()

	fmt.Printf("Starting CORS Proxy server on port %s...\n", *port)
	
	// 启动独立服务器
	middleware.StartStandalone(*port)
}

package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SetupWebSocketRoutes 设置WebSocket相关的路由
func SetupWebSocketRoutes(r *gin.Engine) {
	r.GET("/ws", handleWebSocket)
}

func handleWebSocket(c *gin.Context) {
	// 将HTTP连接升级为WebSocket连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade connection:", err)
		return
	}
	defer conn.Close()

	// 处理WebSocket消息
	for {
		// 读取消息
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// 打印收到的消息
		log.Printf("Received: %s", message)

		// 发送消息回客户端
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}
}

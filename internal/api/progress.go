package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ProgressMessage 定义发送给前端的进度消息格式
// 对应前端: this.currentFileName = data.fileName; this.progress = data.progress
type ProgressMessage struct {
	FileName string  `json:"fileName"`
	Progress float64 `json:"progress"`
}

// upgrader 用于将 HTTP 连接升级为 WebSocket 连接
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域，避免前端连接被拒绝
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 全局变量，用于保存当前的 WebSocket 连接
// 注意：这是一个简单实现，同一时间只能有一个活跃的进度推送连接
var (
	wsConn *websocket.Conn
	wsLock sync.Mutex
)

// ProgressWebSocketHandler 是路由处理函数，用于处理 /progress 请求
// 在 main.go 中注册路由时使用
func ProgressWebSocketHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to upgrade to WebSocket:", err)
		return
	}

	// 线程安全地保存连接
	wsLock.Lock()
	if wsConn != nil {
		wsConn.Close() // 如果有旧连接，先关闭
	}
	wsConn = conn
	wsLock.Unlock()

	log.Println("Frontend connected to WebSocket progress channel")

	// 保持连接活跃，监听关闭信号
	for {
		// 我们不需要读取前端发来的消息，但必须读，否则连接会断开
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket disconnected:", err)
			wsLock.Lock()
			wsConn = nil
			wsLock.Unlock()
			break
		}
	}
}

// SendProgress 是一个辅助函数，用于向前端发送进度
// 后续在下载逻辑中会调用这个函数
func SendProgress(fileName string, progress float64) {
	wsLock.Lock()
	defer wsLock.Unlock()

	if wsConn == nil {
		// 如果前端没有连接，直接忽略
		return
	}

	msg := ProgressMessage{
		FileName: fileName,
		Progress: progress,
	}

	msgBytes, _ := json.Marshal(msg)

	// 发送 JSON 消息
	err := wsConn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		log.Println("Error sending progress:", err)
		wsConn.Close()
		wsConn = nil
	}
}

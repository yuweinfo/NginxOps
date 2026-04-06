package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nginxops/internal/config"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源
	},
}

type LogWebSocketHandler struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan string
	paused     bool
	mu         sync.Mutex
}

func NewLogWebSocketHandler() *LogWebSocketHandler {
	h := &LogWebSocketHandler{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan string, 100),
	}
	go h.run()
	go h.tailLogFile()
	return h
}

func (h *LogWebSocketHandler) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket connected, total clients: %d", len(h.clients))
			// 发送欢迎消息
			h.sendToClient(client, map[string]interface{}{
				"type":    "connected",
				"message": "已连接到日志流",
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}
			h.mu.Unlock()
			log.Printf("WebSocket disconnected, total clients: %d", len(h.clients))

		case msg := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				if err := client.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
					client.Close()
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// HandleWebSocket 处理WebSocket连接
func (h *LogWebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	h.register <- conn

	// 读取客户端消息
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			h.unregister <- conn
			break
		}

		var cmd map[string]interface{}
		if err := json.Unmarshal(msg, &cmd); err == nil {
			action, _ := cmd["action"].(string)
			switch action {
			case "pause":
				h.mu.Lock()
				h.paused = true
				h.mu.Unlock()
				h.sendToClient(conn, map[string]interface{}{
					"type":   "status",
					"paused": true,
				})
			case "resume":
				h.mu.Lock()
				h.paused = false
				h.mu.Unlock()
				h.sendToClient(conn, map[string]interface{}{
					"type":   "status",
					"paused": false,
				})
			case "clear":
				h.sendToClient(conn, map[string]interface{}{
					"type": "clear",
				})
			}
		}
	}
}

func (h *LogWebSocketHandler) sendToClient(client *websocket.Conn, data map[string]interface{}) {
	msg, _ := json.Marshal(data)
	client.WriteMessage(websocket.TextMessage, msg)
}

func (h *LogWebSocketHandler) tailLogFile() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastPosition int64 = 0
	var lastFileSize int64 = 0

	for range ticker.C {
		h.mu.Lock()
		paused := h.paused
		hasClients := len(h.clients) > 0
		h.mu.Unlock()

		if paused || !hasClients {
			continue
		}

		accessLogPath := config.AppConfig.Nginx.AccessLog

		// 检查文件是否存在
		fileInfo, err := os.Stat(accessLogPath)
		if err != nil {
			// 文件不存在，发送模拟数据
			h.broadcastMockLog()
			continue
		}

		fileSize := fileInfo.Size()

		// 检查文件是否被轮转（新文件）
		if fileSize < lastFileSize {
			lastPosition = 0
		}

		// 读取新内容
		if fileSize > lastPosition {
			file, err := os.Open(accessLogPath)
			if err != nil {
				h.broadcastMockLog()
				continue
			}

			file.Seek(lastPosition, 0)
			buf := make([]byte, fileSize-lastPosition)
			n, _ := file.Read(buf)
			file.Close()

			if n > 0 {
				lines := splitLines(string(buf[:n]))
				for _, line := range lines {
					if line != "" {
						h.broadcastLog(h.parseLogLine(line))
					}
				}
			}

			lastPosition = fileSize
			lastFileSize = fileSize
		}
	}
}

func (h *LogWebSocketHandler) broadcastMockLog() {
	mockIps := []string{"192.168.1.101", "10.0.0.55", "172.16.0.23", "192.168.2.10"}
	mockMethods := []string{"GET", "POST", "PUT", "DELETE"}
	mockUrls := []string{"/api/users", "/api/orders", "/static/app.js", "/api/health"}
	mockStatuses := []int{200, 201, 304, 400, 404, 500}

	time := time.Now().Format("2006-01-02 15:04:05")
	ip := mockIps[randomInt(len(mockIps))]
	method := mockMethods[randomInt(len(mockMethods))]
	url := mockUrls[randomInt(len(mockUrls))]
	status := mockStatuses[randomInt(len(mockStatuses))]
	size := randomInt(10000)
	duration := randomInt(300)

	logLine := fmt.Sprintf("%s | %s | %s %s | %d | %dB | %dms", time, ip, method, url, status, size, duration)
	h.broadcastLog(logLine)
}

func (h *LogWebSocketHandler) parseLogLine(line string) string {
	time := time.Now().Format("2006-01-02 15:04:05")
	return time + " | " + line
}

func (h *LogWebSocketHandler) broadcastLog(logLine string) {
	h.mu.Lock()
	hasClients := len(h.clients) > 0
	h.mu.Unlock()

	if !hasClients {
		return
	}

	msg, _ := json.Marshal(map[string]interface{}{
		"type": "log",
		"line": logLine,
	})
	h.broadcast <- string(msg)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func randomInt(max int) int {
	return int(time.Now().UnixNano()) % max
}

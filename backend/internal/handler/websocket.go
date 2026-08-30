package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WSClient WebSocket 客户端连接
type WSClient struct {
	conn   *websocket.Conn
	send   chan []byte
	closed bool
	mu     sync.Mutex
}

// WSHub WebSocket 管理中心
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

var Hub = &WSHub{
	clients:    make(map[*WSClient]bool),
	broadcast:  make(chan []byte, 256),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
}

func init() {
	go Hub.run()
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WS] client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("[WS] client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					client.mu.Lock()
					if !client.closed {
						close(client.send)
						client.closed = true
					}
					client.mu.Unlock()
					h.mu.RUnlock()
					h.mu.Lock()
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastDeviceData 广播设备数据到所有 WebSocket 客户端
func BroadcastDeviceData(deviceSN string, data map[string]interface{}) {
	msg := map[string]interface{}{
		"type":      "device_data",
		"device_sn": deviceSN,
		"data":      data,
		"ts":        time.Now().Unix(),
	}
_bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	Hub.broadcast <- _bytes
}

// BroadcastDeviceStatus 广播设备状态变更
func BroadcastDeviceStatus(deviceSN string, online bool) {
	msg := map[string]interface{}{
		"type":      "device_status",
		"device_sn": deviceSN,
		"online":    online,
		"ts":        time.Now().Unix(),
	}
	_bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	Hub.broadcast <- _bytes
}

// BroadcastNotification 广播通知消息
func BroadcastNotification(userID uint, title, content, msgType string) {
	msg := map[string]interface{}{
		"type":    "notification",
		"user_id": userID,
		"title":   title,
		"content": content,
		"msg_type": msgType,
		"ts":      time.Now().Unix(),
	}
	_bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	Hub.broadcast <- _bytes
}

// HandleWebSocket WebSocket 连接端点
func HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WS] upgrade error: %v", err)
		return
	}

	client := &WSClient{
		conn: conn,
		send: make(chan []byte, 256),
	}

	Hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *WSClient) readPump() {
	defer func() {
		Hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		// 客户端消息暂时忽略（未来可支持订阅特定设备）
		_ = message
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

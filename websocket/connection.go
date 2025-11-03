package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"project/pkg/logger"
)

// Connection WebSocket连接封装
type Connection struct {
	ID            string                 // 连接ID
	UserID        uint                   // 用户ID
	Conn          *websocket.Conn        // WebSocket连接
	SendChan      chan []byte            // 发送消息通道
	Hub           *Hub                   // 连接池管理器
	IsAlive       bool                   // 连接是否存活
	LastHeartbeat time.Time              // 最后心跳时间
	mu            sync.RWMutex           // 读写锁
	closeChan     chan struct{}          // 关闭信号
	closeOnce     sync.Once              // 确保只关闭一次
	metadata      map[string]interface{} // 连接元数据
}

// Message WebSocket消息结构
type Message struct {
	Type      string          `json:"type"`      // 消息类型
	From      uint            `json:"from"`      // 发送者ID
	To        []uint          `json:"to"`        // 接收者ID列表
	Content   json.RawMessage `json:"content"`   // 消息内容
	Timestamp int64           `json:"timestamp"` // 时间戳
}

// MessageType 消息类型常量
const (
	MessageTypeText      = "text"      // 文本消息
	MessageTypeBroadcast = "broadcast" // 广播消息
	MessageTypeSystem    = "system"    // 系统消息
	MessageTypeHeartbeat = "heartbeat" // 心跳消息
	MessageTypeAck       = "ack"       // 确认消息
)

// 配置常量
const (
	// 写入超时
	writeWait = 10 * time.Second
	// 心跳超时
	pongWait = 60 * time.Second
	// 心跳发送间隔
	pingPeriod = (pongWait * 9) / 10
	// 最大消息大小
	maxMessageSize = 512 * 1024 // 512KB
	// 发送缓冲区大小
	sendChannelSize = 256
)

// NewConnection 创建新的WebSocket连接
func NewConnection(userID uint, conn *websocket.Conn, hub *Hub) *Connection {
	return &Connection{
		ID:            generateConnectionID(),
		UserID:        userID,
		Conn:          conn,
		SendChan:      make(chan []byte, sendChannelSize),
		Hub:           hub,
		IsAlive:       true,
		LastHeartbeat: time.Now(),
		closeChan:     make(chan struct{}),
		metadata:      make(map[string]interface{}),
	}
}

// Start 启动连接的读写协程
func (c *Connection) Start() {
	// 注册到连接池
	c.Hub.Register(c)

	// 启动读协程
	go c.readPump()
	// 启动写协程
	go c.writePump()
	// 启动心跳检测
	go c.heartbeatCheck()

	logger.Info("WebSocket连接已建立",
		zap.String("conn_id", c.ID),
		zap.Uint("user_id", c.UserID),
	)
}

// readPump 读取消息
func (c *Connection) readPump() {
	defer func() {
		c.Close()
	}()

	// 设置读取限制
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		c.updateHeartbeat()
		return nil
	})

	for {
		select {
		case <-c.closeChan:
			return
		default:
			// 读取消息
			messageType, data, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Error("WebSocket读取错误",
						zap.String("conn_id", c.ID),
						zap.Error(err),
					)
				}
				return
			}

			// 处理不同类型的消息
			switch messageType {
			case websocket.TextMessage:
				c.handleTextMessage(data)
			case websocket.BinaryMessage:
				c.handleBinaryMessage(data)
			}
		}
	}
}

// writePump 写入消息
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.SendChan:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// 通道已关闭
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 写入消息
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("WebSocket写入错误",
					zap.String("conn_id", c.ID),
					zap.Error(err),
				)
				return
			}

			// 批量发送缓冲区中的消息
			n := len(c.SendChan)
			for i := 0; i < n; i++ {
				if msg, ok := <-c.SendChan; ok {
					c.Conn.WriteMessage(websocket.TextMessage, msg)
				}
			}

		case <-ticker.C:
			// 发送心跳
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-c.closeChan:
			return
		}
	}
}

// handleTextMessage 处理文本消息
func (c *Connection) handleTextMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		logger.Error("消息解析失败",
			zap.String("conn_id", c.ID),
			zap.Error(err),
		)
		c.sendError("消息格式错误")
		return
	}

	// 设置发送者ID
	msg.From = c.UserID
	msg.Timestamp = time.Now().Unix()

	// 根据消息类型处理
	switch msg.Type {
	case MessageTypeHeartbeat:
		c.handleHeartbeat()
	case MessageTypeBroadcast:
		c.Hub.Broadcast(&msg)
	case MessageTypeText:
		c.Hub.Route(&msg)
	default:
		// 交给Hub的消息处理器处理
		if handler, ok := c.Hub.messageHandlers[msg.Type]; ok {
			handler(c, &msg)
		} else {
			c.sendError("未知的消息类型")
		}
	}
}

// handleBinaryMessage 处理二进制消息
func (c *Connection) handleBinaryMessage(data []byte) {
	// 处理二进制数据（如文件传输）
	logger.Debug("收到二进制消息",
		zap.String("conn_id", c.ID),
		zap.Int("size", len(data)),
	)
}

// handleHeartbeat 处理心跳消息
func (c *Connection) handleHeartbeat() {
	c.updateHeartbeat()
	// 回复心跳
	ack := Message{
		Type:      MessageTypeAck,
		Timestamp: time.Now().Unix(),
	}
	c.SendMessage(&ack)
}

// heartbeatCheck 心跳检测
func (c *Connection) heartbeatCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			lastHeartbeat := c.LastHeartbeat
			c.mu.RUnlock()

			if time.Since(lastHeartbeat) > 90*time.Second {
				logger.Warn("连接心跳超时",
					zap.String("conn_id", c.ID),
					zap.Uint("user_id", c.UserID),
				)
				c.Close()
				return
			}

		case <-c.closeChan:
			return
		}
	}
}

// SendMessage 发送消息
func (c *Connection) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// Send 发送原始数据
func (c *Connection) Send(data []byte) error {
	c.mu.RLock()
	if !c.IsAlive {
		c.mu.RUnlock()
		return ErrConnectionClosed
	}
	c.mu.RUnlock()

	select {
	case c.SendChan <- data:
		return nil
	case <-time.After(time.Second * 5):
		return ErrSendTimeout
	}
}

// SendJSON 发送JSON数据
func (c *Connection) SendJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.Send(data)
}

// sendError 发送错误消息
func (c *Connection) sendError(errMsg string) {
	msg := Message{
		Type:      MessageTypeSystem,
		Content:   json.RawMessage(`{"error":"` + errMsg + `"}`),
		Timestamp: time.Now().Unix(),
	}
	c.SendMessage(&msg)
}

// updateHeartbeat 更新心跳时间
func (c *Connection) updateHeartbeat() {
	c.mu.Lock()
	c.LastHeartbeat = time.Now()
	c.mu.Unlock()
}

// SetMetadata 设置元数据
func (c *Connection) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	c.metadata[key] = value
	c.mu.Unlock()
}

// GetMetadata 获取元数据
func (c *Connection) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.metadata[key]
	return value, exists
}

// Close 关闭连接
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.IsAlive = false
		c.mu.Unlock()

		// 从连接池中注销
		c.Hub.Unregister(c)

		// 关闭通道
		close(c.closeChan)
		close(c.SendChan)

		// 关闭WebSocket连接
		c.Conn.Close()

		logger.Info("WebSocket连接已关闭",
			zap.String("conn_id", c.ID),
			zap.Uint("user_id", c.UserID),
		)
	})
}

// generateConnectionID 生成连接ID
func generateConnectionID() string {
	return time.Now().Format("20060102150405") + "-" + generateRandomString(8)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

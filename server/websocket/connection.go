package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sunyuanling/server/pkg/logger"
	"go.uber.org/zap"
)

// Connection WebSocket连接封装
type Connection struct {
	ID            string                 // 连接ID
	UserID        uint                   // 用户ID
	Device        *DeviceInfo            // 设备信息
	Conn          *websocket.Conn        // WebSocket连接
	SendChan      chan []byte            // 发送消息通道
	Hub           *Hub                   // 连接池管理器
	IsAlive       bool                   // 连接是否存活
	ConnectedAt   time.Time              // 连接时间
	LastHeartbeat time.Time              // 最后心跳时间
	IP            string                 // 客户端IP
	mu            sync.RWMutex           // 读写锁
	closeChan     chan struct{}          // 关闭信号
	closeOnce     sync.Once              // 确保只关闭一次
	metadata      map[string]interface{} // 连接元数据
}

// NewConnection 创建新的WebSocket连接
func NewConnection(userID uint, conn *websocket.Conn, hub *Hub, device *DeviceInfo) *Connection {
	if device == nil {
		device = &DeviceInfo{
			DeviceID:   generateUUID(),
			DeviceType: DeviceTypeUnknown,
			Status:     DeviceStatusOnline,
		}
	}

	// 确保设备有ID
	if device.DeviceID == "" {
		device.DeviceID = generateUUID()
	}

	return &Connection{
		ID:            generateConnectionID(userID),
		UserID:        userID,
		Device:        device,
		Conn:          conn,
		SendChan:      make(chan []byte, SendChannelSize),
		Hub:           hub,
		IsAlive:       true,
		ConnectedAt:   time.Now(),
		LastHeartbeat: time.Now(),
		closeChan:     make(chan struct{}),
		metadata:      make(map[string]interface{}),
	}
}

// Start 启动连接的读写协程
func (c *Connection) Start() {
	// 注册到连接池
	c.Hub.Register(c)

	// 启动协程
	go c.readPump()
	go c.writePump()
	go c.heartbeatCheck()

	logger.Info("WebSocket连接已建立",
		zap.String("conn_id", c.ID),
		zap.Uint("user_id", c.UserID),
		zap.String("device_id", c.Device.DeviceID),
		zap.String("device_type", string(c.Device.DeviceType)),
	)
}

// readPump 读取消息
func (c *Connection) readPump() {
	defer func() {
		c.Close()
	}()

	c.Conn.SetReadLimit(MaxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		c.updateHeartbeat()
		return nil
	})

	for {
		select {
		case <-c.closeChan:
			return
		default:
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
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case message, ok := <-c.SendChan:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("WebSocket写入错误",
					zap.String("conn_id", c.ID),
					zap.Error(err),
				)
				return
			}

			// 批量发送
			n := len(c.SendChan)
			for i := 0; i < n; i++ {
				if msg, ok := <-c.SendChan; ok {
					if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						return
					}
				}
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
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

	// 设置发送者信息
	msg.From = &Sender{
		UserID:   c.UserID,
		ConnID:   c.ID,
		DeviceID: c.Device.DeviceID,
	}
	msg.Timestamp = time.Now().Unix()

	// 根据消息类型处理
	switch msg.Type {
	case MessageTypeHeartbeat:
		c.handleHeartbeat()
	default:
		c.Hub.RouteMessage(c, &msg)
	}
}

// handleBinaryMessage 处理二进制消息
func (c *Connection) handleBinaryMessage(data []byte) {
	logger.Debug("收到二进制消息",
		zap.String("conn_id", c.ID),
		zap.Int("size", len(data)),
	)
}

// handleHeartbeat 处理心跳
func (c *Connection) handleHeartbeat() {
	c.updateHeartbeat()
	ack := NewMessage(MessageTypeAck, map[string]interface{}{
		"type": "heartbeat_ack",
	})
	_ = c.SendMessage(ack)
}

// heartbeatCheck 心跳检测
func (c *Connection) heartbeatCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.RLock()
			lastHB := c.LastHeartbeat
			c.mu.RUnlock()

			if time.Since(lastHB) > 90*time.Second {
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
	case <-time.After(5 * time.Second):
		return ErrSendTimeout
	}
}

// SendMessage 发送消息对象
func (c *Connection) SendMessage(msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return c.Send(data)
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
	msg := NewMessage(MessageTypeSystem, map[string]interface{}{
		"error": errMsg,
	})
	_ = c.SendMessage(msg)
}

// updateHeartbeat 更新心跳时间
func (c *Connection) updateHeartbeat() {
	c.mu.Lock()
	c.LastHeartbeat = time.Now()
	c.mu.Unlock()
}

// SetDeviceStatus 设置设备状态
func (c *Connection) SetDeviceStatus(status DeviceStatus) {
	c.mu.Lock()
	c.Device.Status = status
	c.mu.Unlock()
}

// GetDeviceStatus 获取设备状态
func (c *Connection) GetDeviceStatus() DeviceStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Device.Status
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

// GetInfo 获取连接信息
func (c *Connection) GetInfo() *ConnectionInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &ConnectionInfo{
		ConnID:        c.ID,
		UserID:        c.UserID,
		Device:        c.Device,
		IP:            c.IP,
		ConnectedAt:   c.ConnectedAt,
		LastHeartbeat: c.LastHeartbeat,
		Status:        c.Device.Status,
	}
}

// Close 关闭连接
func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.IsAlive = false
		c.Device.Status = DeviceStatusOffline
		c.mu.Unlock()

		c.Hub.Unregister(c)

		close(c.closeChan)
		close(c.SendChan)

		_ = c.Conn.Close()

		logger.Info("WebSocket连接已关闭",
			zap.String("conn_id", c.ID),
			zap.Uint("user_id", c.UserID),
			zap.String("device_id", c.Device.DeviceID),
		)
	})
}

// generateConnectionID 生成连接ID
func generateConnectionID(userID uint) string {
	return fmt.Sprintf("conn-%d-%s", userID, uuid.New().String()[:8])
}

// generateUUID 生成UUID
func generateUUID() string {
	return uuid.New().String()
}

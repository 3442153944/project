package websocket

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/sunyuanling/server/pkg/logger"

	"go.uber.org/zap"
)

var (
	// 全局Hub实例
	globalHub *Hub
	once      sync.Once
)

// 错误定义
var (
	ErrConnectionClosed = errors.New("connection is closed")
	ErrSendTimeout      = errors.New("send timeout")
	ErrUserNotFound     = errors.New("user not found")
	ErrConnectionExists = errors.New("connection already exists")
)

// Hub WebSocket连接池管理器
type Hub struct {
	// 用户ID到连接的映射
	connections map[uint]*Connection
	// 连接ID到连接的映射
	connectionsByID map[string]*Connection
	// 用户分组
	groups map[string]map[uint]bool
	// 注册通道
	register chan *Connection
	// 注销通道
	unregister chan *Connection
	// 广播消息通道
	broadcast chan *Message
	// 读写锁
	mu sync.RWMutex
	// 消息处理器
	messageHandlers map[string]MessageHandler
	// 连接事件处理器
	onConnect    func(*Connection)
	onDisconnect func(*Connection)
	// 统计信息
	stats *Stats
}

// MessageHandler 消息处理器类型
type MessageHandler func(*Connection, *Message)

// Stats 统计信息
type Stats struct {
	TotalConnections   int64     `json:"total_connections"`
	ActiveConnections  int       `json:"active_connections"`
	TotalMessagesSent  int64     `json:"total_messages_sent"`
	TotalMessagesRecv  int64     `json:"total_messages_recv"`
	LastConnectionTime time.Time `json:"last_connection_time"`
	mu                 sync.RWMutex
}

// GetHub 获取全局Hub实例（单例模式）
func GetHub() *Hub {
	once.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
		logger.Info("WebSocket Hub已初始化")
	})
	return globalHub
}

// NewHub 创建新的Hub实例
func NewHub() *Hub {
	return &Hub{
		connections:     make(map[uint]*Connection),
		connectionsByID: make(map[string]*Connection),
		groups:          make(map[string]map[uint]bool),
		register:        make(chan *Connection, 256),
		unregister:      make(chan *Connection, 256),
		broadcast:       make(chan *Message, 512),
		messageHandlers: make(map[string]MessageHandler),
		stats:           &Stats{},
	}
}

// Run 运行Hub主循环
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case conn := <-h.register:
			// 注册新连接
			h.handleRegister(conn)

		case conn := <-h.unregister:
			// 注销连接
			h.handleUnregister(conn)

		case msg := <-h.broadcast:
			// 广播消息
			h.handleBroadcast(msg)

		case <-ticker.C:
			// 定期清理和统计
			h.cleanup()
		}
	}
}

// handleRegister 处理连接注册
func (h *Hub) handleRegister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查是否已存在相同用户的连接
	if oldConn, exists := h.connections[conn.UserID]; exists {
		// 关闭旧连接
		logger.Info("用户已存在连接，关闭旧连接",
			zap.Uint("user_id", conn.UserID),
			zap.String("old_conn_id", oldConn.ID),
			zap.String("new_conn_id", conn.ID),
		)
		go oldConn.Close()
	}

	// 注册新连接
	h.connections[conn.UserID] = conn
	h.connectionsByID[conn.ID] = conn

	// 更新统计
	h.stats.mu.Lock()
	h.stats.TotalConnections++
	h.stats.ActiveConnections = len(h.connections)
	h.stats.LastConnectionTime = time.Now()
	h.stats.mu.Unlock()

	// 触发连接事件
	if h.onConnect != nil {
		go h.onConnect(conn)
	}

	logger.Info("连接已注册",
		zap.String("conn_id", conn.ID),
		zap.Uint("user_id", conn.UserID),
		zap.Int("total_connections", len(h.connections)),
	)

	// 发送欢迎消息
	welcomeMsg := Message{
		Type: MessageTypeSystem,
		Content: jsonRawMessage(map[string]interface{}{
			"event":   "connected",
			"conn_id": conn.ID,
			"time":    time.Now().Unix(),
		}),
		Timestamp: time.Now().Unix(),
	}
	conn.SendMessage(&welcomeMsg)
}

// handleUnregister 处理连接注销
func (h *Hub) handleUnregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查连接是否存在
	if _, exists := h.connections[conn.UserID]; !exists {
		return
	}

	// 移除连接
	delete(h.connections, conn.UserID)
	delete(h.connectionsByID, conn.ID)

	// 从所有分组中移除
	for groupName, group := range h.groups {
		delete(group, conn.UserID)
		if len(group) == 0 {
			delete(h.groups, groupName)
		}
	}

	// 更新统计
	h.stats.mu.Lock()
	h.stats.ActiveConnections = len(h.connections)
	h.stats.mu.Unlock()

	// 触发断开连接事件
	if h.onDisconnect != nil {
		go h.onDisconnect(conn)
	}

	logger.Info("连接已注销",
		zap.String("conn_id", conn.ID),
		zap.Uint("user_id", conn.UserID),
		zap.Int("remaining_connections", len(h.connections)),
	)
}

// handleBroadcast 处理广播消息
func (h *Hub) handleBroadcast(msg *Message) {
	h.mu.RLock()
	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	// 并发发送消息
	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.SendMessage(msg); err != nil {
				logger.Error("广播消息失败",
					zap.String("conn_id", c.ID),
					zap.Error(err),
				)
			}
		}(conn)
	}
	wg.Wait()

	// 更新统计
	h.stats.mu.Lock()
	h.stats.TotalMessagesSent += int64(len(connections))
	h.stats.mu.Unlock()
}

// Register 注册连接
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister 注销连接
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

// Broadcast 广播消息给所有连接
func (h *Hub) Broadcast(msg *Message) {
	h.broadcast <- msg
}

// SendToUser 发送消息给指定用户
func (h *Hub) SendToUser(userID uint, msg *Message) error {
	h.mu.RLock()
	conn, exists := h.connections[userID]
	h.mu.RUnlock()

	if !exists {
		return ErrUserNotFound
	}

	return conn.SendMessage(msg)
}

// SendToUsers 发送消息给多个用户
func (h *Hub) SendToUsers(userIDs []uint, msg *Message) {
	h.mu.RLock()
	connections := make([]*Connection, 0, len(userIDs))
	for _, userID := range userIDs {
		if conn, exists := h.connections[userID]; exists {
			connections = append(connections, conn)
		}
	}
	h.mu.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.SendMessage(msg); err != nil {
				logger.Error("发送消息失败",
					zap.String("conn_id", c.ID),
					zap.Uint("user_id", c.UserID),
					zap.Error(err),
				)
			}
		}(conn)
	}
	wg.Wait()

	// 更新统计
	h.stats.mu.Lock()
	h.stats.TotalMessagesSent += int64(len(connections))
	h.stats.mu.Unlock()
}

// SendToGroup 发送消息给分组
func (h *Hub) SendToGroup(groupName string, msg *Message) {
	h.mu.RLock()
	group, exists := h.groups[groupName]
	if !exists {
		h.mu.RUnlock()
		return
	}

	// 获取分组中的所有连接
	connections := make([]*Connection, 0, len(group))
	for userID := range group {
		if conn, exists := h.connections[userID]; exists {
			connections = append(connections, conn)
		}
	}
	h.mu.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	for _, conn := range connections {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			c.SendMessage(msg)
		}(conn)
	}
	wg.Wait()
}

// Route 路由消息
func (h *Hub) Route(msg *Message) {
	if len(msg.To) == 0 {
		// 没有指定接收者，广播消息
		h.Broadcast(msg)
	} else {
		// 发送给指定用户
		h.SendToUsers(msg.To, msg)
	}
}

// AddToGroup 将用户添加到分组
func (h *Hub) AddToGroup(groupName string, userID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.groups[groupName] == nil {
		h.groups[groupName] = make(map[uint]bool)
	}
	h.groups[groupName][userID] = true

	logger.Debug("用户已添加到分组",
		zap.String("group", groupName),
		zap.Uint("user_id", userID),
	)
}

// RemoveFromGroup 从分组中移除用户
func (h *Hub) RemoveFromGroup(groupName string, userID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if group, exists := h.groups[groupName]; exists {
		delete(group, userID)
		if len(group) == 0 {
			delete(h.groups, groupName)
		}
	}
}

// GetConnection 获取用户连接
func (h *Hub) GetConnection(userID uint) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, exists := h.connections[userID]
	return conn, exists
}

// GetConnectionByID 通过连接ID获取连接
func (h *Hub) GetConnectionByID(connID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, exists := h.connectionsByID[connID]
	return conn, exists
}

// GetAllConnections 获取所有连接
func (h *Hub) GetAllConnections() []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}
	return connections
}

// GetOnlineUsers 获取在线用户列表
func (h *Hub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uint, 0, len(h.connections))
	for userID := range h.connections {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.connections[userID]
	return exists
}

// RegisterHandler 注册消息处理器
func (h *Hub) RegisterHandler(msgType string, handler MessageHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.messageHandlers[msgType] = handler
}

// SetConnectionHandler 设置连接事件处理器
func (h *Hub) SetConnectionHandler(onConnect, onDisconnect func(*Connection)) {
	h.onConnect = onConnect
	h.onDisconnect = onDisconnect
}

// GetStats 获取统计信息
func (h *Hub) GetStats() Stats {
	h.stats.mu.RLock()
	defer h.stats.mu.RUnlock()
	return *h.stats
}

// cleanup 定期清理
func (h *Hub) cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for userID, conn := range h.connections {
		conn.mu.RLock()
		lastHeartbeat := conn.LastHeartbeat
		isAlive := conn.IsAlive
		conn.mu.RUnlock()

		// 清理超时连接
		if !isAlive || now.Sub(lastHeartbeat) > 120*time.Second {
			logger.Warn("清理超时连接",
				zap.String("conn_id", conn.ID),
				zap.Uint("user_id", userID),
			)
			go conn.Close()
		}
	}

	// 清理空分组
	for name, group := range h.groups {
		if len(group) == 0 {
			delete(h.groups, name)
		}
	}
}

// Shutdown 关闭Hub
func (h *Hub) Shutdown() {
	logger.Info("正在关闭WebSocket Hub...")

	h.mu.Lock()
	connections := make([]*Connection, 0, len(h.connections))
	for _, conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.Unlock()

	// 关闭所有连接
	for _, conn := range connections {
		conn.Close()
	}

	logger.Info("WebSocket Hub已关闭")
}

// jsonRawMessage 将map转换为json.RawMessage
func jsonRawMessage(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

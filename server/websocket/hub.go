package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/sunyuanling/server/pkg/logger"
	"go.uber.org/zap"
)

var (
	globalHub *Hub
	hubOnce   sync.Once
)

// Hub WebSocket连接池管理器
type Hub struct {
	// 连接ID -> 连接 (主索引)
	connByID map[string]*Connection

	// 用户ID -> 连接ID集合 (一个用户可以有多个连接)
	connsByUser map[uint]map[string]bool

	// 设备ID -> 连接ID (一个设备同时只有一个连接)
	connByDevice map[string]string

	// 分组: 分组名 -> 用户ID集合
	groups map[string]map[uint]bool

	// 通道
	register   chan *Connection
	unregister chan *Connection
	broadcast  chan *Message

	// 消息处理器
	messageHandlers map[MessageType]MessageHandler

	// 事件回调
	onConnect    func(*Connection)
	onDisconnect func(*Connection)

	// 统计
	stats *Stats

	mu sync.RWMutex
}

// MessageHandler 消息处理器
type MessageHandler func(*Connection, *Message)

// Stats 统计信息
type Stats struct {
	TotalConnections   int64     `json:"total_connections"`
	ActiveConnections  int       `json:"active_connections"`
	ActiveUsers        int       `json:"active_users"`
	TotalMessagesSent  int64     `json:"total_messages_sent"`
	TotalMessagesRecv  int64     `json:"total_messages_recv"`
	LastConnectionTime time.Time `json:"last_connection_time"`
	mu                 sync.RWMutex
}

// GetHub 获取全局Hub实例
func GetHub() *Hub {
	hubOnce.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
		logger.Info("WebSocket Hub已初始化")
	})
	return globalHub
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		connByID:        make(map[string]*Connection),
		connsByUser:     make(map[uint]map[string]bool),
		connByDevice:    make(map[string]string),
		groups:          make(map[string]map[uint]bool),
		register:        make(chan *Connection, 256),
		unregister:      make(chan *Connection, 256),
		broadcast:       make(chan *Message, 512),
		messageHandlers: make(map[MessageType]MessageHandler),
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
			h.handleRegister(conn)

		case conn := <-h.unregister:
			h.handleUnregister(conn)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)

		case <-ticker.C:
			h.cleanup()
		}
	}
}

// handleRegister 处理连接注册
func (h *Hub) handleRegister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果同一设备已有连接，关闭旧连接
	if oldConnID, exists := h.connByDevice[conn.Device.DeviceID]; exists {
		if oldConn, ok := h.connByID[oldConnID]; ok {
			logger.Info("同一设备已存在连接，关闭旧连接",
				zap.String("device_id", conn.Device.DeviceID),
				zap.String("old_conn_id", oldConnID),
				zap.String("new_conn_id", conn.ID),
			)
			go oldConn.Close()
		}
	}

	// 注册连接
	h.connByID[conn.ID] = conn
	h.connByDevice[conn.Device.DeviceID] = conn.ID

	// 添加到用户连接集合
	if h.connsByUser[conn.UserID] == nil {
		h.connsByUser[conn.UserID] = make(map[string]bool)
	}
	h.connsByUser[conn.UserID][conn.ID] = true

	// 更新统计
	h.stats.mu.Lock()
	h.stats.TotalConnections++
	h.stats.ActiveConnections = len(h.connByID)
	h.stats.ActiveUsers = len(h.connsByUser)
	h.stats.LastConnectionTime = time.Now()
	h.stats.mu.Unlock()

	// 触发回调
	if h.onConnect != nil {
		go h.onConnect(conn)
	}

	logger.Info("连接已注册",
		zap.String("conn_id", conn.ID),
		zap.Uint("user_id", conn.UserID),
		zap.String("device_id", conn.Device.DeviceID),
		zap.Int("user_conn_count", len(h.connsByUser[conn.UserID])),
	)

	// 发送欢迎消息
	welcome := NewMessage(MessageTypeSystem, map[string]interface{}{
		"event":     "connected",
		"conn_id":   conn.ID,
		"device_id": conn.Device.DeviceID,
	})
	_ = conn.SendMessage(welcome)
}

// handleUnregister 处理连接注销
func (h *Hub) handleUnregister(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.connByID[conn.ID]; !exists {
		return
	}

	// 移除连接
	delete(h.connByID, conn.ID)
	delete(h.connByDevice, conn.Device.DeviceID)

	// 从用户连接集合中移除
	if userConns, exists := h.connsByUser[conn.UserID]; exists {
		delete(userConns, conn.ID)
		if len(userConns) == 0 {
			delete(h.connsByUser, conn.UserID)
			// 用户完全离线，从所有分组移除
			for groupName, group := range h.groups {
				delete(group, conn.UserID)
				if len(group) == 0 {
					delete(h.groups, groupName)
				}
			}
		}
	}

	// 更新统计
	h.stats.mu.Lock()
	h.stats.ActiveConnections = len(h.connByID)
	h.stats.ActiveUsers = len(h.connsByUser)
	h.stats.mu.Unlock()

	// 触发回调
	if h.onDisconnect != nil {
		go h.onDisconnect(conn)
	}

	logger.Info("连接已注销",
		zap.String("conn_id", conn.ID),
		zap.Uint("user_id", conn.UserID),
		zap.Int("remaining_conns", len(h.connByID)),
	)
}

// handleBroadcast 处理广播
func (h *Hub) handleBroadcast(msg *Message) {
	h.mu.RLock()
	conns := make([]*Connection, 0, len(h.connByID))
	for _, conn := range h.connByID {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	h.sendToConnections(conns, msg)
}

// Register 注册连接
func (h *Hub) Register(conn *Connection) {
	h.register <- conn
}

// Unregister 注销连接
func (h *Hub) Unregister(conn *Connection) {
	h.unregister <- conn
}

// RouteMessage 路由消息
func (h *Hub) RouteMessage(from *Connection, msg *Message) {
	// 先检查是否有注册的处理器
	if handler, ok := h.messageHandlers[msg.Type]; ok {
		handler(from, msg)
		return
	}

	// 根据目标类型路由
	if msg.Target == nil {
		return
	}

	switch msg.Target.Type {
	case TargetTypeUser:
		h.SendToUsers(msg.Target.UserIDs, msg)
	case TargetTypeConn:
		h.SendToConns(msg.Target.ConnIDs, msg)
	case TargetTypeDevice:
		h.SendToDevices(msg.Target.DeviceIDs, msg)
	case TargetTypeGroup:
		for _, group := range msg.Target.Groups {
			h.SendToGroup(group, msg)
		}
	case TargetTypeAll:
		h.Broadcast(msg)
	}
}

// Broadcast 广播消息
func (h *Hub) Broadcast(msg *Message) {
	h.broadcast <- msg
}

// SendToUsers 发送消息给指定用户（用户的所有连接都会收到）
func (h *Hub) SendToUsers(userIDs []uint, msg *Message) {
	h.mu.RLock()
	var conns []*Connection
	for _, userID := range userIDs {
		if connIDs, exists := h.connsByUser[userID]; exists {
			for connID := range connIDs {
				if conn, ok := h.connByID[connID]; ok {
					conns = append(conns, conn)
				}
			}
		}
	}
	h.mu.RUnlock()

	h.sendToConnections(conns, msg)
}

// SendToUser 发送消息给单个用户
func (h *Hub) SendToUser(userID uint, msg *Message) error {
	h.mu.RLock()
	connIDs, exists := h.connsByUser[userID]
	if !exists || len(connIDs) == 0 {
		h.mu.RUnlock()
		return ErrUserNotFound
	}

	var conns []*Connection
	for connID := range connIDs {
		if conn, ok := h.connByID[connID]; ok {
			conns = append(conns, conn)
		}
	}
	h.mu.RUnlock()

	h.sendToConnections(conns, msg)
	return nil
}

// SendToConns 发送消息给指定连接
func (h *Hub) SendToConns(connIDs []string, msg *Message) {
	h.mu.RLock()
	var conns []*Connection
	for _, connID := range connIDs {
		if conn, exists := h.connByID[connID]; exists {
			conns = append(conns, conn)
		}
	}
	h.mu.RUnlock()

	h.sendToConnections(conns, msg)
}

// SendToConn 发送消息给单个连接
func (h *Hub) SendToConn(connID string, msg *Message) error {
	h.mu.RLock()
	conn, exists := h.connByID[connID]
	h.mu.RUnlock()

	if !exists {
		return ErrConnNotFound
	}

	return conn.SendMessage(msg)
}

// SendToDevices 发送消息给指定设备
func (h *Hub) SendToDevices(deviceIDs []string, msg *Message) {
	h.mu.RLock()
	var conns []*Connection
	for _, deviceID := range deviceIDs {
		if connID, exists := h.connByDevice[deviceID]; exists {
			if conn, ok := h.connByID[connID]; ok {
				conns = append(conns, conn)
			}
		}
	}
	h.mu.RUnlock()

	h.sendToConnections(conns, msg)
}

// SendToDevice 发送消息给单个设备
func (h *Hub) SendToDevice(deviceID string, msg *Message) error {
	h.mu.RLock()
	connID, exists := h.connByDevice[deviceID]
	if !exists {
		h.mu.RUnlock()
		return ErrDeviceNotFound
	}
	conn, exists := h.connByID[connID]
	h.mu.RUnlock()

	if !exists {
		return ErrConnNotFound
	}

	return conn.SendMessage(msg)
}

// SendToGroup 发送消息给分组
func (h *Hub) SendToGroup(groupName string, msg *Message) {
	h.mu.RLock()
	group, exists := h.groups[groupName]
	if !exists {
		h.mu.RUnlock()
		return
	}
	userIDs := make([]uint, 0, len(group))
	for userID := range group {
		userIDs = append(userIDs, userID)
	}
	h.mu.RUnlock()

	h.SendToUsers(userIDs, msg)
}

// sendToConnections 发送消息给连接列表
func (h *Hub) sendToConnections(conns []*Connection, msg *Message) {
	if len(conns) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		go func(c *Connection) {
			defer wg.Done()
			if err := c.SendMessage(msg); err != nil {
				logger.Error("发送消息失败",
					zap.String("conn_id", c.ID),
					zap.Error(err),
				)
			}
		}(conn)
	}
	wg.Wait()

	h.stats.mu.Lock()
	h.stats.TotalMessagesSent += int64(len(conns))
	h.stats.mu.Unlock()
}

// AddToGroup 将用户添加到分组
func (h *Hub) AddToGroup(groupName string, userID uint) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.groups[groupName] == nil {
		h.groups[groupName] = make(map[uint]bool)
	}
	h.groups[groupName][userID] = true
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

// GetGroupUsers 获取分组中的用户
func (h *Hub) GetGroupUsers(groupName string) []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	group, exists := h.groups[groupName]
	if !exists {
		return nil
	}

	users := make([]uint, 0, len(group))
	for userID := range group {
		users = append(users, userID)
	}
	return users
}

// GetConnection 通过连接ID获取连接
func (h *Hub) GetConnection(connID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, exists := h.connByID[connID]
	return conn, exists
}

// GetConnectionByDevice 通过设备ID获取连接
func (h *Hub) GetConnectionByDevice(deviceID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connID, exists := h.connByDevice[deviceID]
	if !exists {
		return nil, false
	}
	conn, exists := h.connByID[connID]
	return conn, exists
}

// GetUserConnections 获取用户的所有连接
func (h *Hub) GetUserConnections(userID uint) []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	connIDs, exists := h.connsByUser[userID]
	if !exists {
		return nil
	}

	conns := make([]*Connection, 0, len(connIDs))
	for connID := range connIDs {
		if conn, ok := h.connByID[connID]; ok {
			conns = append(conns, conn)
		}
	}
	return conns
}

// GetUserConnectionsInfo 获取用户连接详情
func (h *Hub) GetUserConnectionsInfo(userID uint) *UserConnectionsInfo {
	conns := h.GetUserConnections(userID)
	if conns == nil {
		return nil
	}

	infos := make([]*ConnectionInfo, len(conns))
	for i, conn := range conns {
		infos[i] = conn.GetInfo()
	}

	return &UserConnectionsInfo{
		UserID:      userID,
		Connections: infos,
		TotalCount:  len(infos),
	}
}

// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.connsByUser[userID]
	return exists
}

// IsDeviceOnline 检查设备是否在线
func (h *Hub) IsDeviceOnline(deviceID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.connByDevice[deviceID]
	return exists
}

// GetOnlineUsers 获取在线用户列表
func (h *Hub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uint, 0, len(h.connsByUser))
	for userID := range h.connsByUser {
		users = append(users, userID)
	}
	return users
}

// GetAllConnections 获取所有连接
func (h *Hub) GetAllConnections() []*Connection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns := make([]*Connection, 0, len(h.connByID))
	for _, conn := range h.connByID {
		conns = append(conns, conn)
	}
	return conns
}

// RegisterHandler 注册消息处理器
func (h *Hub) RegisterHandler(msgType MessageType, handler MessageHandler) {
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
	for connID, conn := range h.connByID {
		conn.mu.RLock()
		lastHB := conn.LastHeartbeat
		alive := conn.IsAlive
		conn.mu.RUnlock()

		if !alive || now.Sub(lastHB) > 120*time.Second {
			logger.Warn("清理超时连接",
				zap.String("conn_id", connID),
				zap.Uint("user_id", conn.UserID),
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
	conns := make([]*Connection, 0, len(h.connByID))
	for _, conn := range h.connByID {
		conns = append(conns, conn)
	}
	h.mu.Unlock()

	for _, conn := range conns {
		conn.Close()
	}

	logger.Info("WebSocket Hub已关闭")
}

// jsonRawMessage 辅助函数
func jsonRawMessage(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

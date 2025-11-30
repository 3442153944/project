package websocket

import (
	"encoding/json"
	"time"

	"github.com/sunyuanling/server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InitWebSocket 初始化WebSocket系统
func InitWebSocket(db *gorm.DB) *Handler {
	hub := GetHub()

	// 设置连接事件处理器
	hub.SetConnectionHandler(
		// 连接成功
		func(conn *Connection) {
			msg := NewMessage(MessageTypeSystem, map[string]interface{}{
				"event":       "user_online",
				"user_id":     conn.UserID,
				"conn_id":     conn.ID,
				"device_id":   conn.Device.DeviceID,
				"device_type": conn.Device.DeviceType,
			})
			hub.Broadcast(msg)
		},
		// 断开连接
		func(conn *Connection) {
			// 检查用户是否还有其他连接
			remaining := len(hub.GetUserConnections(conn.UserID))

			msg := NewMessage(MessageTypeSystem, map[string]interface{}{
				"event":                 "user_offline",
				"user_id":               conn.UserID,
				"conn_id":               conn.ID,
				"device_id":             conn.Device.DeviceID,
				"remaining_connections": remaining,
			})
			hub.Broadcast(msg)
		},
	)

	// 注册默认消息处理器
	RegisterDefaultHandlers(hub)

	return NewHandler(db)
}

// RegisterDefaultHandlers 注册默认消息处理器
func RegisterDefaultHandlers(hub *Hub) {
	// 文件同步消息
	hub.RegisterHandler(MessageTypeFileSync, func(conn *Connection, msg *Message) {
		logger.Debug("收到文件同步消息",
			zap.String("conn_id", conn.ID),
			zap.Uint("user_id", conn.UserID),
		)
		// 路由给目标
		if msg.Target != nil {
			hub.RouteMessage(conn, msg)
		}
	})

	// 通知消息
	hub.RegisterHandler(MessageTypeNotify, func(conn *Connection, msg *Message) {
		if msg.Target != nil {
			hub.RouteMessage(conn, msg)
		}
	})

	// 文本消息
	hub.RegisterHandler(MessageTypeText, func(conn *Connection, msg *Message) {
		if msg.Target != nil {
			hub.RouteMessage(conn, msg)
		}
	})

	// 广播消息
	hub.RegisterHandler(MessageTypeBroadcast, func(conn *Connection, msg *Message) {
		hub.Broadcast(msg)
	})
}

// ========== 便捷函数 ==========

// Broadcast 广播消息给所有用户
func Broadcast(msgType MessageType, content interface{}) {
	msg := NewMessage(msgType, content)
	msg.Target = NewTargetAll()
	GetHub().Broadcast(msg)
}

// SendToUser 发送消息给用户（所有设备）
func SendToUser(userID uint, msgType MessageType, content interface{}) error {
	msg := NewMessage(msgType, content)
	msg.Target = NewTargetUser(userID)
	return GetHub().SendToUser(userID, msg)
}

// SendToUsers 发送消息给多个用户
func SendToUsers(userIDs []uint, msgType MessageType, content interface{}) {
	msg := NewMessage(msgType, content)
	msg.Target = NewTargetUser(userIDs...)
	GetHub().SendToUsers(userIDs, msg)
}

// SendToConn 发送消息给指定连接
func SendToConn(connID string, msgType MessageType, content interface{}) error {
	msg := NewMessage(msgType, content)
	msg.Target = NewTargetConn(connID)
	return GetHub().SendToConn(connID, msg)
}

// SendToDevice 发送消息给指定设备
func SendToDevice(deviceID string, msgType MessageType, content interface{}) error {
	msg := NewMessage(msgType, content)
	msg.Target = NewTargetDevice(deviceID)
	return GetHub().SendToDevice(deviceID, msg)
}

// SendToGroup 发送消息给分组
func SendToGroup(groupName string, msgType MessageType, content interface{}) {
	msg := NewMessage(msgType, content)
	msg.Target = NewTargetGroup(groupName)
	GetHub().SendToGroup(groupName, msg)
}

// NotifyUser 发送通知给用户
func NotifyUser(userID uint, title, message, level string) error {
	content := map[string]interface{}{
		"title":   title,
		"message": message,
		"level":   level,
		"time":    time.Now().Unix(),
	}
	return SendToUser(userID, MessageTypeNotify, content)
}

// NotifyDevice 发送通知给设备
func NotifyDevice(deviceID string, title, message, level string) error {
	content := map[string]interface{}{
		"title":   title,
		"message": message,
		"level":   level,
		"time":    time.Now().Unix(),
	}
	return SendToDevice(deviceID, MessageTypeNotify, content)
}

// NotifyAll 通知所有用户
func NotifyAll(title, message, level string) {
	content := map[string]interface{}{
		"title":   title,
		"message": message,
		"level":   level,
		"time":    time.Now().Unix(),
	}
	Broadcast(MessageTypeNotify, content)
}

// ========== 查询函数 ==========

// IsUserOnline 检查用户是否在线
func IsUserOnline(userID uint) bool {
	return GetHub().IsUserOnline(userID)
}

// IsDeviceOnline 检查设备是否在线
func IsDeviceOnline(deviceID string) bool {
	return GetHub().IsDeviceOnline(deviceID)
}

// GetOnlineUserCount 获取在线用户数量
func GetOnlineUserCount() int {
	return len(GetHub().GetOnlineUsers())
}

// GetOnlineUserIDs 获取所有在线用户ID
func GetOnlineUserIDs() []uint {
	return GetHub().GetOnlineUsers()
}

// GetUserConnections 获取用户的所有连接
func GetUserConnections(userID uint) []*ConnectionInfo {
	info := GetHub().GetUserConnectionsInfo(userID)
	if info == nil {
		return nil
	}
	return info.Connections
}

// GetUserDevices 获取用户的所有在线设备
func GetUserDevices(userID uint) []*DeviceInfo {
	conns := GetHub().GetUserConnections(userID)
	if conns == nil {
		return nil
	}

	devices := make([]*DeviceInfo, len(conns))
	for i, conn := range conns {
		devices[i] = conn.Device
	}
	return devices
}

// ========== 断开连接 ==========

// DisconnectUser 断开用户所有连接
func DisconnectUser(userID uint) {
	conns := GetHub().GetUserConnections(userID)
	for _, conn := range conns {
		conn.Close()
	}
}

// DisconnectDevice 断开指定设备
func DisconnectDevice(deviceID string) {
	if conn, exists := GetHub().GetConnectionByDevice(deviceID); exists {
		conn.Close()
	}
}

// DisconnectConn 断开指定连接
func DisconnectConn(connID string) {
	if conn, exists := GetHub().GetConnection(connID); exists {
		conn.Close()
	}
}

// ========== 分组管理 ==========

// AddUserToGroup 将用户添加到分组
func AddUserToGroup(groupName string, userID uint) {
	GetHub().AddToGroup(groupName, userID)
}

// RemoveUserFromGroup 从分组移除用户
func RemoveUserFromGroup(groupName string, userID uint) {
	GetHub().RemoveFromGroup(groupName, userID)
}

// GetGroupUsers 获取分组中的用户
func GetGroupUsers(groupName string) []uint {
	return GetHub().GetGroupUsers(groupName)
}

// ========== 辅助函数 ==========

// jsonRawMessageHelper 将interface{}转换为json.RawMessage
func jsonRawMessageHelper(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

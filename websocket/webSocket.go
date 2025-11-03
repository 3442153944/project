package websocket

import (
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"project/pkg/logger"
)

// 全局函数，便于使用

// InitWebSocket 初始化WebSocket系统
func InitWebSocket(db *gorm.DB) *Handler {
	// 获取全局Hub实例
	hub := GetHub()

	// 设置连接事件处理器
	hub.SetConnectionHandler(
		// 连接成功
		func(conn *Connection) {
			// 通知其他用户有新用户上线
			msg := Message{
				Type: MessageTypeSystem,
				Content: jsonRawMessageMain(map[string]interface{}{
					"event":     "user_online",
					"user_id":   conn.UserID,
					"conn_id":   conn.ID,
					"timestamp": time.Now().Unix(),
				}),
				Timestamp: time.Now().Unix(),
			}
			hub.Broadcast(&msg)
		},
		// 断开连接
		func(conn *Connection) {
			// 通知其他用户有用户下线
			msg := Message{
				Type: MessageTypeSystem,
				Content: jsonRawMessageMain(map[string]interface{}{
					"event":     "user_offline",
					"user_id":   conn.UserID,
					"conn_id":   conn.ID,
					"timestamp": time.Now().Unix(),
				}),
				Timestamp: time.Now().Unix(),
			}
			hub.Broadcast(&msg)
		},
	)

	// 注册默认消息处理器
	RegisterDefaultHandlers(hub)

	// 返回HTTP处理器
	return NewHandler(db)
}

// RegisterDefaultHandlers 注册默认消息处理器
func RegisterDefaultHandlers(hub *Hub) {
	// 文件同步消息处理器
	hub.RegisterHandler("file_sync", func(conn *Connection, msg *Message) {
		// TODO: 实现文件同步逻辑
		logger.Debug("收到文件同步消息",
			zap.String("conn_id", conn.ID),
			zap.Uint("user_id", conn.UserID),
		)
	})

	// 通知消息处理器
	hub.RegisterHandler("notification", func(conn *Connection, msg *Message) {
		// 转发通知给指定用户
		if len(msg.To) > 0 {
			hub.SendToUsers(msg.To, msg)
		}
	})

	// 状态更新处理器
	hub.RegisterHandler("status_update", func(conn *Connection, msg *Message) {
		// 更新用户状态
		var status struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(msg.Content, &status); err == nil {
			conn.SetMetadata("status", status.Status)
			// 广播状态更新
			hub.Broadcast(msg)
		}
	})
}

// 便捷函数

// BroadcastMessage 广播消息给所有在线用户
func BroadcastMessage(msgType string, content interface{}) error {
	hub := GetHub()
	msg := &Message{
		Type:      msgType,
		Content:   jsonRawMessageMain(content),
		Timestamp: time.Now().Unix(),
	}
	hub.Broadcast(msg)
	return nil
}

// SendToUser 发送消息给指定用户
func SendToUser(userID uint, msgType string, content interface{}) error {
	hub := GetHub()
	msg := &Message{
		Type:      msgType,
		To:        []uint{userID},
		Content:   jsonRawMessageMain(content),
		Timestamp: time.Now().Unix(),
	}
	return hub.SendToUser(userID, msg)
}

// SendToUsers 发送消息给多个用户
func SendToUsers(userIDs []uint, msgType string, content interface{}) {
	hub := GetHub()
	msg := &Message{
		Type:      msgType,
		To:        userIDs,
		Content:   jsonRawMessageMain(content),
		Timestamp: time.Now().Unix(),
	}
	hub.SendToUsers(userIDs, msg)
}

// NotifyUser 发送通知给用户
func NotifyUser(userID uint, title, message string, level string) error {
	content := map[string]interface{}{
		"title":   title,
		"message": message,
		"level":   level, // info, warning, error, success
		"time":    time.Now().Unix(),
	}
	return SendToUser(userID, "notification", content)
}

// NotifyAllUsers 通知所有用户
func NotifyAllUsers(title, message string, level string) {
	content := map[string]interface{}{
		"title":   title,
		"message": message,
		"level":   level,
		"time":    time.Now().Unix(),
	}
	BroadcastMessage("notification", content)
}

// IsUserOnline 检查用户是否在线
func IsUserOnline(userID uint) bool {
	return GetHub().IsUserOnline(userID)
}

// GetOnlineUserCount 获取在线用户数量
func GetOnlineUserCount() int {
	return len(GetHub().GetOnlineUsers())
}

// GetOnlineUserIDs 获取所有在线用户ID
func GetOnlineUserIDs() []uint {
	return GetHub().GetOnlineUsers()
}

// DisconnectUser 断开指定用户的连接
func DisconnectUser(userID uint) {
	if conn, exists := GetHub().GetConnection(userID); exists {
		conn.Close()
	}
}

// jsonRawMessageMain 将interface{}转换为json.RawMessage
func jsonRawMessageMain(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return json.RawMessage(data)
}

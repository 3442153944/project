package websocket

import (
	"encoding/json"
	"errors"
	"time"
)

// 错误定义
var (
	ErrConnectionClosed = errors.New("connection is closed")
	ErrSendTimeout      = errors.New("send timeout")
	ErrUserNotFound     = errors.New("user not found")
	ErrConnNotFound     = errors.New("connection not found")
	ErrDeviceNotFound   = errors.New("device not found")
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeText      MessageType = "text"
	MessageTypeBroadcast MessageType = "broadcast"
	MessageTypeSystem    MessageType = "system"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeAck       MessageType = "ack"
	MessageTypeFileSync  MessageType = "file_sync"
	MessageTypeNotify    MessageType = "notification"
)

// TargetType 消息目标类型
type TargetType string

const (
	TargetTypeUser   TargetType = "user"   // 按用户ID发送（该用户所有设备都收到）
	TargetTypeConn   TargetType = "conn"   // 按连接ID发送（仅指定连接收到）
	TargetTypeDevice TargetType = "device" // 按设备ID发送（该设备的连接收到）
	TargetTypeGroup  TargetType = "group"  // 按分组发送
	TargetTypeAll    TargetType = "all"    // 广播
)

// DeviceType 设备类型
type DeviceType string

const (
	DeviceTypeUnknown DeviceType = "unknown"
	DeviceTypeWeb     DeviceType = "web"
	DeviceTypeAndroid DeviceType = "android"
	DeviceTypeIOS     DeviceType = "ios"
	DeviceTypeDesktop DeviceType = "desktop"
	DeviceTypeServer  DeviceType = "server" // 服务端客户端
)

// DeviceStatus 设备状态
type DeviceStatus string

const (
	DeviceStatusOnline  DeviceStatus = "online"
	DeviceStatusAway    DeviceStatus = "away"
	DeviceStatusBusy    DeviceStatus = "busy"
	DeviceStatusOffline DeviceStatus = "offline"
)

// Message WebSocket消息结构
type Message struct {
	ID        string          `json:"id,omitempty"`    // 消息ID
	Type      MessageType     `json:"type"`            // 消息类型
	From      *Sender         `json:"from,omitempty"`  // 发送者信息
	Target    *Target         `json:"target"`          // 目标信息
	Content   json.RawMessage `json:"content"`         // 消息内容
	Timestamp int64           `json:"timestamp"`       // 时间戳
	Extra     json.RawMessage `json:"extra,omitempty"` // 扩展字段
}

// Sender 发送者信息
type Sender struct {
	UserID   uint   `json:"user_id"`
	ConnID   string `json:"conn_id,omitempty"`
	DeviceID string `json:"device_id,omitempty"`
}

// Target 消息目标
type Target struct {
	Type      TargetType `json:"type"`                 // 目标类型
	UserIDs   []uint     `json:"user_ids,omitempty"`   // 用户ID列表
	ConnIDs   []string   `json:"conn_ids,omitempty"`   // 连接ID列表
	DeviceIDs []string   `json:"device_ids,omitempty"` // 设备ID列表
	Groups    []string   `json:"groups,omitempty"`     // 分组名称列表
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	DeviceID   string                 `json:"device_id"`       // 设备唯一标识
	DeviceType DeviceType             `json:"device_type"`     // 设备类型
	DeviceName string                 `json:"device_name"`     // 设备名称
	Status     DeviceStatus           `json:"status"`          // 设备状态
	Platform   string                 `json:"platform"`        // 平台信息 (如 Android 14, Windows 11)
	AppVersion string                 `json:"app_version"`     // 应用版本
	PushToken  string                 `json:"push_token"`      // 推送Token（用于离线推送）
	Extra      map[string]interface{} `json:"extra,omitempty"` // 扩展信息
}

// ConnectionInfo 连接完整信息（用于外部查询）
type ConnectionInfo struct {
	ConnID        string       `json:"conn_id"`
	UserID        uint         `json:"user_id"`
	Device        *DeviceInfo  `json:"device"`
	IP            string       `json:"ip"`
	ConnectedAt   time.Time    `json:"connected_at"`
	LastHeartbeat time.Time    `json:"last_heartbeat"`
	Status        DeviceStatus `json:"status"`
}

// UserConnectionsInfo 用户所有连接信息
type UserConnectionsInfo struct {
	UserID      uint              `json:"user_id"`
	Connections []*ConnectionInfo `json:"connections"`
	TotalCount  int               `json:"total_count"`
}

// 配置常量
const (
	WriteWait       = 10 * time.Second
	PongWait        = 60 * time.Second
	PingPeriod      = (PongWait * 9) / 10
	MaxMessageSize  = 512 * 1024 // 512KB
	SendChannelSize = 256
)

// NewMessage 创建新消息
func NewMessage(msgType MessageType, content interface{}) *Message {
	data, _ := json.Marshal(content)
	return &Message{
		ID:        generateMessageID(),
		Type:      msgType,
		Content:   data,
		Timestamp: time.Now().Unix(),
	}
}

// NewTargetUser 创建用户目标
func NewTargetUser(userIDs ...uint) *Target {
	return &Target{
		Type:    TargetTypeUser,
		UserIDs: userIDs,
	}
}

// NewTargetConn 创建连接目标
func NewTargetConn(connIDs ...string) *Target {
	return &Target{
		Type:    TargetTypeConn,
		ConnIDs: connIDs,
	}
}

// NewTargetDevice 创建设备目标
func NewTargetDevice(deviceIDs ...string) *Target {
	return &Target{
		Type:      TargetTypeDevice,
		DeviceIDs: deviceIDs,
	}
}

// NewTargetGroup 创建分组目标
func NewTargetGroup(groups ...string) *Target {
	return &Target{
		Type:   TargetTypeGroup,
		Groups: groups,
	}
}

// NewTargetAll 创建广播目标
func NewTargetAll() *Target {
	return &Target{
		Type: TargetTypeAll,
	}
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return generateUUID()
}

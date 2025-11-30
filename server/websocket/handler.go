package websocket

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sunyuanling/server/internal/model"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
	token "github.com/sunyuanling/server/pkg/tokn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler WebSocket处理器
type Handler struct {
	db       *gorm.DB
	upgrader websocket.Upgrader
	hub      *Hub
}

// NewHandler 创建处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		hub: GetHub(),
	}
}

// ConnectRequest 连接请求参数
type ConnectRequest struct {
	DeviceID   string `form:"device_id"`
	DeviceType string `form:"device_type"`
	DeviceName string `form:"device_name"`
	Platform   string `form:"platform"`
	AppVersion string `form:"app_version"`
}

// Connect WebSocket连接入口
func (h *Handler) Connect(c *gin.Context) {
	// 获取用户信息
	userInfoAny, exists := c.Get("UserInfo")
	if !exists || userInfoAny == nil {
		logger.Warn("WebSocket处理失败", zap.String("error", "未授权"))
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	payload, ok := userInfoAny.(*token.TokenPayload)
	if !ok {
		logger.Warn("用户信息类型错误")
		response.Error(c, http.StatusInternalServerError, "用户信息类型错误")
		return
	}

	userID := uint(payload.UserID)

	// 验证用户
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 解析设备信息
	var req ConnectRequest
	_ = c.ShouldBindQuery(&req)

	deviceInfo := &DeviceInfo{
		DeviceID:   req.DeviceID,
		DeviceType: parseDeviceType(req.DeviceType),
		DeviceName: req.DeviceName,
		Platform:   req.Platform,
		AppVersion: req.AppVersion,
		Status:     DeviceStatusOnline,
	}

	// 升级连接
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket升级失败", zap.Uint("user_id", userID), zap.Error(err))
		return
	}

	// 创建连接
	connection := NewConnection(userID, conn, h.hub, deviceInfo)
	connection.IP = c.ClientIP()
	connection.SetMetadata("username", user.Username)
	connection.SetMetadata("role", user.Role)
	connection.Start()
}

// parseDeviceType 解析设备类型
func parseDeviceType(s string) DeviceType {
	switch s {
	case "web":
		return DeviceTypeWeb
	case "android":
		return DeviceTypeAndroid
	case "ios":
		return DeviceTypeIOS
	case "desktop":
		return DeviceTypeDesktop
	case "server":
		return DeviceTypeServer
	default:
		return DeviceTypeUnknown
	}
}

// GetOnlineUsers 获取在线用户
func (h *Handler) GetOnlineUsers(c *gin.Context) {
	users := h.hub.GetOnlineUsers()

	var userList []map[string]interface{}
	for _, userID := range users {
		var user model.User
		if err := h.db.First(&user, userID).Error; err == nil {
			info := h.hub.GetUserConnectionsInfo(userID)
			userList = append(userList, map[string]interface{}{
				"id":          user.ID,
				"username":    user.Username,
				"avatar":      user.Avatar,
				"role":        user.Role,
				"connections": info.Connections,
				"conn_count":  info.TotalCount,
			})
		}
	}

	response.Success(c, gin.H{
		"total": len(userList),
		"users": userList,
	})
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	TargetType string          `json:"target_type" binding:"required"` // user/conn/device/group
	UserIDs    []uint          `json:"user_ids,omitempty"`
	ConnIDs    []string        `json:"conn_ids,omitempty"`
	DeviceIDs  []string        `json:"device_ids,omitempty"`
	Groups     []string        `json:"groups,omitempty"`
	Type       string          `json:"type" binding:"required"`
	Content    json.RawMessage `json:"content" binding:"required"`
}

// SendMessage 发送消息
func (h *Handler) SendMessage(c *gin.Context) {
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	senderIDVal, _ := c.Get("user_id")
	senderID := uint(senderIDVal.(float64))

	msg := &Message{
		ID:   generateUUID(),
		Type: MessageType(req.Type),
		From: &Sender{
			UserID: senderID,
		},
		Content:   req.Content,
		Timestamp: time.Now().Unix(),
	}

	var targetDesc string

	switch req.TargetType {
	case "user":
		msg.Target = NewTargetUser(req.UserIDs...)
		h.hub.SendToUsers(req.UserIDs, msg)
		targetDesc = "users"
	case "conn":
		msg.Target = NewTargetConn(req.ConnIDs...)
		h.hub.SendToConns(req.ConnIDs, msg)
		targetDesc = "connections"
	case "device":
		msg.Target = NewTargetDevice(req.DeviceIDs...)
		h.hub.SendToDevices(req.DeviceIDs, msg)
		targetDesc = "devices"
	case "group":
		msg.Target = NewTargetGroup(req.Groups...)
		for _, group := range req.Groups {
			h.hub.SendToGroup(group, msg)
		}
		targetDesc = "groups"
	default:
		response.Error(c, http.StatusBadRequest, "无效的目标类型")
		return
	}

	response.Success(c, gin.H{
		"message":     "消息已发送",
		"target_type": targetDesc,
	})
}

// BroadcastMessage 广播消息
func (h *Handler) BroadcastMessage(c *gin.Context) {
	var req struct {
		Type    string          `json:"type" binding:"required"`
		Content json.RawMessage `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	senderIDVal, _ := c.Get("user_id")
	senderID := uint(senderIDVal.(float64))

	msg := &Message{
		ID:   generateUUID(),
		Type: MessageType(req.Type),
		From: &Sender{
			UserID: senderID,
		},
		Target:    NewTargetAll(),
		Content:   req.Content,
		Timestamp: time.Now().Unix(),
	}

	h.hub.Broadcast(msg)

	response.Success(c, gin.H{
		"message": "消息已广播",
		"online":  len(h.hub.GetOnlineUsers()),
	})
}

// GetUserConnections 获取用户连接信息
func (h *Handler) GetUserConnections(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	info := h.hub.GetUserConnectionsInfo(uint(userID))
	if info == nil {
		response.Error(c, http.StatusNotFound, "用户不在线")
		return
	}

	response.Success(c, info)
}

// DisconnectConn 断开指定连接
func (h *Handler) DisconnectConn(c *gin.Context) {
	connID := c.Param("conn_id")

	conn, exists := h.hub.GetConnection(connID)
	if !exists {
		response.Error(c, http.StatusNotFound, "连接不存在")
		return
	}

	conn.Close()
	response.Success(c, gin.H{
		"message": "连接已断开",
		"conn_id": connID,
	})
}

// DisconnectUser 断开用户所有连接
func (h *Handler) DisconnectUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	conns := h.hub.GetUserConnections(uint(userID))
	if len(conns) == 0 {
		response.Error(c, http.StatusNotFound, "用户不在线")
		return
	}

	for _, conn := range conns {
		conn.Close()
	}

	response.Success(c, gin.H{
		"message":    "用户所有连接已断开",
		"user_id":    userID,
		"conn_count": len(conns),
	})
}

// DisconnectDevice 断开指定设备
func (h *Handler) DisconnectDevice(c *gin.Context) {
	deviceID := c.Param("device_id")

	conn, exists := h.hub.GetConnectionByDevice(deviceID)
	if !exists {
		response.Error(c, http.StatusNotFound, "设备不在线")
		return
	}

	conn.Close()
	response.Success(c, gin.H{
		"message":   "设备已断开",
		"device_id": deviceID,
	})
}

// GetStats 获取统计信息
func (h *Handler) GetStats(c *gin.Context) {
	stats := h.hub.GetStats()
	response.Success(c, gin.H{
		"stats": stats,
	})
}

// CreateGroup 创建分组
func (h *Handler) CreateGroup(c *gin.Context) {
	var req struct {
		GroupName string `json:"group_name" binding:"required"`
		UserIDs   []uint `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	for _, userID := range req.UserIDs {
		h.hub.AddToGroup(req.GroupName, userID)
	}

	response.Success(c, gin.H{
		"message":    "分组创建成功",
		"group_name": req.GroupName,
		"users":      req.UserIDs,
	})
}

// SendToGroup 发送消息给分组
func (h *Handler) SendToGroup(c *gin.Context) {
	var req struct {
		GroupName string          `json:"group_name" binding:"required"`
		Type      string          `json:"type" binding:"required"`
		Content   json.RawMessage `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	senderIDVal, _ := c.Get("user_id")
	senderID := uint(senderIDVal.(float64))

	msg := &Message{
		ID:   generateUUID(),
		Type: MessageType(req.Type),
		From: &Sender{
			UserID: senderID,
		},
		Target:    NewTargetGroup(req.GroupName),
		Content:   req.Content,
		Timestamp: time.Now().Unix(),
	}

	h.hub.SendToGroup(req.GroupName, msg)

	response.Success(c, gin.H{
		"message":    "消息已发送到分组",
		"group_name": req.GroupName,
	})
}

// GetGroupUsers 获取分组用户
func (h *Handler) GetGroupUsers(c *gin.Context) {
	groupName := c.Param("name")
	users := h.hub.GetGroupUsers(groupName)

	response.Success(c, gin.H{
		"group_name": groupName,
		"users":      users,
		"count":      len(users),
	})
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	ws := router.Group("/ws")
	{
		// WebSocket连接
		ws.GET("/connect", h.Connect)

		// 在线状态
		ws.GET("/online", h.GetOnlineUsers)
		ws.GET("/user/:id/connections", h.GetUserConnections)
		ws.GET("/stats", h.GetStats)

		// 消息发送
		ws.POST("/send", h.SendMessage)
		ws.POST("/broadcast", h.BroadcastMessage)

		// 断开连接
		ws.DELETE("/conn/:conn_id", h.DisconnectConn)
		ws.DELETE("/user/:id", h.DisconnectUser)
		ws.DELETE("/device/:device_id", h.DisconnectDevice)

		// 分组管理
		ws.POST("/group", h.CreateGroup)
		ws.POST("/group/send", h.SendToGroup)
		ws.GET("/group/:name/users", h.GetGroupUsers)
	}
}

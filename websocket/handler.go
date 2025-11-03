package websocket

import (
	"encoding/json"
	"net/http"
	token "project/pkg/tokn"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"project/internal/model"
	"project/pkg/logger"
	"project/pkg/response"
)

// Handler WebSocket处理器
type Handler struct {
	db       *gorm.DB
	upgrader websocket.Upgrader
	hub      *Hub
}

// NewHandler 创建WebSocket处理器
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

// Connect WebSocket Connect
func (h *Handler) Connect(c *gin.Context) {
	// 直接从 context 拿
	userInfoAny, exists := c.Get("UserInfo")
	if !exists || userInfoAny == nil {
		logger.Warn("WebSocket处理失败", zap.String("error", "未授权"))
		response.Error(c, http.StatusUnauthorized, "未授权")
		return
	}

	// 类型断言成 *TokenPayload
	payload, ok := userInfoAny.(*token.TokenPayload)
	if !ok {
		logger.Warn("用户信息类型错误")
		response.Error(c, http.StatusInternalServerError, "用户信息类型错误")
		return
	}

	userID := uint(payload.UserID)
	// 验证用户是否存在
	var user model.User
	if err := h.db.First(&user, userID).Error; err != nil {
		response.Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	// 升级HTTP连接为WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket升级失败", zap.Uint("user_id", userID), zap.Error(err))
		return
	}

	connection := NewConnection(userID, conn, h.hub)
	connection.SetMetadata("username", user.Username)
	connection.SetMetadata("role", user.Role)
	connection.SetMetadata("ip", c.ClientIP())
	connection.Start()
}

// GetOnlineUsers 获取在线用户列表
func (h *Handler) GetOnlineUsers(c *gin.Context) {
	users := h.hub.GetOnlineUsers()

	// 获取用户详细信息
	var userList []map[string]interface{}
	for _, userID := range users {
		var user model.User
		if err := h.db.First(&user, userID).Error; err == nil {
			userInfo := map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"avatar":   user.Avatar,
				"role":     user.Role,
			}

			// 获取连接信息
			if conn, exists := h.hub.GetConnection(userID); exists {
				userInfo["conn_id"] = conn.ID
				userInfo["last_heartbeat"] = conn.LastHeartbeat
			}

			userList = append(userList, userInfo)
		}
	}

	response.Success(c, gin.H{
		"total": len(userList),
		"users": userList,
	})
}

// SendMessage 发送消息给指定用户
func (h *Handler) SendMessage(c *gin.Context) {
	var req struct {
		To      []uint          `json:"to" binding:"required"`
		Type    string          `json:"type" binding:"required"`
		Content json.RawMessage `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	// 获取发送者ID
	senderIDStr, _ := c.Get("user_id")
	senderID := uint(senderIDStr.(float64))

	// 构建消息
	msg := &Message{
		Type:      req.Type,
		From:      senderID,
		To:        req.To,
		Content:   req.Content,
		Timestamp: time.Now().Unix(),
	}

	// 发送消息
	h.hub.SendToUsers(req.To, msg)

	response.Success(c, gin.H{
		"message": "消息已发送",
		"to":      req.To,
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

	// 获取发送者ID
	senderIDStr, _ := c.Get("user_id")
	senderID := uint(senderIDStr.(float64))

	// 构建消息
	msg := &Message{
		Type:      req.Type,
		From:      senderID,
		Content:   req.Content,
		Timestamp: time.Now().Unix(),
	}

	// 广播消息
	h.hub.Broadcast(msg)

	response.Success(c, gin.H{
		"message": "消息已广播",
		"online":  len(h.hub.GetOnlineUsers()),
	})
}

// GetStats 获取WebSocket统计信息
func (h *Handler) GetStats(c *gin.Context) {
	stats := h.hub.GetStats()
	response.Success(c, gin.H{
		"stats": stats,
	})
}

// DisconnectUser 断开指定用户连接（管理员功能）
func (h *Handler) DisconnectUser(c *gin.Context) {
	// 检查是否为管理员
	roleStr, _ := c.Get("role")
	if roleStr != "admin" {
		response.Error(c, http.StatusForbidden, "权限不足")
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的用户ID")
		return
	}

	// 获取连接并关闭
	if conn, exists := h.hub.GetConnection(uint(userID)); exists {
		conn.Close()
		response.Success(c, gin.H{
			"message": "用户连接已断开",
			"user_id": userID,
		})
	} else {
		response.Error(c, http.StatusNotFound, "用户未在线")
	}
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

	// 添加用户到分组
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

	// 获取发送者ID
	senderIDStr, _ := c.Get("user_id")
	senderID := uint(senderIDStr.(float64))

	// 构建消息
	msg := &Message{
		Type:      req.Type,
		From:      senderID,
		Content:   req.Content,
		Timestamp: time.Now().Unix(),
	}

	// 发送给分组
	h.hub.SendToGroup(req.GroupName, msg)

	response.Success(c, gin.H{
		"message":    "消息已发送到分组",
		"group_name": req.GroupName,
	})
}

// RegisterRoutes 注册WebSocket路由
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	ws := router.Group("/ws")
	{
		// WebSocket连接端点
		ws.GET("/connect", h.Connect)

		// HTTP API端点
		ws.GET("/online", h.GetOnlineUsers)
		ws.POST("/send", h.SendMessage)
		ws.POST("/broadcast", h.BroadcastMessage)
		ws.GET("/stats", h.GetStats)

		// 管理功能
		ws.DELETE("/disconnect/:id", h.DisconnectUser)

		// 分组功能
		ws.POST("/group", h.CreateGroup)
		ws.POST("/group/send", h.SendToGroup)
	}
}

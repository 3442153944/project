package base

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/pkg/response"
)

// BaseHandler 基类Handler，所有业务Handler都继承它
type BaseHandler struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// NewBaseHandler 创建基类Handler
func NewBaseHandler(db *gorm.DB, redis *redis.Client) *BaseHandler {
	return &BaseHandler{
		DB:    db,
		Redis: redis,
	}
}

// HandleGET 默认的GET处理（未重写则返回405）
func (h *BaseHandler) HandleGET(c *gin.Context) {
	response.MethodNotAllowed(c, "GET method not implemented")
}

// HandlePOST 默认的POST处理（未重写则返回405）
func (h *BaseHandler) HandlePOST(c *gin.Context) {
	response.MethodNotAllowed(c, "POST method not implemented")
}

// HandlePUT 默认的PUT处理（未重写则返回405）
func (h *BaseHandler) HandlePUT(c *gin.Context) {
	response.MethodNotAllowed(c, "PUT method not implemented")
}

// HandleDELETE 默认的DELETE处理（未重写则返回405）
func (h *BaseHandler) HandleDELETE(c *gin.Context) {
	response.MethodNotAllowed(c, "DELETE method not implemented")
}

// HandlePATCH 默认的PATCH处理（未重写则返回405）
func (h *BaseHandler) HandlePATCH(c *gin.Context) {
	response.MethodNotAllowed(c, "PATCH method not implemented")
}

// GetDB 获取数据库连接
func (h *BaseHandler) GetDB() *gorm.DB {
	return h.DB
}

// GetRedis 获取Redis连接
func (h *BaseHandler) GetRedis() *redis.Client {
	return h.Redis
}

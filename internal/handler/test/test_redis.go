package test

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/base"
	"project/pkg/response"
	"time"
)

// TestRedisHandler Redis测试Handler
type TestRedisHandler struct {
	*base.BaseHandler // 继承基类
}

// NewTestRedisHandler 创建Handler
func NewTestRedisHandler(db *gorm.DB, redis *redis.Client) *TestRedisHandler {
	return &TestRedisHandler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

// HandleGET 重写GET方法
func (h *TestRedisHandler) HandleGET(c *gin.Context) {
	ctx := context.Background()

	// 从基类获取Redis连接
	rdb := h.GetRedis()

	// 测试读取
	val, err := rdb.Get(ctx, "test_key").Result()
	if err != nil {
		response.Error(c, 500, "Redis读取失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"message": "Redis GET成功",
		"value":   val,
	})
}

// HandlePOST 重写POST方法
func (h *TestRedisHandler) HandlePOST(c *gin.Context) {
	ctx := context.Background()

	// 解析请求参数
	var req struct {
		Key   string `json:"key" binding:"required"`
		Value string `json:"value" binding:"required"`
		TTL   int    `json:"ttl"` // 秒
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 从基类获取Redis连接
	rdb := h.GetRedis()

	// 写入Redis
	ttl := time.Duration(req.TTL) * time.Second
	if req.TTL == 0 {
		ttl = 0 // 永不过期
	}

	err := rdb.Set(ctx, req.Key, req.Value, ttl).Err()
	if err != nil {
		response.InternalError(c, "Redis写入失败: "+err.Error())
		return
	}

	response.SuccessWithMsg(c, "Redis写入成功", gin.H{
		"key":   req.Key,
		"value": req.Value,
		"ttl":   req.TTL,
	})
}

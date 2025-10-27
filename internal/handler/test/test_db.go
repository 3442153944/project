package test

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"project/internal/base"
	"project/pkg/response"
)

// TestDBHandler 数据库测试Handler
type TestDBHandler struct {
	*base.BaseHandler
}

// NewTestDBHandler 创建Handler
func NewTestDBHandler(db *gorm.DB, redis *redis.Client) *TestDBHandler {
	return &TestDBHandler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

// HandleGET 重写GET方法
func (h *TestDBHandler) HandleGET(c *gin.Context) {
	// 从基类获取数据库连接
	db := h.GetDB()

	// 查询当前时间
	var result struct {
		Now string
	}
	db.Raw("SELECT NOW() as now").Scan(&result)

	response.Success(c, gin.H{
		"message": "数据库查询成功",
		"time":    result.Now,
	})
}

// HandlePOST 重写POST方法
func (h *TestDBHandler) HandlePOST(c *gin.Context) {
	// 解析请求参数
	var req struct {
		SQL string `json:"sql" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 从基类获取数据库连接
	db := h.GetDB()

	// 执行查询
	var results []map[string]interface{}
	db.Raw(req.SQL).Scan(&results)

	response.Success(c, gin.H{
		"message": "SQL执行成功",
		"sql":     req.SQL,
		"results": results,
	})
}

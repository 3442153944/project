package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/base"
	"github.com/sunyuanling/server/internal/model"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
	tokenFunc "github.com/sunyuanling/server/pkg/tokn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	*base.BaseHandler
}

func NewGetDownloadHis(db *gorm.DB, redis *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

type handlerReq struct {
	PageNum  int `json:"pageNum"`
	PageSize int `json:"pageSize"`
}

type handlerResp struct {
	List     []model.DownloadHistory `json:"list"`
	Total    int64                   `json:"total"`
	PageNum  int                     `json:"pageNum"`
	PageSize int                     `json:"pageSize"`
}

func (h *Handler) HandlerPOST(c *gin.Context) {
	isAuth := c.GetBool("Auth")
	if !isAuth {
		response.Unauthorized(c, "请先登录")
		return
	}

	var req handlerReq
	if c.Request.Body != nil && c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Error("参数错误", zap.Error(err))
			response.BadRequest(c, "参数错误")
			return
		}
	}

	userInfo, exists := c.Get("UserInfo")
	if !exists || userInfo == nil {
		response.InternalError(c, "用户信息获取失败")
		return
	}
	payload, ok := userInfo.(*tokenFunc.TokenPayload)
	if !ok {
		response.InternalError(c, "用户信息类型错误")
		return
	}
	userID := uint(payload.UserID)

	// 查询总数
	var total int64
	if err := h.DB.Model(&model.DownloadHistory{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		logger.Error("查询下载记录总数失败", zap.Error(err))
		response.Error(c, 500, "查询下载记录失败")
		return
	}

	// 构建查询
	query := h.DB.Where("user_id = ?", userID).Order("created_at desc")

	// 有分页参数才分页，否则查全部
	if req.PageNum > 0 && req.PageSize > 0 {
		query = query.Limit(req.PageSize).Offset((req.PageNum - 1) * req.PageSize)
	}

	var downloadHis []model.DownloadHistory
	if err := query.Find(&downloadHis).Error; err != nil {
		logger.Error("查询下载记录失败", zap.Error(err))
		response.Error(c, 500, "查询下载记录失败")
		return
	}

	response.Success(c, handlerResp{
		List:     downloadHis,
		Total:    total,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
}

package handler

import (
	"bytes"
	"fmt"
	"github.com/sunyuanling/server/pkg/logger"
	"go.uber.org/zap"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/base"
	"github.com/sunyuanling/server/internal/model"
	_ "github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
	tokenPkg "github.com/sunyuanling/server/pkg/tokn"
	"gorm.io/gorm"
)

type UpdateUserInfoHandler struct {
	*base.BaseHandler
	cfg *config.Config
}

func NewUpdateUserInfoHandler(db *gorm.DB, redis *redis.Client, cfg *config.Config) *UpdateUserInfoHandler {
	return &UpdateUserInfoHandler{
		BaseHandler: base.NewBaseHandler(db, redis),
		cfg:         cfg,
	}
}

// UpdateUserInfoRequest 只允许更新这些字段，ID 从 token 取，username/password 不可改
type UpdateUserInfoRequest struct {
	Username string `form:"username"`
	Email    string `form:"email"`
	Phone    string `form:"phone"`
	Avatar   string `form:"-"`
}

func (h *UpdateUserInfoHandler) HandlePOST(c *gin.Context) {
	// 1. 验证登录状态
	isAuth := c.GetBool("Auth")
	if !isAuth {
		response.Unauthorized(c, "请先登录")
		return
	}

	// 2. 从 token payload 取用户 ID（中间件存的是 *tokenPkg.TokenPayload）
	payloadRaw, exists := c.Get("UserInfo")
	if !exists || payloadRaw == nil {
		response.InternalError(c, "用户信息获取失败")
		return
	}
	payload, ok := payloadRaw.(*tokenPkg.TokenPayload)
	if !ok {
		response.InternalError(c, "用户信息类型错误")
		return
	}
	userID := uint(payload.UserID)

	// 3. 解析请求体
	var req UpdateUserInfoRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Error("参数解析失败", zap.Error(err))
		response.BadRequest(c, "参数解析失败: "+err.Error())
		return
	}

	// 4. 处理头像上传（可选）
	avatarRelPath, err := h.handleAvatarUpload(c, userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 5. 构建更新 map
	updates := map[string]any{}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if avatarRelPath != "" {
		updates["avatar"] = avatarRelPath
	}
	if len(updates) == 0 {
		logger.Warn("没有需要更新的字段", zap.Any("updates", updates))
		response.BadRequest(c, "没有需要更新的字段")
		return
	}

	// 6. 写库
	if err := h.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		logger.Error("更新用户信息失败", zap.Error(err))
		response.InternalError(c, "更新用户信息失败")
		return
	}

	response.Success(c, gin.H{"message": "更新成功"})
}

// handleAvatarUpload 处理头像上传，统一转为 PNG 保存，返回相对路径
// 未上传头像时返回 ("", nil)
func (h *UpdateUserInfoHandler) handleAvatarUpload(c *gin.Context, userID uint) (string, error) {
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		return "", nil
	}
	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			logger.Error("关闭文件失败", zap.Error(err))
		}
	}(file)

	userCfg := h.cfg.GetUserConfig()

	// 校验文件大小
	if header.Size > userCfg.MaxSize {
		return "", fmt.Errorf("头像文件超过最大限制 %dMB", userCfg.MaxSize/1024/1024)
	}

	// 校验扩展名
	origExt := strings.ToLower(filepath.Ext(header.Filename))
	if !isAvatarExtAllowed(origExt, userCfg.AllowedExtensions) {
		return "", fmt.Errorf("不支持的头像格式: %s", origExt)
	}

	// 解码图片
	img, err := decodeImage(file)
	if err != nil {
		return "", fmt.Errorf("图片解析失败: %s", err.Error())
	}

	// 头像保存到项目根目录 static/avatar/
	// os.Executable() 获取可执行文件路径，往上找到项目根
	projectRoot, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取项目路径失败: %s", err.Error())
	}
	avatarDir := filepath.Join(projectRoot, "static", "avatar")
	println("avatarDir:", avatarDir)

	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return "", fmt.Errorf("创建头像目录失败: %s", err.Error())
	}

	// 文件名：avatar_{userID}_{时间戳}.png
	filename := fmt.Sprintf("avatar_%d_%d.png", userID, time.Now().UnixMilli())
	savePath := filepath.Join(avatarDir, filename)
	println("savePath:", savePath)

	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("创建头像文件失败: %s", err.Error())
	}
	defer func(out *os.File) {
		if err := out.Close(); err != nil {
			logger.Error("关闭文件失败", zap.Error(err))
		}
	}(out)

	if err := png.Encode(out, img); err != nil {

		_ = os.Remove(savePath)
		return "", fmt.Errorf("保存头像失败: %s", err.Error())
	}

	// 返回相对 URL 路径，前端拼 baseURL 访问
	return "static/avatar/" + filename, nil
}

// decodeImage 将上传文件解码为 image.Image
func decodeImage(f multipart.File) (image.Image, error) {
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// isAvatarExtAllowed 检查扩展名是否在允许列表中（配置为空则不限制）
func isAvatarExtAllowed(ext string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if strings.ToLower(a) == ext {
			return true
		}
	}
	return false
}

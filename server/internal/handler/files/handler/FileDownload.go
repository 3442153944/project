package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	tokenFunc "github.com/sunyuanling/server/pkg/tokn"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/base"
	_interface "github.com/sunyuanling/server/internal/handler/files/interface"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
	"github.com/sunyuanling/server/websocket"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FileRequest 文件请求参数
type FileRequest struct {
	Path       string `json:"path" binding:"required"` // 文件路径
	Name       string `json:"name" binding:"required"` // 文件名
	FileID     string `json:"file_id,omitempty"`       // 文件ID（分片下载时必填）
	ChunkIndex *int   `json:"chunk_index,omitempty"`   // 分片索引（nil=初始化，>=0=下载分片）
	ChunkSize  int64  `json:"chunk_size,omitempty"`    // 分片大小（可选，默认1MB）
}

// FileInfoResponse 文件信息响应（初始化阶段返回）
type FileInfoResponse struct {
	FileID      string `json:"file_id"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
	MimeType    string `json:"mime_type"`
	ModTime     int64  `json:"mod_time"`
	Hash        string `json:"hash,omitempty"` // 文件哈希（可用于校验）
}

// FileSession 文件下载会话（存储在Redis中）
type FileSession struct {
	FileID      string `json:"file_id"`
	FilePath    string `json:"file_path"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int    `json:"total_chunks"`
	UserID      uint   `json:"user_id"`
	CreateTime  int64  `json:"create_time"`
}

const (
	DefaultChunkSize int64 = 1 << 20          // 1MB
	MaxChunkSize     int64 = 5 << 20          // 5MB
	SessionExpire          = time.Hour        // 会话过期时间
	SessionKeyPrefix       = "file:download:" // Redis key前缀
)

type getFile struct {
	*base.BaseHandler
	cfg *config.Config
}

func (g *getFile) HandlerPOST(c *gin.Context) {
	// 验证登录状态
	isAuth := c.GetBool("Auth")
	logger.Info("isAuth状态", zap.Bool("isAuth", isAuth))

	if !isAuth {
		response.Unauthorized(c, "请先登录")
		return
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

	userID := uint(payload.UserID) // TokenPayload.UserID 是 int64

	// 解析请求参数
	var req FileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数解析失败", zap.Error(err))
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	logger.Info("请求参数",
		zap.String("path", req.Path),
		zap.String("name", req.Name),
		zap.Any("chunk_index", req.ChunkIndex),
	)

	if req.ChunkIndex == nil {
		g.initDownload(c, &req, userID)
	} else {
		g.downloadChunk(c, &req, userID)
	}
}

// initDownload 初始化下载，返回文件信息并创建会话
func (g *getFile) initDownload(c *gin.Context, req *FileRequest, userID uint) {
	// 构建完整路径
	fullPath := filepath.Join(req.Path, req.Name)

	// 验证路径安全性
	if !g.isPathAllowed(fullPath) {
		response.Forbidden(c, "无权访问该路径")
		return
	}

	// 获取文件信息
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			response.NotFound(c, "文件不存在")
		} else {
			response.InternalError(c, "获取文件信息失败: "+err.Error())
		}
		return
	}

	if fileInfo.IsDir() {
		response.BadRequest(c, "不能下载目录")
		return
	}

	// 确定分片大小
	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if chunkSize > MaxChunkSize {
		chunkSize = MaxChunkSize
	}

	// 计算总分片数
	totalChunks := int((fileInfo.Size() + chunkSize - 1) / chunkSize)
	if fileInfo.Size() == 0 {
		totalChunks = 1 // 空文件也算1个分片
	}

	// 生成文件ID
	fileID := uuid.New().String()

	// 创建会话并存储到Redis
	session := FileSession{
		FileID:      fileID,
		FilePath:    fullPath,
		FileName:    req.Name,
		FileSize:    fileInfo.Size(),
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		UserID:      userID,
		CreateTime:  time.Now().Unix(),
	}

	if err := g.saveSession(c, &session); err != nil {
		response.InternalError(c, "创建下载会话失败")
		return
	}

	// 通过WebSocket通知客户端
	err = websocket.SendToUser(userID, "file_download", map[string]interface{}{
		"event":        "init",
		"file_id":      fileID,
		"file_name":    req.Name,
		"file_size":    fileInfo.Size(),
		"total_chunks": totalChunks,
		"status":       "ready",
		"time":         time.Now().Unix(),
	})
	if err != nil {
		return
	}

	// 返回文件信息
	response.Success(c, FileInfoResponse{
		FileID:      fileID,
		FileName:    req.Name,
		FilePath:    req.Path,
		FileSize:    fileInfo.Size(),
		ChunkSize:   chunkSize,
		TotalChunks: totalChunks,
		MimeType:    getMimeType(req.Name),
		ModTime:     fileInfo.ModTime().Unix(),
	})
}

// downloadChunk 下载指定分片
func (g *getFile) downloadChunk(c *gin.Context, req *FileRequest, userID uint) {
	// 验证FileID
	if req.FileID == "" {
		response.BadRequest(c, "缺少file_id参数")
		return
	}

	// 从Redis获取会话
	session, err := g.getSession(c, req.FileID)
	if err != nil {
		response.BadRequest(c, "下载会话不存在或已过期，请重新初始化")
		return
	}

	// 验证用户
	if session.UserID != userID {
		response.Forbidden(c, "无权访问此下载会话")
		return
	}

	chunkIndex := *req.ChunkIndex

	// 验证分片索引
	if chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		response.BadRequest(c, fmt.Sprintf("分片索引越界，有效范围: 0-%d", session.TotalChunks-1))
		return
	}

	// 打开文件
	file, err := os.Open(session.FilePath)
	if err != nil {
		response.InternalError(c, "打开文件失败")
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// 计算偏移量和读取大小
	offset := int64(chunkIndex) * session.ChunkSize
	readSize := session.ChunkSize
	if offset+readSize > session.FileSize {
		readSize = session.FileSize - offset
	}

	// 空文件特殊处理
	if session.FileSize == 0 {
		readSize = 0
	}

	// 定位到指定位置
	if offset > 0 {
		if _, err := file.Seek(offset, io.SeekStart); err != nil {
			response.InternalError(c, "文件定位失败")
			return
		}
	}

	// 计算进度
	progress := int(float64(chunkIndex+1) / float64(session.TotalChunks) * 100)
	isLast := chunkIndex == session.TotalChunks-1

	// 通过WebSocket通知进度
	status := "downloading"
	if isLast {
		status = "completed"
	}

	err = websocket.SendToUser(userID, "file_download", map[string]interface{}{
		"event":        "progress",
		"file_id":      session.FileID,
		"file_name":    session.FileName,
		"chunk_index":  chunkIndex,
		"total_chunks": session.TotalChunks,
		"progress":     progress,
		"status":       status,
		"time":         time.Now().Unix(),
	})
	if err != nil {
		return
	}

	// 如果是最后一个分片，删除会话
	if isLast {
		g.deleteSession(c, req.FileID)
	}

	// 设置响应头
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(readSize, 10))
	c.Header("X-File-ID", session.FileID)
	c.Header("X-Chunk-Index", strconv.Itoa(chunkIndex))
	c.Header("X-Total-Chunks", strconv.Itoa(session.TotalChunks))
	c.Header("X-File-Size", strconv.FormatInt(session.FileSize, 10))

	// 输出数据
	c.Status(200)
	if readSize > 0 {
		written, err := io.CopyN(c.Writer, file, readSize)
		if err != nil && err != io.EOF {
			logger.Error("写入分片数据失败",
				zap.Error(err),
				zap.Int("chunk_index", chunkIndex),
				zap.Int64("written", written),
			)
		}
	}
}

// saveSession 保存会话到Redis
func (g *getFile) saveSession(c context.Context, session *FileSession) error {
	key := SessionKeyPrefix + session.FileID
	data, _ := json.Marshal(session)
	return g.Redis.Set(c, key, data, SessionExpire).Err()
}

// getSession 从Redis获取会话
func (g *getFile) getSession(c context.Context, fileID string) (*FileSession, error) {
	key := SessionKeyPrefix + fileID
	data, err := g.Redis.Get(c, key).Bytes()
	if err != nil {
		return nil, err
	}

	var session FileSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

// deleteSession 删除会话
func (g *getFile) deleteSession(c context.Context, fileID string) {
	key := SessionKeyPrefix + fileID
	g.Redis.Del(c, key)
}

// isPathAllowed 检查路径是否在允许范围内
func (g *getFile) isPathAllowed(path string) bool {
	// 1. 清理路径，防止 ../../../ 等穿透攻击
	cleanPath := filepath.Clean(path)
	cleanPathUpper := strings.ToUpper(cleanPath)

	// 2. 检查是否是绝对路径
	if !filepath.IsAbs(cleanPath) {
		logger.Warn("拒绝相对路径", zap.String("path", path))
		return false
	}

	// 3. 检查是否包含可疑字符（双重保险）
	if strings.Contains(path, "..") {
		logger.Warn("拒绝包含..的路径", zap.String("path", path))
		return false
	}

	allowedPaths := g.cfg.GetAllowedPaths()

	for _, allowed := range allowedPaths {
		allowedUpper := strings.ToUpper(allowed)

		// Windows 盘符格式：如 "E:"
		if len(allowedUpper) == 2 && allowedUpper[1] == ':' {
			if strings.HasPrefix(cleanPathUpper, allowedUpper) {
				return true
			}
			continue
		}

		// Linux 路径或完整目录路径
		allowedClean := filepath.Clean(allowed)
		allowedCleanUpper := strings.ToUpper(allowedClean)

		// 检查是否在允许目录下
		if strings.HasPrefix(cleanPathUpper, allowedCleanUpper) {
			// 确保是目录边界，防止 /home/user 匹配 /home/username
			if len(cleanPath) == len(allowedClean) ||
				cleanPath[len(allowedClean)] == filepath.Separator {
				return true
			}
		}
	}

	return false
}

// getMimeType 根据文件扩展名获取MIME类型
func getMimeType(filename string) string {
	ext := filepath.Ext(filename)
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
	}

	if mime, ok := mimeTypes[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}

func NewGetFile(db *gorm.DB, redis *redis.Client, cfg *config.Config) _interface.GetFile {
	return &getFile{
		BaseHandler: base.NewBaseHandler(db, redis),
		cfg:         cfg,
	}
}

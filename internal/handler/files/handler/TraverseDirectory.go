package handler

import (
	"os"
	"path/filepath"
	_interface "project/internal/handler/files/interface"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"project/internal/base"
	"project/pkg/logger"
	"project/pkg/response"
)

type traverseDirectory struct {
	*base.BaseHandler
}

func NewTraverseDirectory(db *gorm.DB, redis *redis.Client) _interface.TraverseDirectory {
	return &traverseDirectory{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

// TraverseRequest 请求参数结构体
type TraverseRequest struct {
	Path string `json:"path" binding:"required"`
}

// FileItem 文件/目录项结构体
type FileItem struct {
	Name          string    `json:"name"`                     // 文件名
	Path          string    `json:"path"`                     // 完整路径
	IsDir         bool      `json:"is_dir"`                   // 是否为目录
	Size          int64     `json:"size,omitempty"`           // 文件大小（字节）
	ModTime       time.Time `json:"mod_time"`                 // 修改时间
	Mode          string    `json:"mode,omitempty"`           // 文件权限
	Extension     string    `json:"extension,omitempty"`      // 文件扩展名
	ChildrenCount int       `json:"children_count,omitempty"` // 子项目数量（仅目录有）
}

// TraverseResponse 目录遍历响应结构体
type TraverseResponse struct {
	CurrentPath string     `json:"current_path"`          // 当前路径
	ParentPath  string     `json:"parent_path,omitempty"` // 父级路径（如果有）
	Items       []FileItem `json:"items"`                 // 文件/目录列表
	TotalCount  int        `json:"total_count"`           // 总项目数
	DirCount    int        `json:"dir_count"`             // 目录数量
	FileCount   int        `json:"file_count"`            // 文件数量
}

func (h *traverseDirectory) HandlerPOST(c *gin.Context) {
	var req TraverseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数错误", zap.Error(err))
		response.BadRequest(c, "参数错误")
		return
	}

	// 清理路径
	req.Path = filepath.Clean(req.Path)

	// 检查路径是否存在
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		logger.Error("目录不存在", zap.String("path", req.Path), zap.Error(err))
		response.Error(c, 404, "目录不存在")
		return
	}

	// 检查是否为目录
	fileInfo, err := os.Stat(req.Path)
	if err != nil {
		logger.Error("无法访问目录", zap.String("path", req.Path), zap.Error(err))
		response.Error(c, 500, "无法访问目录")
		return
	}

	if !fileInfo.IsDir() {
		logger.Error("路径不是目录", zap.String("path", req.Path))
		response.Error(c, 400, "路径不是目录")
		return
	}

	// 遍历目录（只返回一层）
	result, err := h.traverseSingleLevel(req.Path)
	if err != nil {
		logger.Error("遍历目录失败", zap.String("path", req.Path), zap.Error(err))
		response.Error(c, 500, "遍历目录失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// traverseSingleLevel 只遍历当前目录的直接子项
func (h *traverseDirectory) traverseSingleLevel(path string) (*TraverseResponse, error) {
	// 读取目录内容
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var items []FileItem
	var dirCount, fileCount int

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			// 如果无法获取文件信息，跳过该项
			continue
		}

		fullPath := filepath.Join(path, entry.Name())
		item := FileItem{
			Name:    entry.Name(),
			Path:    fullPath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode().String(),
		}

		// 如果是目录，统计子项目数量
		if item.IsDir {
			dirCount++
			subEntries, err := os.ReadDir(fullPath)
			if err == nil {
				item.ChildrenCount = len(subEntries)
			}
		} else {
			fileCount++
			// 设置文件扩展名
			item.Extension = strings.ToLower(filepath.Ext(item.Name))
			// 如果扩展名是空的，可能是隐藏文件或无扩展名文件
			if item.Extension == "" {
				item.Extension = "unknown"
			}
		}

		items = append(items, item)
	}

	// 排序：目录在前，文件在后，按名称排序
	sort.Slice(items, func(i, j int) bool {
		// 目录优先
		if items[i].IsDir && !items[j].IsDir {
			return true
		}
		if !items[i].IsDir && items[j].IsDir {
			return false
		}
		// 同类型按名称排序（不区分大小写）
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	// 构建响应
	res := &TraverseResponse{
		CurrentPath: path,
		Items:       items,
		TotalCount:  len(items),
		DirCount:    dirCount,
		FileCount:   fileCount,
	}

	// 计算父级路径（如果不是根目录）
	if path != filepath.VolumeName(path)+string(filepath.Separator) && path != "/" {
		res.ParentPath = filepath.Dir(path)
	}

	return res, nil
}

// HandlePOSTWithPagination 可选：添加分页支持的方法
func (h *traverseDirectory) HandlePOSTWithPagination(c *gin.Context) {
	var req struct {
		TraverseRequest
		Page     int `json:"page" binding:"min=1"`      // 页码，从1开始
		PageSize int `json:"page_size" binding:"min=1"` // 每页大小
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数错误", zap.Error(err))
		response.BadRequest(c, "参数错误")
		return
	}

	req.Path = filepath.Clean(req.Path)

	// 路径验证（同上）
	fileInfo, err := os.Stat(req.Path)
	if err != nil || !fileInfo.IsDir() {
		response.Error(c, 500, "目录不存在")
		return
	}

	// 获取所有项目
	result, err := h.traverseSingleLevel(req.Path)
	if err != nil {
		logger.Error("遍历目录失败", zap.String("path", req.Path), zap.Error(err))
		response.Error(c, 500, "遍历目录失败")
		return
	}

	// 分页处理
	start := (req.Page - 1) * req.PageSize
	end := req.Page * req.PageSize

	if start > len(result.Items) {
		start = len(result.Items)
	}
	if end > len(result.Items) {
		end = len(result.Items)
	}

	pagedItems := result.Items[start:end]

	response.Success(c, gin.H{
		"current_path": result.CurrentPath,
		"parent_path":  result.ParentPath,
		"items":        pagedItems,
		"pagination": gin.H{
			"page":        req.Page,
			"page_size":   req.PageSize,
			"total_count": result.TotalCount,
			"total_pages": (result.TotalCount + req.PageSize - 1) / req.PageSize,
		},
		"summary": gin.H{
			"dir_count":  result.DirCount,
			"file_count": result.FileCount,
		},
	})
}

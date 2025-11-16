package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/disk"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sunyuanling/server/config"
	"github.com/sunyuanling/server/internal/base"
	_interface "github.com/sunyuanling/server/internal/handler/files/interface"
	"github.com/sunyuanling/server/pkg/logger"
	"github.com/sunyuanling/server/pkg/response"
)

// DiskInfoBrief 磁盘简略信息
type DiskInfoBrief struct {
	Path         string  `json:"path"`          // 磁盘路径（Windows: D:, Linux: /home）
	Mountpoint   string  `json:"mountpoint"`    // 挂载点
	Device       string  `json:"device"`        // 设备名
	Fstype       string  `json:"fstype"`        // 文件系统类型
	Total        uint64  `json:"total"`         // 总容量（字节）
	Free         uint64  `json:"free"`          // 可用容量（字节）
	Used         uint64  `json:"used"`          // 已用容量（字节）
	UsedPercent  float64 `json:"used_percent"`  // 使用百分比
	TotalGB      string  `json:"total_gb"`      // 总容量（格式化）
	FreeGB       string  `json:"free_gb"`       // 可用容量（格式化）
	IsAllowed    bool    `json:"is_allowed"`    // 是否在配置的允许列表中
	IsAccessible bool    `json:"is_accessible"` // 是否可访问
	IsSSD        bool    `json:"is_ssd"`        // 是否为SSD
}

// DiskInfoDetail 磁盘详细信息
type DiskInfoDetail struct {
	DiskInfoBrief
	StoragePaths StoragePaths `json:"storage_paths"` // 存储路径信息
	FileStats    FileStats    `json:"file_stats"`    // 文件统计
	IOStats      *IOStats     `json:"io_stats"`      // IO统计（可选）
	HealthStatus HealthStatus `json:"health_status"` // 健康状态
}

// StoragePaths 存储路径信息
type StoragePaths struct {
	BasePath   string `json:"base_path"`   // 基础路径（如：D:/FileSync）
	UploadPath string `json:"upload_path"` // 上传路径
	TempPath   string `json:"temp_path"`   // 临时路径
	TrashPath  string `json:"trash_path"`  // 回收站路径
}

// FileStats 文件统计信息
type FileStats struct {
	TotalFiles       int64            `json:"total_files"`       // 总文件数
	TotalDirectories int64            `json:"total_directories"` // 总目录数
	TotalSize        uint64           `json:"total_size"`        // 总大小（字节）
	TotalSizeGB      string           `json:"total_size_gb"`     // 总大小（格式化）
	LargestFile      string           `json:"largest_file"`      // 最大文件
	LargestFileSize  uint64           `json:"largest_file_size"` // 最大文件大小
	RecentFiles      []RecentFileInfo `json:"recent_files"`      // 最近文件
}

// RecentFileInfo 最近文件信息
type RecentFileInfo struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       uint64    `json:"size"`
	SizeFormat string    `json:"size_format"`
	ModTime    time.Time `json:"mod_time"`
	IsDir      bool      `json:"is_dir"`
}

// IOStats IO统计信息
type IOStats struct {
	ReadCount  uint64 `json:"read_count"`
	WriteCount uint64 `json:"write_count"`
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadTime   uint64 `json:"read_time"`
	WriteTime  uint64 `json:"write_time"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status       string  `json:"status"`        // healthy, warning, critical
	SpaceWarning bool    `json:"space_warning"` // 空间不足警告
	Message      string  `json:"message"`       // 状态消息
	Temperature  float64 `json:"temperature"`   // 温度（如果可用）
}

// GetDiskRequest 请求参数
type GetDiskRequest struct {
	DiskPath string `json:"disk_path"` // 磁盘路径（可选）
	Detailed bool   `json:"detailed"`  // 是否返回详细信息
}

type getAvailableDiskList struct {
	*base.BaseHandler
	cfg *config.Config
}

func NewGetAvailableDiskList(db *gorm.DB, redis *redis.Client, cfg *config.Config) _interface.GetAvailableDiskList {
	return &getAvailableDiskList{
		BaseHandler: base.NewBaseHandler(db, redis),
		cfg:         cfg,
	}
}

// HandlerPOST 处理POST请求
func (h *getAvailableDiskList) HandlerPOST(c *gin.Context) {
	var req GetDiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果解析失败，使用默认值
		req = GetDiskRequest{
			DiskPath: "",
			Detailed: false,
		}
	}

	// 如果指定了磁盘路径，返回详细信息
	if req.DiskPath != "" {
		h.handleDetailedDiskInfo(c, req.DiskPath)
		return
	}

	// 否则返回所有磁盘的简略信息
	h.handleBriefDiskList(c)
}

// handleBriefDiskList 处理简略磁盘列表
func (h *getAvailableDiskList) handleBriefDiskList(c *gin.Context) {
	disks, err := h.getBriefDiskList()
	if err != nil {
		logger.Error("获取磁盘列表失败", zap.Error(err))
		response.Error(c, 500, "获取磁盘信息失败")
		return
	}

	// 过滤出允许的磁盘
	var allowedDisks []DiskInfoBrief
	var allDisks []DiskInfoBrief

	for _, diskInfoBrief := range disks {
		allDisks = append(allDisks, diskInfoBrief)
		if diskInfoBrief.IsAllowed && diskInfoBrief.IsAccessible {
			allowedDisks = append(allowedDisks, diskInfoBrief)
		}
	}

	response.Success(c, gin.H{
		"total":         len(allDisks),
		"allowed_count": len(allowedDisks),
		"allowed_disks": allowedDisks,
		"all_disks":     allDisks,
	})
}

// handleDetailedDiskInfo 处理详细磁盘信息
func (h *getAvailableDiskList) handleDetailedDiskInfo(c *gin.Context, diskPath string) {
	// 规范化路径
	diskPath = h.normalizeDiskPath(diskPath)

	// 检查磁盘是否在允许列表中
	if !h.cfg.IsPathAllowed(diskPath) {
		response.Error(c, 403, "该磁盘不在允许列表中")
		return
	}

	// 获取详细信息
	detail, err := h.getDetailedDiskInfo(diskPath)
	if err != nil {
		logger.Error("获取磁盘详细信息失败",
			zap.String("disk", diskPath),
			zap.Error(err),
		)
		response.Error(c, 500, "获取磁盘详细信息失败")
		return
	}

	response.Success(c, detail)
}

// getBriefDiskList 获取简略磁盘列表
func (h *getAvailableDiskList) getBriefDiskList() ([]DiskInfoBrief, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘分区失败: %w", err)
	}

	var disks []DiskInfoBrief
	allowedPaths := h.cfg.GetAllowedPaths()

	for _, partition := range partitions {
		if h.shouldSkipPartition(partition) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		diskPath := h.normalizeDiskPath(partition.Mountpoint)
		isAllowed := h.isPathInList(diskPath, allowedPaths)
		isAccessible := h.isDiskAccessible(partition.Mountpoint)

		diskInfo := DiskInfoBrief{
			Path:         diskPath,
			Mountpoint:   partition.Mountpoint,
			Device:       partition.Device,
			Fstype:       partition.Fstype,
			Total:        usage.Total,
			Free:         usage.Free,
			Used:         usage.Used,
			UsedPercent:  usage.UsedPercent,
			TotalGB:      formatBytes(usage.Total),
			FreeGB:       formatBytes(usage.Free),
			IsAllowed:    isAllowed,
			IsAccessible: isAccessible,
			IsSSD:        h.checkIfSSD(partition.Device),
		}

		disks = append(disks, diskInfo)
	}

	return disks, nil
}

// getDetailedDiskInfo 获取详细磁盘信息
func (h *getAvailableDiskList) getDetailedDiskInfo(diskPath string) (*DiskInfoDetail, error) {
	// 先获取基础信息
	usage, err := disk.Usage(diskPath)
	if err != nil {
		return nil, fmt.Errorf("无法获取磁盘使用情况: %w", err)
	}

	// 获取分区信息
	partitions, _ := disk.Partitions(false)
	var partition *disk.PartitionStat
	for i, p := range partitions {
		if h.normalizeDiskPath(p.Mountpoint) == diskPath {
			partition = &partitions[i]
			break
		}
	}

	if partition == nil {
		return nil, fmt.Errorf("找不到对应的磁盘分区")
	}

	// 构建基础信息
	detail := &DiskInfoDetail{
		DiskInfoBrief: DiskInfoBrief{
			Path:         diskPath,
			Mountpoint:   partition.Mountpoint,
			Device:       partition.Device,
			Fstype:       partition.Fstype,
			Total:        usage.Total,
			Free:         usage.Free,
			Used:         usage.Used,
			UsedPercent:  usage.UsedPercent,
			TotalGB:      formatBytes(usage.Total),
			FreeGB:       formatBytes(usage.Free),
			IsAllowed:    true,
			IsAccessible: true,
			IsSSD:        h.checkIfSSD(partition.Device),
		},
	}

	// 获取存储路径信息
	detail.StoragePaths = h.getStoragePaths(diskPath)

	// 获取文件统计（在goroutine中异步计算，超时则返回部分数据）
	fileStatsChan := make(chan FileStats, 1)
	go func() {
		stats := h.calculateFileStats(diskPath)
		fileStatsChan <- stats
	}()

	// 等待文件统计（最多5秒）
	select {
	case stats := <-fileStatsChan:
		detail.FileStats = stats
	case <-time.After(5 * time.Second):
		logger.Warn("文件统计超时", zap.String("disk", diskPath))
		detail.FileStats = FileStats{
			TotalFiles: -1, // 表示统计超时
		}
	}

	// 获取IO统计
	detail.IOStats = h.getIOStats(partition.Device)

	// 计算健康状态
	detail.HealthStatus = h.calculateHealthStatus(usage)

	return detail, nil
}

// getStoragePaths 获取存储路径信息
func (h *getAvailableDiskList) getStoragePaths(diskPath string) StoragePaths {
	return StoragePaths{
		BasePath:   h.cfg.GetStoragePath(diskPath, ""),
		UploadPath: h.cfg.GetUploadPath(diskPath),
		TempPath:   h.cfg.GetTempPath(diskPath),
		TrashPath:  h.cfg.GetTrashPath(diskPath),
	}
}

// calculateFileStats 计算文件统计
func (h *getAvailableDiskList) calculateFileStats(diskPath string) FileStats {
	stats := FileStats{
		RecentFiles: make([]RecentFileInfo, 0),
	}

	basePath := h.cfg.GetStoragePath(diskPath, "")

	// 如果目录不存在，返回空统计
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return stats
	}

	var largestSize uint64
	var largestFile string
	recentFiles := make([]RecentFileInfo, 0, 10)

	// 遍历目录统计文件
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}

		if info.IsDir() {
			stats.TotalDirectories++
		} else {
			stats.TotalFiles++
			stats.TotalSize += uint64(info.Size())

			// 记录最大文件
			if uint64(info.Size()) > largestSize {
				largestSize = uint64(info.Size())
				largestFile = path
			}

			// 收集最近修改的文件（最多10个）
			if len(recentFiles) < 10 {
				recentFiles = append(recentFiles, RecentFileInfo{
					Name:       info.Name(),
					Path:       path,
					Size:       uint64(info.Size()),
					SizeFormat: formatBytes(uint64(info.Size())),
					ModTime:    info.ModTime(),
					IsDir:      info.IsDir(),
				})
			}
		}

		return nil
	})

	if err != nil {
		logger.Warn("遍历目录失败", zap.Error(err))
	}

	stats.LargestFile = largestFile
	stats.LargestFileSize = largestSize
	stats.TotalSizeGB = formatBytes(stats.TotalSize)
	stats.RecentFiles = recentFiles

	return stats
}

// getIOStats 获取IO统计
func (h *getAvailableDiskList) getIOStats(device string) *IOStats {
	// 尝试获取IO统计
	ioCounters, err := disk.IOCounters(device)
	if err != nil {
		return nil
	}

	if counter, ok := ioCounters[device]; ok {
		return &IOStats{
			ReadCount:  counter.ReadCount,
			WriteCount: counter.WriteCount,
			ReadBytes:  counter.ReadBytes,
			WriteBytes: counter.WriteBytes,
			ReadTime:   counter.ReadTime,
			WriteTime:  counter.WriteTime,
		}
	}

	return nil
}

// calculateHealthStatus 计算健康状态
func (h *getAvailableDiskList) calculateHealthStatus(usage *disk.UsageStat) HealthStatus {
	status := HealthStatus{
		Status:       "healthy",
		SpaceWarning: false,
	}

	// 根据使用率判断健康状态
	if usage.UsedPercent >= 95 {
		status.Status = "critical"
		status.SpaceWarning = true
		status.Message = "磁盘空间严重不足"
	} else if usage.UsedPercent >= 90 {
		status.Status = "warning"
		status.SpaceWarning = true
		status.Message = "磁盘空间不足"
	} else {
		status.Message = "磁盘状态正常"
	}

	return status
}

// shouldSkipPartition 判断是否跳过分区
func (h *getAvailableDiskList) shouldSkipPartition(partition disk.PartitionStat) bool {
	skipFstypes := map[string]bool{
		"tmpfs":    true,
		"devtmpfs": true,
		"squashfs": true,
		"overlay":  true,
		"cgroup":   true,
		"cgroup2":  true,
	}

	if skipFstypes[partition.Fstype] {
		return true
	}

	skipMountpoints := []string{"/boot", "/snap", "/var/snap"}
	for _, skip := range skipMountpoints {
		if strings.HasPrefix(partition.Mountpoint, skip) {
			return true
		}
	}

	return false
}

// checkIfSSD 检查是否为SSD
func (h *getAvailableDiskList) checkIfSSD(device string) bool {
	if runtime.GOOS == "linux" {
		// Linux: 检查 /sys/block/*/queue/rotational
		deviceName := strings.TrimPrefix(device, "/dev/")
		rotationalPath := fmt.Sprintf("/sys/block/%s/queue/rotational", deviceName)

		data, err := os.ReadFile(rotationalPath)
		if err == nil {
			return strings.TrimSpace(string(data)) == "0"
		}
	}

	// 其他系统暂不支持
	return false
}

// normalizeDiskPath 规范化磁盘路径
func (h *getAvailableDiskList) normalizeDiskPath(path string) string {
	if runtime.GOOS == "windows" {
		// Windows: 提取盘符（如 C:）
		vol := filepath.VolumeName(path)
		if vol != "" {
			return strings.ToUpper(vol)
		}
	}
	// Linux: 使用完整路径
	return filepath.Clean(path)
}

// isPathInList 检查路径是否在列表中
func (h *getAvailableDiskList) isPathInList(path string, list []string) bool {
	path = strings.ToUpper(path)
	for _, item := range list {
		if strings.ToUpper(item) == path {
			return true
		}
	}
	return false
}

// isDiskAccessible 检查磁盘是否可访问
func (h *getAvailableDiskList) isDiskAccessible(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// formatBytes 格式化字节大小
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

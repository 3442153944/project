package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/disk"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"project/internal/base"
	_interface "project/internal/handler/files/interface"
	"project/pkg/logger"
	"project/pkg/response"
)

// 磁盘信息结构体
type DiskInfo struct {
	Mountpoint  string  `json:"mountpoint"`   // 挂载点
	Device      string  `json:"device"`       // 设备名
	Fstype      string  `json:"fstype"`       // 文件系统类型
	Total       uint64  `json:"total"`        // 总容量 (字节)
	Free        uint64  `json:"free"`         // 可用容量 (字节)
	Used        uint64  `json:"used"`         // 已用容量 (字节)
	UsedPercent float64 `json:"used_percent"` // 使用百分比
	IsSSD       bool    `json:"is_ssd"`       // 是否为SSD
}

// 改为 struct 而不是 interface
type getAvailableDiskList struct {
	*base.BaseHandler
}

func NewGetAvailableDiskList(db *gorm.DB, redis *redis.Client) _interface.GetAvailableDiskList {
	return &getAvailableDiskList{
		BaseHandler: base.NewBaseHandler(db, redis),
	}
}

// HandlerPOST 方法名改为 HandlerPOST 以匹配接口定义
func (h *getAvailableDiskList) HandlerPOST(c *gin.Context) {
	disks, err := h.getAvailableDisks()
	if err != nil {
		logger.Error("获取磁盘信息失败:", zap.Error(err))
		response.Error(c, 500, "获取磁盘信息失败: "+err.Error())
		return
	}

	response.Success(c, gin.H{
		"disks": disks,
		"count": len(disks),
	})
}

// getAvailableDisks 获取所有可用磁盘信息
func (h *getAvailableDiskList) getAvailableDisks() ([]DiskInfo, error) {
	// 获取所有磁盘分区
	partitions, err := disk.Partitions(true)
	if err != nil {
		logger.Error("获取磁盘分区失败:", zap.Error(err))
		return nil, fmt.Errorf("获取磁盘分区失败: %v", err)
	}

	var disks []DiskInfo

	for _, partition := range partitions {
		// 跳过一些特殊的文件系统
		if h.shouldSkipPartition(partition) {
			continue
		}

		// 获取磁盘使用情况
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			// 如果无法获取使用情况，跳过该分区
			continue
		}

		// 检查是否为SSD（这是一个简化的检查，实际可能需要更复杂的方法）
		isSSD := h.checkIfSSD(partition.Device)

		diskInfo := DiskInfo{
			Mountpoint:  partition.Mountpoint,
			Device:      partition.Device,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
			IsSSD:       isSSD,
		}

		disks = append(disks, diskInfo)
	}

	return disks, nil
}

// shouldSkipPartition 判断是否应该跳过该分区
func (h *getAvailableDiskList) shouldSkipPartition(partition disk.PartitionStat) bool {
	// 跳过一些特殊的文件系统类型
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

	// 跳过一些系统挂载点
	skipMountpoints := []string{
		"/boot",
		"/snap",
		"/var/snap",
	}

	for _, skipMount := range skipMountpoints {
		if partition.Mountpoint == skipMount {
			return true
		}
	}

	return false
}

// checkIfSSD 简化的SSD检查（实际实现可能需要系统特定的调用）
func (h *getAvailableDiskList) checkIfSSD(device string) bool {
	// 这里是一个简化的实现
	// 在实际项目中，你可能需要：
	// - 在Linux上检查 /sys/block/*/queue/rotational
	// - 在Windows上使用WMI查询
	// - 在macOS上使用系统调用

	// 临时返回false，实际项目中需要根据具体系统实现
	return false
}

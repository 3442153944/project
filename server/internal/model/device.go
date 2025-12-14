package model

import (
	"gorm.io/gorm"
	"time"
)

// Device 设备表
type Device struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`                         // 设备ID
	UserID     uint       `gorm:"not null;index:idx_device_user" json:"user_id"`              // 用户ID
	DeviceName string     `gorm:"type:varchar(100);not null" json:"device_name"`              // 设备名称
	DeviceType string     `gorm:"type:varchar(20);not null;index" json:"device_type"`         // 设备类型：mobile/web/windows/mac/linux
	DeviceID   string     `gorm:"type:varchar(100);not null" json:"device_id"`                // 设备唯一标识
	OSVersion  string     `gorm:"type:varchar(50)" json:"os_version"`                         // 操作系统版本
	AppVersion string     `gorm:"type:varchar(50)" json:"app_version"`                        // 应用版本
	IPAddress  string     `gorm:"type:varchar(50)" json:"ip_address"`                         // IP地址
	LastActive *time.Time `gorm:"type:timestamp" json:"last_active"`                          // 最后活跃时间
	Status     int16      `gorm:"type:smallint;default:1;index" json:"status"`                // 状态：1正常/0禁用
	CreatedAt  time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"` // 创建时间
	UpdatedAt  time.Time  `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"` // 更新时间

	// 关联
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"` // 所属用户
}

// TableName 指定表名
func (Device) TableName() string {
	return "device"
}

// 设备类型常量
const (
	DeviceTypeMobile  = "mobile"  // 手机端
	DeviceTypeWeb     = "web"     // Web端
	DeviceTypeWindows = "windows" // Windows端
	DeviceTypeMac     = "mac"     // Mac端
	DeviceTypeLinux   = "linux"   // Linux端
)

// 设备状态常量
const (
	DeviceStatusActive   = 1 // 正常
	DeviceStatusInactive = 0 // 禁用
)

// BeforeCreate GORM钩子：创建前
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	// 设置最后活跃时间为当前时间
	now := time.Now()
	d.LastActive = &now
	return nil
}

// BeforeUpdate GORM钩子：更新前
func (d *Device) BeforeUpdate(tx *gorm.DB) error {
	// 更新时自动更新 updated_at
	d.UpdatedAt = time.Now()
	return nil
}

// IsActive 判断设备是否在线
func (d *Device) IsActive() bool {
	return d.Status == DeviceStatusActive
}

// UpdateLastActive 更新最后活跃时间
func (d *Device) UpdateLastActive(tx *gorm.DB) error {
	now := time.Now()
	d.LastActive = &now
	return tx.Model(d).Update("last_active", now).Error
}

package model

import "time"

// UserRole 用户角色关联表
type UserRole struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_user_role" json:"user_id"`
	RoleID    uint      `gorm:"not null;uniqueIndex:uk_user_role" json:"role_id"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (UserRole) TableName() string { return "user_role" }

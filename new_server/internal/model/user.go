package model

import "time"

// User 用户表
type User struct {
	ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string     `gorm:"size:50;not null;uniqueIndex" json:"username"`
	Password  string     `gorm:"size:255;not null" json:"-"`
	Email     *string    `gorm:"size:100;uniqueIndex" json:"email"`
	Phone     *string    `gorm:"size:20" json:"phone"`
	Avatar    *string    `gorm:"size:255" json:"avatar"`
	Role      string     `gorm:"size:20;not null;default:user" json:"role"`
	Status    int8       `gorm:"not null;default:1" json:"status"`
	LastLogin *time.Time `json:"last_login"`
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (User) TableName() string { return "user" }

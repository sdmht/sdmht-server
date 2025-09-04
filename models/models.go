package models

import (
	"time"

	"gorm.io/gorm"
)

type id struct {
	ID uint `gorm:"primarykey" json:"ID"` // 主键ID
}
type ct struct {
	id
	CreatedAt time.Time // 创建时间
}
type ut struct {
	id
	UpdatedAt time.Time // 更新时间
}
type dt struct {
	id
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // 删除时间
}
type m struct {
	id
	ct
	ut
	dt
}

// 用户模型
type User struct {
	m
	Username string `gorm:"unique;not null" json:"username"` // 用户名，唯一且不能为空
	Password string `gorm:"not null" json:"password"`        // 密码，不能为空
}

// 推送订阅模型
type PushSubscription struct {
	m
	Endpoint       string     `gorm:"unique;not null" json:"endpoint"`       // 端点URL，唯一且不能为空
	P256dh         string     `gorm:"not null" json:"p256dh"`                // 公钥
	Auth           string     `gorm:"not null" json:"auth"`                  // 认证密钥
	ExpirationTime *time.Time `gorm:"index" json:"expirationTime,omitempty"` // 过期时间
}

func AutoMigrate(db *gorm.DB) {
	db.AutoMigrate(&User{}, &PushSubscription{})
}

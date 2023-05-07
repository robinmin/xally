package serverdb

import (
	"time"

	"gorm.io/gorm"
)

// type User struct {
// 	gorm.Model

// 	UserID     string `gorm:"size:64;not null;" json:"uid"`
// 	Status     uint8  `gorm:"not null;default:0;" json:"status"`
// 	DeviceInfo string `gorm:"size:255;not null;uniqueIndex:uniq_user" json:"device_info"`
// 	Email      string `gorm:"size:128;not null;uniqueIndex:uniq_user" json:"email"`
// 	Username   string `gorm:"size:64;not null;" json:"username"`
// 	// Token      string `gorm:"size:512;not null;" json:"token"`

// 	RegisterAt   time.Time `gorm:"autoCreateTime" json:"register_at"`
// 	ActivateAt   time.Time `json:"activate_at"`
// 	ExpiredAt    time.Time `gorm:"index" json:"expired_at"`
// 	DeactivateAt time.Time `json:"deactivate_at"`
// }

type UserToken struct {
	gorm.Model
	TokenType      string    `gorm:"type:varchar(64); not null;" json:"token_type"`
	Token          string    `gorm:"type:char(36); not null;" json:"token"`
	ExpiredAt      time.Time `gorm:"index" json:"expired_at"`
	UserID         uint      `json:"user_id"`
	ConsumeCounter uint      `gorm:"not null; default:0;" json:"consume_counter"`
}

type WhiteListUser struct {
	UserID    uint      `json:"user_id"`
	Token     string    `gorm:"size:512;not null;" json:"token"`
	ExpiredAt time.Time `gorm:"index" json:"expired_at"`
}

type AuthUser struct {
	gorm.Model

	Username   string `gorm:"type:varchar(64); not null;" json:"username"`
	Hostname   string `gorm:"type:varchar(64); not null;" json:"hostname"`
	Email      string `gorm:"type:varchar(128); not null;uniqueIndex:uniq_user" json:"email"`
	DeviceInfo string `gorm:"type:varchar(255); not null;" json:"device_info"`

	Password   string `gorm:"type:varchar(255); not null;" json:"password"`
	IsActived  uint   `gorm:"not null; default:0;" json:"is_actived"`
	IsVerified uint   `gorm:"not null; default:0;" json:"is_verified"`

	RegisterAt   time.Time `gorm:"autoCreateTime" json:"register_at"`
	ExpiredAt    time.Time `gorm:"index" json:"expired_at"`
	ActivateAt   time.Time `json:"activate_at"`
	DeactivateAt time.Time `json:"deactivate_at"`
}

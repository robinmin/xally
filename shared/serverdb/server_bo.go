package serverdb

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	UserID     string `gorm:"size:64;not null;" json:"uid"`
	Status     uint8  `gorm:"not null;default:0;" json:"status"`
	DeviceInfo string `gorm:"size:255;not null;unique_index:uniq_user" json:"device_info"`
	Email      string `gorm:"size:128;not null;unique_index:uniq_user" json:"email"`
	Username   string `gorm:"size:64;not null;" json:"username"`
	Token      string `gorm:"size:512;not null;" json:"token"`

	RegisterAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP()" json:"register_at"`
	ActivateAt   time.Time `json:"activate_at"`
	ExpiredAt    time.Time `gorm:"index" json:"expired_at"`
	DeactivateAt time.Time `json:"deactivate_at"`
}

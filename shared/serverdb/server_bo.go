package serverdb

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username     string    `gorm:"size:255;not null;" json:"username"`
	UserID       string    `gorm:"size:255;not null;" json:"uid"`
	Email        string    `gorm:"size:255;not null;unique" json:"email"`
	DeviceInfo   string    `gorm:"size:255;not null;unique" json:"device_info"`
	Enabled      uint8     `gorm:"not null;default:0;" json:"enabled"`
	RegisterAt   time.Time `gorm:"not null;" json:"register_at"`
	ActivateAt   time.Time `json:"activate_at"`
	DeactivateAt time.Time `json:"deactivate_at"`
}

package clientdb

import "gorm.io/gorm"

type OptionHistory struct {
	gorm.Model

	Role   string `gorm:"type:varchar(32);<-:create"`  // allow read and create
	Option string `gorm:"type:varchar(256);<-:create"` // allow read and create
}

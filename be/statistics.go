package be

import (
	"gorm.io/gorm"
)

type Statistics struct {
	gorm.Model
	UserId  uint
	Success string
}

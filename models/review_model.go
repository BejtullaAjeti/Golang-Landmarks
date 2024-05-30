package models

import (
	"github.com/jinzhu/gorm"
)

type Review struct {
	gorm.Model
	UUID       string `gorm:"not null;index" json:"uuid"`
	Name       string `gorm:"not null" json:"name"`
	Comment    string `gorm:"type:text" json:"comment"`
	Rating     int    `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	LandmarkID uint   `gorm:"foreignkey:LandmarkID" json:"landmark_id"`
}

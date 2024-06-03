package models

import (
	"time"
)

type Review struct {
	ID         uint       `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	DeviceID   string     `gorm:"not null;index" json:"device_id"`
	Name       string     `gorm:"not null" json:"name"`
	Comment    string     `gorm:"type:text" json:"comment"`
	Rating     int        `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	LandmarkID uint       `gorm:"foreignkey:LandmarkID" json:"landmark_id"`
}

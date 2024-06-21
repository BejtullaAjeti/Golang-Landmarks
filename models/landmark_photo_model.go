package models

import (
	"time"
)

type LandmarkPhoto struct {
	ID         uint       `gorm:"primary_key" json:"id"`
	LandmarkID uint       `json:"landmark_id"`
	Name       string     `json:"name"`
	Image      string     `gorm:"size:mediumtext" json:"image"` // base64 encoded image
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

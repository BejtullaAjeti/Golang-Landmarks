package models

import (
	"time"
)

type City struct {
	ID         uint       `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Name       string     `json:"name"`
	Area       float64    `json:"area"`
	Latitude   string     `json:"latitude"`
	Longitude  string     `json:"longitude"`
	Population int        `json:"population"`
	RegionID   uint       `gorm:"foreignkey:RegionID" json:"region_id"`
}

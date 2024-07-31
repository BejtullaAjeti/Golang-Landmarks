package models

import "time"

type Country struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Name      string     `json:"name"`
	Latitude  string     `json:"latitude"`
	Longitude string     `json:"longitude"`
	Regions   []Region   `gorm:"foreignkey:RegionID" json:"regions"`
}

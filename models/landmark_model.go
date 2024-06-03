package models

import (
	"time"
)

type Landmark struct {
	ID          uint       `gorm:"primary_key" json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Information string     `json:"information"`
	Description string     `gorm:"size:mediumtext" json:"description"`
	Latitude    string     `json:"latitude"`
	Longitude   string     `json:"longitude"`
	CityID      uint       `gorm:"foreignkey:CityID" json:"city_id"`
	Image       string     `gorm:"size:mediumtext" json:"image"`
}

package models

import (
	"github.com/jinzhu/gorm"
)

type City struct {
	gorm.Model
	Name       string  `json:"name"`
	Area       float64 `json:"area"`
	Latitude   string  `json:"latitude"`
	Longitude  string  `json:"longitude"`
	Population int     `json:"population"`
	RegionID   uint    `gorm:"foreignkey:RegionID" json:"region_id"`
}

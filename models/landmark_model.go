package models

import (
	"github.com/jinzhu/gorm"
)

type Landmark struct {
	gorm.Model
	Name        string `json:"name"`
	Type        string `json:"type"`
	Information string `json:"information"`
	Description string `gorm:"size:mediumtext" json:"description"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	CityID      uint   `gorm:"foreignkey:CityID" json:"city_id"`
	Image       string `gorm:"size:mediumtext" json:"image"`
}

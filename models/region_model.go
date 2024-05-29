package models

import (
	"github.com/jinzhu/gorm"
)

type Region struct {
	gorm.Model
	Name       string  `json:"name"`
	Area       float64 `json:"area"`
	Latitude   string  `json:"latitude"`
	Longitude  string  `json:"longitude"`
	Image      string  `json:"image"`
	Population int     `json:"population"`
}

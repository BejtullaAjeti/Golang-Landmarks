package models

import "time"

type GeoJSON struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	RegionID    uint       `json:"region_id"`
	GeoJSONData string     `gorm:"type:longtext" json:"geojson_data"` // Raw GeoJSON data as a string
	MiddlePoint string     `json:"middle_point"`                      // Middle point as a string
	Zoom        float64    `json:"zoom"`                              // Zoom field
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

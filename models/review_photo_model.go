package models

import (
	"time"
)

type ReviewPhoto struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	ReviewID  uint       `json:"review_id"`
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

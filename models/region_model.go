package models

import (
	"time"
)

type Region struct {
	ID          uint       `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Name        string     `json:"name"`
	Area        float64    `json:"area"`
	Coordinates string     `json:"coordinates"`
	Population  int        `json:"population"`
}

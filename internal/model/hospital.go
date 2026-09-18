package model

import "time"

type Hospital struct {
	ID uint `gorm:"primaryKey"`

	Code string `gorm:"size:50;uniqueIndex;not null"`

	Name string `gorm:"size:255;not null"`

	APIURL string `gorm:"size:500"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

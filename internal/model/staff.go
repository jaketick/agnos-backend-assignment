package model

import "time"

type Staff struct {
	ID uint `gorm:"primaryKey"`

	Username string `gorm:"size:100;not null;uniqueIndex:idx_staff_username_hospital"`

	PasswordHash string `gorm:"size:255;not null"`

	HospitalID uint `gorm:"not null;uniqueIndex:idx_staff_username_hospital"`

	Hospital Hospital `gorm:"foreignKey:HospitalID"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

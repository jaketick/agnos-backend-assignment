package model

import "time"

type Patient struct {
	ID uint `gorm:"primaryKey"`

	HospitalID uint `gorm:"not null;index;index:idx_patient_hospital_national,priority:1;index:idx_patient_hospital_passport,priority:1"`

	Hospital Hospital `gorm:"foreignKey:HospitalID"`

	PatientHN string `gorm:"size:100;index"`

	NationalID string `gorm:"size:50;index:idx_patient_hospital_national,priority:2"`

	PassportID string `gorm:"size:50;index:idx_patient_hospital_passport,priority:2"`

	FirstNameTH  string `gorm:"size:255"`
	MiddleNameTH string `gorm:"size:255"`
	LastNameTH   string `gorm:"size:255"`

	FirstNameEN  string `gorm:"size:255"`
	MiddleNameEN string `gorm:"size:255"`
	LastNameEN   string `gorm:"size:255"`

	DateOfBirth *time.Time

	PhoneNumber string `gorm:"size:50"`

	Email string `gorm:"size:255"`

	Gender string `gorm:"size:20"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

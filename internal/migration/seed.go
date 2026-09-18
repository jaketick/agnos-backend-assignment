package migration

import (
	"fmt"

	"agnos-backend-assignment/internal/config"
	"agnos-backend-assignment/internal/model"

	"gorm.io/gorm"
)

func Seed(db *gorm.DB, cfg *config.Config) error {
	hospital := model.Hospital{
		Code:   "hospital-a",
		Name:   "Hospital A",
		APIURL: cfg.HospitalAAPIURL,
	}

	err := db.
		Where("code = ?", hospital.Code).
		FirstOrCreate(&hospital).
		Error

	if err != nil {
		return fmt.Errorf("failed to seed hospital: %w", err)
	}

	return nil
}

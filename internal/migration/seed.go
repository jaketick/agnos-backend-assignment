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

	/* passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("password123"),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return fmt.Errorf("failed to hash seed staff password: %w", err)
	}

	staff := model.Staff{
		Username:     "staff01",
		PasswordHash: string(passwordHash),
		HospitalID:   hospital.ID,
	}

	err = db.
		Where("username = ? AND hospital_id = ?", staff.Username, staff.HospitalID).
		FirstOrCreate(&staff).
		Error
	if err != nil {
		return fmt.Errorf("failed to seed staff: %w", err)
	} */

	return nil
}

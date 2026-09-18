package migration

import (
	"fmt"

	"agnos-backend-assignment/internal/model"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.Hospital{},
		&model.Staff{},
		&model.Patient{},
	)

	if err != nil {
		return fmt.Errorf("failed to run database migration: %w", err)
	}

	return nil
}

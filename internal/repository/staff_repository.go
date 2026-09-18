package repository

import (
	"agnos-backend-assignment/internal/model"

	"gorm.io/gorm"
)

type StaffRepository interface {
	ExistsByUsernameAndHospitalID(
		username string,
		hospitalID uint,
	) (bool, error)

	Create(staff *model.Staff) error

	FindByUsernameAndHospitalID(
		username string,
		hospitalID uint,
	) (*model.Staff, error)
}

type staffRepository struct {
	db *gorm.DB
}

func NewStaffRepository(db *gorm.DB) StaffRepository {
	return &staffRepository{
		db: db,
	}
}

func (r *staffRepository) ExistsByUsernameAndHospitalID(
	username string,
	hospitalID uint,
) (bool, error) {

	var count int64

	err := r.db.
		Model(&model.Staff{}).
		Where(
			"username = ? AND hospital_id = ?",
			username,
			hospitalID,
		).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *staffRepository) Create(staff *model.Staff) error {
	return r.db.Create(staff).Error
}
func (r *staffRepository) FindByUsernameAndHospitalID(
	username string,
	hospitalID uint,
) (*model.Staff, error) {
	var staff model.Staff

	err := r.db.
		Where(
			"username = ? AND hospital_id = ?",
			username,
			hospitalID,
		).
		First(&staff).
		Error

	if err != nil {
		return nil, err
	}

	return &staff, nil
}

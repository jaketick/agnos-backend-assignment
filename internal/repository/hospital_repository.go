package repository

import (
	"agnos-backend-assignment/internal/model"

	"gorm.io/gorm"
)

type HospitalRepository interface {
	FindByCode(
		code string,
	) (*model.Hospital, error)

	FindByID(
		id uint,
	) (*model.Hospital, error)
}

type hospitalRepository struct {
	db *gorm.DB
}

func NewHospitalRepository(
	db *gorm.DB,
) HospitalRepository {

	return &hospitalRepository{
		db: db,
	}
}

func (r *hospitalRepository) FindByCode(
	code string,
) (*model.Hospital, error) {

	var hospital model.Hospital

	err := r.db.
		Where(
			"code = ?",
			code,
		).
		First(&hospital).
		Error

	if err != nil {
		return nil, err
	}

	return &hospital, nil
}

func (r *hospitalRepository) FindByID(
	id uint,
) (*model.Hospital, error) {

	var hospital model.Hospital

	err := r.db.
		First(
			&hospital,
			id,
		).
		Error

	if err != nil {
		return nil, err
	}

	return &hospital, nil
}

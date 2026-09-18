package repository

import (
	"errors"
	"fmt"
	"time"

	"agnos-backend-assignment/internal/model"

	"gorm.io/gorm"
)

type PatientSearchFilter struct {
	NationalID  string
	PassportID  string
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth *time.Time
	PhoneNumber string
	Email       string
}

type PatientRepository interface {
	Search(
		hospitalID uint,
		filter PatientSearchFilter,
	) ([]model.Patient, error)

	Upsert(
		patient *model.Patient,
	) error
}

type patientRepository struct {
	db *gorm.DB
}

func NewPatientRepository(
	db *gorm.DB,
) PatientRepository {

	return &patientRepository{
		db: db,
	}
}

func (r *patientRepository) Search(
	hospitalID uint,
	filter PatientSearchFilter,
) ([]model.Patient, error) {

	var patients []model.Patient

	query := r.db.
		Model(&model.Patient{}).
		Where(
			"hospital_id = ?",
			hospitalID,
		)

	if filter.NationalID != "" {
		query = query.Where(
			"national_id = ?",
			filter.NationalID,
		)
	}

	if filter.PassportID != "" {
		query = query.Where(
			"passport_id = ?",
			filter.PassportID,
		)
	}

	if filter.FirstName != "" {
		pattern := "%" + filter.FirstName + "%"

		query = query.Where(
			`(
				first_name_th ILIKE ?
				OR first_name_en ILIKE ?
			)`,
			pattern,
			pattern,
		)
	}

	if filter.MiddleName != "" {
		pattern := "%" + filter.MiddleName + "%"

		query = query.Where(
			`(
				middle_name_th ILIKE ?
				OR middle_name_en ILIKE ?
			)`,
			pattern,
			pattern,
		)
	}

	if filter.LastName != "" {
		pattern := "%" + filter.LastName + "%"

		query = query.Where(
			`(
				last_name_th ILIKE ?
				OR last_name_en ILIKE ?
			)`,
			pattern,
			pattern,
		)
	}

	if filter.DateOfBirth != nil {
		query = query.Where(
			"date_of_birth = ?",
			*filter.DateOfBirth,
		)
	}

	if filter.PhoneNumber != "" {
		query = query.Where(
			"phone_number = ?",
			filter.PhoneNumber,
		)
	}

	if filter.Email != "" {
		query = query.Where(
			"LOWER(email) = LOWER(?)",
			filter.Email,
		)
	}

	err := query.
		Order("id ASC").
		Find(&patients).
		Error

	if err != nil {
		return nil, err
	}

	return patients, nil
}

func (r *patientRepository) Upsert(
	patient *model.Patient,
) error {

	if patient.NationalID == "" &&
		patient.PassportID == "" {

		return fmt.Errorf(
			"patient identifier is required",
		)
	}

	var existing model.Patient

	query := r.db.
		Where(
			"hospital_id = ?",
			patient.HospitalID,
		)

	if patient.NationalID != "" {
		query = query.Where(
			"national_id = ?",
			patient.NationalID,
		)
	} else {
		query = query.Where(
			"passport_id = ?",
			patient.PassportID,
		)
	}

	err := query.
		First(&existing).
		Error

	if errors.Is(
		err,
		gorm.ErrRecordNotFound,
	) {
		return r.db.
			Create(patient).
			Error
	}

	if err != nil {
		return err
	}

	patient.ID = existing.ID

	return r.db.
		Model(&existing).
		Updates(
			map[string]interface{}{
				"patient_hn": patient.PatientHN,

				"national_id": patient.NationalID,
				"passport_id": patient.PassportID,

				"first_name_th":  patient.FirstNameTH,
				"middle_name_th": patient.MiddleNameTH,
				"last_name_th":   patient.LastNameTH,

				"first_name_en":  patient.FirstNameEN,
				"middle_name_en": patient.MiddleNameEN,
				"last_name_en":   patient.LastNameEN,

				"date_of_birth": patient.DateOfBirth,

				"phone_number": patient.PhoneNumber,
				"email":        patient.Email,
				"gender":       patient.Gender,
			},
		).
		Error
}

package testutil

import (
	"context"

	"agnos-backend-assignment/internal/client"
	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/model"
	"agnos-backend-assignment/internal/repository"
	"agnos-backend-assignment/internal/service"
)

type MockStaffRepository struct {
	ExistsByUsernameAndHospitalIDFunc func(
		username string,
		hospitalID uint,
	) (bool, error)

	FindByUsernameAndHospitalIDFunc func(
		username string,
		hospitalID uint,
	) (*model.Staff, error)

	CreateFunc func(
		staff *model.Staff,
	) error
}

func (m *MockStaffRepository) ExistsByUsernameAndHospitalID(
	username string,
	hospitalID uint,
) (bool, error) {

	if m.ExistsByUsernameAndHospitalIDFunc == nil {
		return false, nil
	}

	return m.ExistsByUsernameAndHospitalIDFunc(
		username,
		hospitalID,
	)
}

func (m *MockStaffRepository) FindByUsernameAndHospitalID(
	username string,
	hospitalID uint,
) (*model.Staff, error) {

	if m.FindByUsernameAndHospitalIDFunc == nil {
		return nil, nil
	}

	return m.FindByUsernameAndHospitalIDFunc(
		username,
		hospitalID,
	)
}

func (m *MockStaffRepository) Create(
	staff *model.Staff,
) error {

	if m.CreateFunc == nil {
		return nil
	}

	return m.CreateFunc(staff)
}

// --------------------------------------------------

type MockHospitalRepository struct {
	FindByCodeFunc func(
		code string,
	) (*model.Hospital, error)

	FindByIDFunc func(
		id uint,
	) (*model.Hospital, error)
}

func (m *MockHospitalRepository) FindByCode(
	code string,
) (*model.Hospital, error) {

	if m.FindByCodeFunc == nil {
		return nil, nil
	}

	return m.FindByCodeFunc(code)
}

func (m *MockHospitalRepository) FindByID(
	id uint,
) (*model.Hospital, error) {

	if m.FindByIDFunc == nil {
		return nil, nil
	}

	return m.FindByIDFunc(id)
}

// --------------------------------------------------

type MockPatientRepository struct {
	SearchFunc func(
		hospitalID uint,
		filter repository.PatientSearchFilter,
	) ([]model.Patient, error)

	UpsertFunc func(
		patient *model.Patient,
	) error
}

func (m *MockPatientRepository) Search(
	hospitalID uint,
	filter repository.PatientSearchFilter,
) ([]model.Patient, error) {

	if m.SearchFunc == nil {
		return []model.Patient{}, nil
	}

	return m.SearchFunc(
		hospitalID,
		filter,
	)
}

func (m *MockPatientRepository) Upsert(
	patient *model.Patient,
) error {

	if m.UpsertFunc == nil {
		return nil
	}

	return m.UpsertFunc(patient)
}

// --------------------------------------------------

type MockJWTService struct {
	GenerateTokenFunc func(
		staffID uint,
		hospitalID uint,
		username string,
	) (string, error)

	ValidateTokenFunc func(
		tokenString string,
	) (*service.JWTClaims, error)
}

func (m *MockJWTService) GenerateToken(
	staffID uint,
	hospitalID uint,
	username string,
) (string, error) {

	if m.GenerateTokenFunc == nil {
		return "", nil
	}

	return m.GenerateTokenFunc(
		staffID,
		hospitalID,
		username,
	)
}

func (m *MockJWTService) ValidateToken(
	tokenString string,
) (*service.JWTClaims, error) {

	if m.ValidateTokenFunc == nil {
		return nil, nil
	}

	return m.ValidateTokenFunc(
		tokenString,
	)
}

// --------------------------------------------------

type MockHospitalClient struct {
	SearchPatientFunc func(
		ctx context.Context,
		baseURL string,
		id string,
	) (*dto.HospitalPatientResponse, error)
}

func (m *MockHospitalClient) SearchPatient(
	ctx context.Context,
	baseURL string,
	id string,
) (*dto.HospitalPatientResponse, error) {

	if m.SearchPatientFunc == nil {
		return nil, nil
	}

	return m.SearchPatientFunc(
		ctx,
		baseURL,
		id,
	)
}

var _ repository.StaffRepository = (*MockStaffRepository)(nil)

var _ repository.HospitalRepository = (*MockHospitalRepository)(nil)

var _ repository.PatientRepository = (*MockPatientRepository)(nil)

var _ service.JWTService = (*MockJWTService)(nil)

var _ client.HospitalClient = (*MockHospitalClient)(nil)

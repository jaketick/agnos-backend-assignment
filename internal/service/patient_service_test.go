package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"agnos-backend-assignment/internal/client"
	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/model"
	"agnos-backend-assignment/internal/repository"
	"agnos-backend-assignment/internal/service"
	"agnos-backend-assignment/internal/testutil"
)

func TestPatientService_Search_LocalFound(
	t *testing.T,
) {

	hospitalClientCalled := false
	hospitalRepoCalled := false
	upsertCalled := false

	dateOfBirth := time.Date(
		1990,
		time.January,
		20,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			if hospitalID != 1 {
				t.Fatalf(
					"expected hospitalID 1, got %d",
					hospitalID,
				)
			}

			if filter.NationalID != "1234567890123" {
				t.Fatalf(
					"unexpected national ID: %s",
					filter.NationalID,
				)
			}

			return []model.Patient{
				{
					ID:          10,
					HospitalID:  1,
					PatientHN:   "HN0001",
					NationalID:  "1234567890123",
					FirstNameEN: "Somchai",
					LastNameEN:  "Jaidee",
					DateOfBirth: &dateOfBirth,
					PhoneNumber: "0812345678",
					Email:       "somchai@example.com",
					Gender:      "M",
				},
			}, nil
		},

		UpsertFunc: func(
			patient *model.Patient,
		) error {

			upsertCalled = true
			return nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			hospitalRepoCalled = true
			return nil, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			hospitalClientCalled = true
			return nil, nil
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	result, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			NationalID: "1234567890123",
		},
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.Count != 1 {
		t.Fatalf(
			"expected 1 patient, got %d",
			result.Count,
		)
	}

	if result.Data[0].FirstNameEN != "Somchai" {
		t.Errorf(
			"expected Somchai, got %s",
			result.Data[0].FirstNameEN,
		)
	}

	if hospitalRepoCalled {
		t.Fatal(
			"hospital repository should not be called on local hit",
		)
	}

	if hospitalClientCalled {
		t.Fatal(
			"HIS should not be called on local hit",
		)
	}

	if upsertCalled {
		t.Fatal(
			"patient should not be upserted on local hit",
		)
	}
}

func TestPatientService_Search_HISFallbackByNationalID(
	t *testing.T,
) {

	searchCount := 0
	upsertCalled := false
	hisCalled := false

	var cachedPatient *model.Patient

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			searchCount++

			if hospitalID != 1 {
				t.Fatalf(
					"expected hospitalID 1, got %d",
					hospitalID,
				)
			}

			if searchCount == 1 {
				return []model.Patient{}, nil
			}

			if cachedPatient == nil {
				t.Fatal(
					"expected patient to be cached before second search",
				)
			}

			cachedPatient.ID = 99

			return []model.Patient{
				*cachedPatient,
			}, nil
		},

		UpsertFunc: func(
			patient *model.Patient,
		) error {

			upsertCalled = true
			cachedPatient = patient

			return nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			if id != 1 {
				t.Fatalf(
					"expected hospital ID 1, got %d",
					id,
				)
			}

			return &model.Hospital{
				ID:     1,
				Code:   "hospital-a",
				APIURL: "https://hospital-a.test",
			}, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			hisCalled = true

			if baseURL != "https://hospital-a.test" {
				t.Fatalf(
					"unexpected HIS URL: %s",
					baseURL,
				)
			}

			if id != "1234567890123" {
				t.Fatalf(
					"expected national ID, got %s",
					id,
				)
			}

			return &dto.HospitalPatientResponse{
				PatientHN:   "HN0001",
				NationalID:  "1234567890123",
				FirstNameTH: "สมชาย",
				LastNameTH:  "ใจดี",
				FirstNameEN: "Somchai",
				LastNameEN:  "Jaidee",
				DateOfBirth: "1990-01-20",
				PhoneNumber: "0812345678",
				Email:       "somchai@example.com",
				Gender:      "M",
			}, nil
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	result, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			NationalID: "1234567890123",
		},
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if !hisCalled {
		t.Fatal(
			"expected HIS to be called",
		)
	}

	if !upsertCalled {
		t.Fatal(
			"expected patient to be upserted",
		)
	}

	if searchCount != 2 {
		t.Fatalf(
			"expected repository search twice, got %d",
			searchCount,
		)
	}

	if cachedPatient == nil {
		t.Fatal(
			"expected cached patient",
		)
	}

	if cachedPatient.HospitalID != 1 {
		t.Fatalf(
			"expected cached patient hospital ID 1, got %d",
			cachedPatient.HospitalID,
		)
	}

	if cachedPatient.NationalID != "1234567890123" {
		t.Errorf(
			"unexpected cached national ID: %s",
			cachedPatient.NationalID,
		)
	}

	if result.Count != 1 {
		t.Fatalf(
			"expected 1 result, got %d",
			result.Count,
		)
	}

	if result.Data[0].FirstNameEN != "Somchai" {
		t.Errorf(
			"expected Somchai, got %s",
			result.Data[0].FirstNameEN,
		)
	}
}
func TestPatientService_Search_HISFallbackByPassportID(
	t *testing.T,
) {

	searchCount := 0

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			searchCount++

			if searchCount == 1 {
				return []model.Patient{}, nil
			}

			return []model.Patient{
				{
					ID:          20,
					HospitalID:  1,
					PatientHN:   "HN0002",
					PassportID:  "AA123456",
					FirstNameEN: "John",
					LastNameEN:  "Doe",
				},
			}, nil
		},

		UpsertFunc: func(
			patient *model.Patient,
		) error {

			if patient.PassportID != "AA123456" {
				t.Fatalf(
					"expected passport AA123456, got %s",
					patient.PassportID,
				)
			}

			return nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			return &model.Hospital{
				ID:     1,
				APIURL: "https://hospital-a.test",
			}, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			if id != "AA123456" {
				t.Fatalf(
					"expected passport ID AA123456, got %s",
					id,
				)
			}

			return &dto.HospitalPatientResponse{
				PatientHN:   "HN0002",
				PassportID:  "AA123456",
				FirstNameEN: "John",
				LastNameEN:  "Doe",
			}, nil
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	result, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			PassportID: "AA123456",
		},
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.Count != 1 {
		t.Fatalf(
			"expected 1 result, got %d",
			result.Count,
		)
	}
}
func TestPatientService_Search_NameOnlyLocalMiss(
	t *testing.T,
) {

	hospitalClientCalled := false
	hospitalRepoCalled := false

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			if filter.FirstName != "Somchai" {
				t.Fatalf(
					"expected Somchai, got %s",
					filter.FirstName,
				)
			}

			return []model.Patient{}, nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			hospitalRepoCalled = true
			return nil, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			hospitalClientCalled = true
			return nil, nil
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	result, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			FirstName: "Somchai",
		},
	)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.Count != 0 {
		t.Fatalf(
			"expected no results, got %d",
			result.Count,
		)
	}

	if hospitalRepoCalled {
		t.Fatal(
			"hospital repository should not be called for name-only local miss",
		)
	}

	if hospitalClientCalled {
		t.Fatal(
			"HIS should not be called without national/passport ID",
		)
	}
}
func TestPatientService_Search_InvalidDateOfBirth(
	t *testing.T,
) {

	repositoryCalled := false
	hisCalled := false

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			repositoryCalled = true
			return nil, nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			hisCalled = true
			return nil, nil
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	_, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			DateOfBirth: "20-01-1990",
		},
	)

	if !errors.Is(
		err,
		service.ErrInvalidDateOfBirth,
	) {
		t.Fatalf(
			"expected ErrInvalidDateOfBirth, got %v",
			err,
		)
	}

	if repositoryCalled {
		t.Fatal(
			"repository should not be called for invalid DOB",
		)
	}

	if hisCalled {
		t.Fatal(
			"HIS should not be called for invalid DOB",
		)
	}
}
func TestPatientService_Search_HISPatientNotFound(
	t *testing.T,
) {

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			return []model.Patient{}, nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			return &model.Hospital{
				ID:     1,
				APIURL: "https://hospital-a.test",
			}, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			return nil, client.ErrPatientNotFound
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	result, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			NationalID: "0000000000000",
		},
	)

	if err != nil {
		t.Fatalf(
			"patient not found should not be a system error: %v",
			err,
		)
	}

	if result.Count != 0 {
		t.Fatalf(
			"expected zero patients, got %d",
			result.Count,
		)
	}

	if len(result.Data) != 0 {
		t.Fatalf(
			"expected empty data, got %d records",
			len(result.Data),
		)
	}
}
func TestPatientService_Search_HISUnavailable(
	t *testing.T,
) {

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			return []model.Patient{}, nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			return &model.Hospital{
				ID:     1,
				APIURL: "https://hospital-a.test",
			}, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			return nil,
				client.ErrHospitalUnavailable
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	_, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			NationalID: "1234567890123",
		},
	)

	if !errors.Is(
		err,
		service.ErrHospitalUnavailable,
	) {
		t.Fatalf(
			"expected ErrHospitalUnavailable, got %v",
			err,
		)
	}
}
func TestPatientService_Search_InvalidHISResponse(
	t *testing.T,
) {

	patientRepo := &testutil.MockPatientRepository{
		SearchFunc: func(
			hospitalID uint,
			filter repository.PatientSearchFilter,
		) ([]model.Patient, error) {

			return []model.Patient{}, nil
		},
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByIDFunc: func(
			id uint,
		) (*model.Hospital, error) {

			return &model.Hospital{
				ID:     1,
				APIURL: "https://hospital-a.test",
			}, nil
		},
	}

	hospitalClient := &testutil.MockHospitalClient{
		SearchPatientFunc: func(
			ctx context.Context,
			baseURL string,
			id string,
		) (*dto.HospitalPatientResponse, error) {

			return nil,
				client.ErrInvalidHospitalReply
		},
	}

	patientService := service.NewPatientService(
		patientRepo,
		hospitalRepo,
		hospitalClient,
	)

	_, err := patientService.Search(
		context.Background(),
		1,
		dto.PatientSearchRequest{
			NationalID: "1234567890123",
		},
	)

	if !errors.Is(
		err,
		service.ErrHospitalUnavailable,
	) {
		t.Fatalf(
			"expected ErrHospitalUnavailable, got %v",
			err,
		)
	}
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"agnos-backend-assignment/internal/client"
	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/model"
	"agnos-backend-assignment/internal/repository"
)

type PatientService interface {
	Search(
		ctx context.Context,
		hospitalID uint,
		req dto.PatientSearchRequest,
	) (*dto.PatientSearchResponse, error)
}

type patientService struct {
	patientRepo    repository.PatientRepository
	hospitalRepo   repository.HospitalRepository
	hospitalClient client.HospitalClient
}

func NewPatientService(
	patientRepo repository.PatientRepository,
	hospitalRepo repository.HospitalRepository,
	hospitalClient client.HospitalClient,
) PatientService {

	return &patientService{
		patientRepo:    patientRepo,
		hospitalRepo:   hospitalRepo,
		hospitalClient: hospitalClient,
	}
}

func (s *patientService) Search(
	ctx context.Context,
	hospitalID uint,
	req dto.PatientSearchRequest,
) (*dto.PatientSearchResponse, error) {

	filter := repository.PatientSearchFilter{
		NationalID: strings.TrimSpace(
			req.NationalID,
		),

		PassportID: strings.TrimSpace(
			req.PassportID,
		),

		FirstName: strings.TrimSpace(
			req.FirstName,
		),

		MiddleName: strings.TrimSpace(
			req.MiddleName,
		),

		LastName: strings.TrimSpace(
			req.LastName,
		),

		PhoneNumber: strings.TrimSpace(
			req.PhoneNumber,
		),

		Email: strings.TrimSpace(
			req.Email,
		),
	}

	dateOfBirthString := strings.TrimSpace(
		req.DateOfBirth,
	)

	if dateOfBirthString != "" {
		dateOfBirth, err := time.Parse(
			"2006-01-02",
			dateOfBirthString,
		)

		if err != nil {
			return nil, ErrInvalidDateOfBirth
		}

		filter.DateOfBirth = &dateOfBirth
	}

	// ค้นหาในฐานข้อมูลของระบบเราก่อน
	// เพื่อไม่ต้องเรียก HIS ทุกครั้ง
	patients, err := s.patientRepo.Search(
		hospitalID,
		filter,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to search patients: %w",
			err,
		)
	}

	// ถ้าพบผู้ป่วยในฐานข้อมูลแล้ว
	// ส่งผลลัพธ์กลับได้เลย โดยไม่ต้องเรียก HIS
	if len(patients) > 0 {
		return buildPatientSearchResponse(
			patients,
		), nil
	}

	// HIS ที่โจทย์ให้มารองรับการค้นหาด้วย
	// เลขบัตรประชาชน หรือเลขหนังสือเดินทางเท่านั้น
	searchID := ""

	if filter.NationalID != "" {
		searchID = filter.NationalID
	} else if filter.PassportID != "" {
		searchID = filter.PassportID
	}

	// ถ้าค้นหาด้วยชื่อ วันเกิด เบอร์โทร หรืออีเมลอย่างเดียว
	// จะเรียก HIS ไม่ได้ จึงคืนผลจากฐานข้อมูลในระบบ
	if searchID == "" {
		return buildPatientSearchResponse(
			patients,
		), nil
	}

	hospital, err := s.hospitalRepo.FindByID(
		hospitalID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to find hospital: %w",
			err,
		)
	}

	hospitalPatient, err :=
		s.hospitalClient.SearchPatient(
			ctx,
			hospital.APIURL,
			searchID,
		)

	if err != nil {
		switch {
		case errors.Is(
			err,
			client.ErrPatientNotFound,
		):
			return buildPatientSearchResponse(
				nil,
			), nil

		case errors.Is(
			err,
			client.ErrHospitalUnavailable,
		):
			return nil, ErrHospitalUnavailable

		case errors.Is(
			err,
			client.ErrInvalidHospitalReply,
		):
			return nil, ErrHospitalUnavailable

		default:
			return nil, fmt.Errorf(
				"failed to search hospital patient: %w",
				err,
			)
		}
	}

	patient, err := hospitalPatientToModel(
		hospitalID,
		*hospitalPatient,
	)

	if err != nil {
		return nil, ErrHospitalUnavailable
	}

	if err := s.patientRepo.Upsert(
		patient,
	); err != nil {

		return nil, fmt.Errorf(
			"failed to save hospital patient: %w",
			err,
		)
	}

	/*
		หลังบันทึกข้อมูลจาก HIS แล้ว ให้ค้นหาในฐานข้อมูลอีกครั้ง
		โดยใช้ filter เดิมที่ผู้ใช้ส่งมา

		ตัวอย่าง:

		national_id=123
		first_name=John

		HIS ค้นหาได้ด้วย national_id=123 เท่านั้น

		เมื่อบันทึกผู้ป่วยไว้ในระบบแล้ว เรายังค้นหาซ้ำด้วย
		first_name=John เพื่อให้ผลลัพธ์ตรงตามเงื่อนไขที่ผู้ใช้ส่งมา
	*/
	patients, err = s.patientRepo.Search(
		hospitalID,
		filter,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to search cached patient: %w",
			err,
		)
	}

	return buildPatientSearchResponse(
		patients,
	), nil
}

func hospitalPatientToModel(
	hospitalID uint,
	p dto.HospitalPatientResponse,
) (*model.Patient, error) {

	if strings.TrimSpace(p.NationalID) == "" &&
		strings.TrimSpace(p.PassportID) == "" {

		return nil, fmt.Errorf(
			"hospital patient identifier is missing",
		)
	}

	var dateOfBirth *time.Time

	rawDateOfBirth := strings.TrimSpace(
		p.DateOfBirth,
	)

	if rawDateOfBirth != "" {
		parsedDateOfBirth, err := time.Parse(
			"2006-01-02",
			rawDateOfBirth,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"invalid hospital date_of_birth: %w",
				err,
			)
		}

		dateOfBirth = &parsedDateOfBirth
	}

	return &model.Patient{
		HospitalID: hospitalID,

		PatientHN: p.PatientHN,

		NationalID: p.NationalID,
		PassportID: p.PassportID,

		FirstNameTH:  p.FirstNameTH,
		MiddleNameTH: p.MiddleNameTH,
		LastNameTH:   p.LastNameTH,

		FirstNameEN:  p.FirstNameEN,
		MiddleNameEN: p.MiddleNameEN,
		LastNameEN:   p.LastNameEN,

		DateOfBirth: dateOfBirth,

		PhoneNumber: p.PhoneNumber,
		Email:       p.Email,
		Gender:      p.Gender,
	}, nil
}

func buildPatientSearchResponse(
	patients []model.Patient,
) *dto.PatientSearchResponse {

	result := make(
		[]dto.PatientResponse,
		0,
		len(patients),
	)

	for _, patient := range patients {

		dateOfBirth := ""

		if patient.DateOfBirth != nil {
			dateOfBirth =
				patient.DateOfBirth.Format(
					"2006-01-02",
				)
		}

		result = append(
			result,
			dto.PatientResponse{
				ID: patient.ID,

				PatientHN: patient.PatientHN,

				NationalID: patient.NationalID,
				PassportID: patient.PassportID,

				FirstNameTH:  patient.FirstNameTH,
				MiddleNameTH: patient.MiddleNameTH,
				LastNameTH:   patient.LastNameTH,

				FirstNameEN:  patient.FirstNameEN,
				MiddleNameEN: patient.MiddleNameEN,
				LastNameEN:   patient.LastNameEN,

				DateOfBirth: dateOfBirth,

				PhoneNumber: patient.PhoneNumber,
				Email:       patient.Email,
				Gender:      patient.Gender,
			},
		)
	}

	return &dto.PatientSearchResponse{
		Count: len(result),
		Data:  result,
	}
}

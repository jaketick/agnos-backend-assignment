package service

import (
	"errors"
	"fmt"
	"strings"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/model"
	"agnos-backend-assignment/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type StaffService interface {
	Create(
		req dto.CreateStaffRequest,
	) (*dto.CreateStaffResponse, error)
}

type staffService struct {
	staffRepo    repository.StaffRepository
	hospitalRepo repository.HospitalRepository
}

func NewStaffService(
	staffRepo repository.StaffRepository,
	hospitalRepo repository.HospitalRepository,
) StaffService {
	return &staffService{
		staffRepo:    staffRepo,
		hospitalRepo: hospitalRepo,
	}
}

func (s *staffService) Create(
	req dto.CreateStaffRequest,
) (*dto.CreateStaffResponse, error) {

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	hospitalCode := strings.TrimSpace(req.Hospital)

	if username == "" || password == "" || hospitalCode == "" {
		return nil, ErrInvalidStaffInput
	}

	hospital, err := s.hospitalRepo.FindByCode(hospitalCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHospitalNotFound
		}

		return nil, fmt.Errorf(
			"failed to find hospital: %w",
			err,
		)
	}

	exists, err := s.staffRepo.ExistsByUsernameAndHospitalID(
		username,
		hospital.ID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to check existing staff: %w",
			err,
		)
	}

	if exists {
		return nil, ErrStaffAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to hash password: %w",
			err,
		)
	}

	staff := model.Staff{
		Username:     username,
		PasswordHash: string(passwordHash),
		HospitalID:   hospital.ID,
	}

	if err := s.staffRepo.Create(&staff); err != nil {
		return nil, fmt.Errorf(
			"failed to create staff: %w",
			err,
		)
	}

	return &dto.CreateStaffResponse{
		ID:       staff.ID,
		Username: staff.Username,
		Hospital: hospital.Code,
	}, nil
}

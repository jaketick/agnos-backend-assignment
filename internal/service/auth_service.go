package service

import (
	"errors"
	"fmt"
	"strings"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(
		req dto.LoginRequest,
	) (*dto.LoginResponse, error)
}

type authService struct {
	staffRepo    repository.StaffRepository
	hospitalRepo repository.HospitalRepository
	jwtService   JWTService
}

func NewAuthService(
	staffRepo repository.StaffRepository,
	hospitalRepo repository.HospitalRepository,
	jwtService JWTService,
) AuthService {

	return &authService{
		staffRepo:    staffRepo,
		hospitalRepo: hospitalRepo,
		jwtService:   jwtService,
	}
}

func (s *authService) Login(
	req dto.LoginRequest,
) (*dto.LoginResponse, error) {

	username := strings.TrimSpace(req.Username)
	hospitalCode := strings.TrimSpace(req.Hospital)

	if username == "" ||
		req.Password == "" ||
		hospitalCode == "" {

		return nil, ErrInvalidCredentials
	}

	hospital, err := s.hospitalRepo.FindByCode(
		hospitalCode,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf(
			"failed to find hospital: %w",
			err,
		)
	}

	staff, err := s.staffRepo.FindByUsernameAndHospitalID(
		username,
		hospital.ID,
	)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, fmt.Errorf(
			"failed to find staff: %w",
			err,
		)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(staff.PasswordHash),
		[]byte(req.Password),
	)

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwtService.GenerateToken(
		staff.ID,
		staff.HospitalID,
		staff.Username,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to generate access token: %w",
			err,
		)
	}

	return &dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	}, nil
}

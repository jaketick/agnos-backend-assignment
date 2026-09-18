package service_test

import (
	"errors"
	"testing"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/model"
	"agnos-backend-assignment/internal/service"
	"agnos-backend-assignment/internal/testutil"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestAuthService_Login_Success(t *testing.T) {
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("password123"),
		bcrypt.DefaultCost,
	)

	if err != nil {
		t.Fatalf(
			"failed to prepare password hash: %v",
			err,
		)
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByCodeFunc: func(
			code string,
		) (*model.Hospital, error) {

			if code != "hospital-a" {
				t.Fatalf(
					"expected hospital-a, got %s",
					code,
				)
			}

			return &model.Hospital{
				ID:   1,
				Code: "hospital-a",
			}, nil
		},
	}

	staffRepo := &testutil.MockStaffRepository{
		FindByUsernameAndHospitalIDFunc: func(
			username string,
			hospitalID uint,
		) (*model.Staff, error) {

			if username != "staff01" {
				t.Fatalf(
					"expected staff01, got %s",
					username,
				)
			}

			if hospitalID != 1 {
				t.Fatalf(
					"expected hospitalID 1, got %d",
					hospitalID,
				)
			}

			return &model.Staff{
				ID:           10,
				Username:     "staff01",
				PasswordHash: string(passwordHash),
				HospitalID:   1,
			}, nil
		},
	}

	var receivedStaffID uint
	var receivedHospitalID uint
	var receivedUsername string

	jwtService := &testutil.MockJWTService{
		GenerateTokenFunc: func(
			staffID uint,
			hospitalID uint,
			username string,
		) (string, error) {

			receivedStaffID = staffID
			receivedHospitalID = hospitalID
			receivedUsername = username

			return "fake-access-token", nil
		},
	}

	authService := service.NewAuthService(
		staffRepo,
		hospitalRepo,
		jwtService,
	)

	req := dto.LoginRequest{
		Username: "staff01",
		Password: "password123",
		Hospital: "hospital-a",
	}

	result, err := authService.Login(req)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected login result, got nil")
	}

	if result.AccessToken != "fake-access-token" {
		t.Errorf(
			"expected fake-access-token, got %s",
			result.AccessToken,
		)
	}

	if result.TokenType != "Bearer" {
		t.Errorf(
			"expected Bearer, got %s",
			result.TokenType,
		)
	}

	if receivedStaffID != 10 {
		t.Errorf(
			"expected staffID 10, got %d",
			receivedStaffID,
		)
	}

	if receivedHospitalID != 1 {
		t.Errorf(
			"expected hospitalID 1, got %d",
			receivedHospitalID,
		)
	}

	if receivedUsername != "staff01" {
		t.Errorf(
			"expected staff01, got %s",
			receivedUsername,
		)
	}
}

func TestAuthService_Login_InvalidPassword(
	t *testing.T,
) {

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte("password123"),
		bcrypt.DefaultCost,
	)

	if err != nil {
		t.Fatalf(
			"failed to prepare password hash: %v",
			err,
		)
	}

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByCodeFunc: func(
			code string,
		) (*model.Hospital, error) {

			return &model.Hospital{
				ID:   1,
				Code: "hospital-a",
			}, nil
		},
	}

	staffRepo := &testutil.MockStaffRepository{
		FindByUsernameAndHospitalIDFunc: func(
			username string,
			hospitalID uint,
		) (*model.Staff, error) {

			return &model.Staff{
				ID:           10,
				Username:     "staff01",
				PasswordHash: string(passwordHash),
				HospitalID:   1,
			}, nil
		},
	}

	tokenGenerated := false

	jwtService := &testutil.MockJWTService{
		GenerateTokenFunc: func(
			staffID uint,
			hospitalID uint,
			username string,
		) (string, error) {

			tokenGenerated = true

			return "should-not-be-generated", nil
		},
	}

	authService := service.NewAuthService(
		staffRepo,
		hospitalRepo,
		jwtService,
	)

	req := dto.LoginRequest{
		Username: "staff01",
		Password: "wrongpassword",
		Hospital: "hospital-a",
	}

	_, err = authService.Login(req)

	if !errors.Is(
		err,
		service.ErrInvalidCredentials,
	) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if tokenGenerated {
		t.Fatal(
			"JWT must not be generated for invalid password",
		)
	}
}

func TestAuthService_Login_StaffNotFound(
	t *testing.T,
) {

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByCodeFunc: func(
			code string,
		) (*model.Hospital, error) {

			return &model.Hospital{
				ID:   1,
				Code: "hospital-a",
			}, nil
		},
	}

	staffRepo := &testutil.MockStaffRepository{
		FindByUsernameAndHospitalIDFunc: func(
			username string,
			hospitalID uint,
		) (*model.Staff, error) {

			return nil, gorm.ErrRecordNotFound
		},
	}

	tokenGenerated := false

	jwtService := &testutil.MockJWTService{
		GenerateTokenFunc: func(
			staffID uint,
			hospitalID uint,
			username string,
		) (string, error) {

			tokenGenerated = true

			return "", nil
		},
	}

	authService := service.NewAuthService(
		staffRepo,
		hospitalRepo,
		jwtService,
	)

	req := dto.LoginRequest{
		Username: "nobody",
		Password: "password123",
		Hospital: "hospital-a",
	}

	_, err := authService.Login(req)

	if !errors.Is(
		err,
		service.ErrInvalidCredentials,
	) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if tokenGenerated {
		t.Fatal(
			"JWT must not be generated when staff is not found",
		)
	}
}
func TestAuthService_Login_HospitalNotFound(
	t *testing.T,
) {

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByCodeFunc: func(
			code string,
		) (*model.Hospital, error) {

			return nil, gorm.ErrRecordNotFound
		},
	}

	staffLookupCalled := false

	staffRepo := &testutil.MockStaffRepository{
		FindByUsernameAndHospitalIDFunc: func(
			username string,
			hospitalID uint,
		) (*model.Staff, error) {

			staffLookupCalled = true

			return nil, nil
		},
	}

	jwtService := &testutil.MockJWTService{}

	authService := service.NewAuthService(
		staffRepo,
		hospitalRepo,
		jwtService,
	)

	req := dto.LoginRequest{
		Username: "staff01",
		Password: "password123",
		Hospital: "hospital-x",
	}

	_, err := authService.Login(req)

	if !errors.Is(
		err,
		service.ErrInvalidCredentials,
	) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if staffLookupCalled {
		t.Fatal(
			"staff repository must not be queried when hospital is not found",
		)
	}
}
func TestAuthService_Login_InvalidInput(
	t *testing.T,
) {

	tests := []struct {
		name string
		req  dto.LoginRequest
	}{
		{
			name: "missing username",
			req: dto.LoginRequest{
				Password: "password123",
				Hospital: "hospital-a",
			},
		},
		{
			name: "missing password",
			req: dto.LoginRequest{
				Username: "staff01",
				Hospital: "hospital-a",
			},
		},
		{
			name: "missing hospital",
			req: dto.LoginRequest{
				Username: "staff01",
				Password: "password123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {

				repositoryCalled := false

				hospitalRepo :=
					&testutil.MockHospitalRepository{
						FindByCodeFunc: func(
							code string,
						) (*model.Hospital, error) {

							repositoryCalled = true

							return nil, nil
						},
					}

				staffRepo :=
					&testutil.MockStaffRepository{
						FindByUsernameAndHospitalIDFunc: func(
							username string,
							hospitalID uint,
						) (*model.Staff, error) {

							repositoryCalled = true

							return nil, nil
						},
					}

				jwtService :=
					&testutil.MockJWTService{}

				authService :=
					service.NewAuthService(
						staffRepo,
						hospitalRepo,
						jwtService,
					)

				_, err := authService.Login(
					tt.req,
				)

				if !errors.Is(
					err,
					service.ErrInvalidCredentials,
				) {
					t.Fatalf(
						"expected ErrInvalidCredentials, got %v",
						err,
					)
				}

				if repositoryCalled {
					t.Fatal(
						"repositories should not be queried for invalid input",
					)
				}
			},
		)
	}
}

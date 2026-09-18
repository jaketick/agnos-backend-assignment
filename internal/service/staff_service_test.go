package service_test

import (
	"testing"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/model"
	"agnos-backend-assignment/internal/service"
	"agnos-backend-assignment/internal/testutil"

	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestStaffService_Create_Success(t *testing.T) {
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
				Name: "Hospital A",
			}, nil
		},
	}

	staffRepo := &testutil.MockStaffRepository{
		ExistsByUsernameAndHospitalIDFunc: func(
			username string,
			hospitalID uint,
		) (bool, error) {

			if username != "staff01" {
				t.Fatalf(
					"expected username staff01, got %s",
					username,
				)
			}

			if hospitalID != 1 {
				t.Fatalf(
					"expected hospitalID 1, got %d",
					hospitalID,
				)
			}

			return false, nil
		},
	}

	var createdStaff *model.Staff

	staffRepo.CreateFunc = func(
		staff *model.Staff,
	) error {

		createdStaff = staff

		// จำลอง behavior หลัง INSERT ของ GORM
		staff.ID = 100

		return nil
	}

	staffService := service.NewStaffService(
		staffRepo,
		hospitalRepo,
	)

	req := dto.CreateStaffRequest{
		Username: "staff01",
		Password: "password123",
		Hospital: "hospital-a",
	}

	result, err := staffService.Create(req)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.ID != 100 {
		t.Errorf(
			"expected ID 100, got %d",
			result.ID,
		)
	}

	if result.Username != "staff01" {
		t.Errorf(
			"expected username staff01, got %s",
			result.Username,
		)
	}

	if result.Hospital != "hospital-a" {
		t.Errorf(
			"expected hospital hospital-a, got %s",
			result.Hospital,
		)
	}

	if createdStaff == nil {
		t.Fatal("expected staff to be created")
	}

	if createdStaff.HospitalID != 1 {
		t.Errorf(
			"expected hospital ID 1, got %d",
			createdStaff.HospitalID,
		)
	}

	if createdStaff.PasswordHash == "password123" {
		t.Fatal(
			"password must not be stored as plain text",
		)
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(createdStaff.PasswordHash),
		[]byte("password123"),
	)

	if err != nil {
		t.Fatalf(
			"generated password hash does not match password: %v",
			err,
		)
	}
}
func TestStaffService_Create_HospitalNotFound(
	t *testing.T,
) {

	hospitalRepo := &testutil.MockHospitalRepository{
		FindByCodeFunc: func(
			code string,
		) (*model.Hospital, error) {

			return nil, gorm.ErrRecordNotFound
		},
	}

	createCalled := false

	staffRepo := &testutil.MockStaffRepository{
		CreateFunc: func(
			staff *model.Staff,
		) error {

			createCalled = true

			return nil
		},
	}

	staffService := service.NewStaffService(
		staffRepo,
		hospitalRepo,
	)

	req := dto.CreateStaffRequest{
		Username: "staff01",
		Password: "password123",
		Hospital: "hospital-x",
	}

	_, err := staffService.Create(req)

	if !errors.Is(
		err,
		service.ErrHospitalNotFound,
	) {
		t.Fatalf(
			"expected ErrHospitalNotFound, got %v",
			err,
		)
	}

	if createCalled {
		t.Fatal(
			"staff must not be created when hospital is not found",
		)
	}
}
func TestStaffService_Create_StaffAlreadyExists(
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

	createCalled := false

	staffRepo := &testutil.MockStaffRepository{
		ExistsByUsernameAndHospitalIDFunc: func(
			username string,
			hospitalID uint,
		) (bool, error) {

			return true, nil
		},

		CreateFunc: func(
			staff *model.Staff,
		) error {

			createCalled = true

			return nil
		},
	}

	staffService := service.NewStaffService(
		staffRepo,
		hospitalRepo,
	)

	req := dto.CreateStaffRequest{
		Username: "staff01",
		Password: "password123",
		Hospital: "hospital-a",
	}

	_, err := staffService.Create(req)

	if !errors.Is(
		err,
		service.ErrStaffAlreadyExists,
	) {
		t.Fatalf(
			"expected ErrStaffAlreadyExists, got %v",
			err,
		)
	}

	if createCalled {
		t.Fatal(
			"duplicate staff must not be created",
		)
	}
}
func TestStaffService_Create_InvalidInput(
	t *testing.T,
) {

	tests := []struct {
		name string
		req  dto.CreateStaffRequest
	}{
		{
			name: "missing username",
			req: dto.CreateStaffRequest{
				Password: "password123",
				Hospital: "hospital-a",
			},
		},
		{
			name: "missing password",
			req: dto.CreateStaffRequest{
				Username: "staff01",
				Hospital: "hospital-a",
			},
		},
		{
			name: "missing hospital",
			req: dto.CreateStaffRequest{
				Username: "staff01",
				Password: "password123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {

				hospitalCalled := false
				createCalled := false

				hospitalRepo :=
					&testutil.MockHospitalRepository{
						FindByCodeFunc: func(
							code string,
						) (*model.Hospital, error) {

							hospitalCalled = true

							return nil, nil
						},
					}

				staffRepo :=
					&testutil.MockStaffRepository{
						CreateFunc: func(
							staff *model.Staff,
						) error {

							createCalled = true

							return nil
						},
					}

				staffService :=
					service.NewStaffService(
						staffRepo,
						hospitalRepo,
					)

				_, err := staffService.Create(
					tt.req,
				)

				if !errors.Is(
					err,
					service.ErrInvalidStaffInput,
				) {
					t.Fatalf(
						"expected ErrInvalidStaffInput, got %v",
						err,
					)
				}

				if hospitalCalled {
					t.Fatal(
						"hospital repository should not be called for invalid input",
					)
				}

				if createCalled {
					t.Fatal(
						"staff should not be created for invalid input",
					)
				}
			},
		)
	}
}

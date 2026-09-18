package service_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"agnos-backend-assignment/internal/service"
)

func TestJWTService_ValidToken(t *testing.T) {
	jwtService := service.NewJWTService(
		"test-secret",
		time.Hour,
	)

	token, err := jwtService.GenerateToken(
		10,
		5,
		"staff01",
	)

	if err != nil {
		t.Fatalf(
			"expected no error generating token, got %v",
			err,
		)
	}

	if token == "" {
		t.Fatal("expected token, got empty string")
	}

	claims, err := jwtService.ValidateToken(token)

	if err != nil {
		t.Fatalf(
			"expected valid token, got error %v",
			err,
		)
	}

	if claims.StaffID != 10 {
		t.Errorf(
			"expected staff ID 10, got %d",
			claims.StaffID,
		)
	}

	if claims.HospitalID != 5 {
		t.Errorf(
			"expected hospital ID 5, got %d",
			claims.HospitalID,
		)
	}

	if claims.Username != "staff01" {
		t.Errorf(
			"expected username staff01, got %s",
			claims.Username,
		)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected expiration claim")
	}
}

func TestJWTService_WrongSecret(t *testing.T) {
	generator := service.NewJWTService(
		"secret-a",
		time.Hour,
	)

	validator := service.NewJWTService(
		"secret-b",
		time.Hour,
	)

	token, err := generator.GenerateToken(
		10,
		5,
		"staff01",
	)

	if err != nil {
		t.Fatalf(
			"failed to generate token: %v",
			err,
		)
	}

	_, err = validator.ValidateToken(token)

	if !errors.Is(
		err,
		service.ErrInvalidToken,
	) {
		t.Fatalf(
			"expected ErrInvalidToken, got %v",
			err,
		)
	}
}

func TestJWTService_MalformedToken(t *testing.T) {
	jwtService := service.NewJWTService(
		"test-secret",
		time.Hour,
	)

	_, err := jwtService.ValidateToken(
		"this-is-not-a-jwt",
	)

	if !errors.Is(
		err,
		service.ErrInvalidToken,
	) {
		t.Fatalf(
			"expected ErrInvalidToken, got %v",
			err,
		)
	}
}

func TestJWTService_ExpiredToken(t *testing.T) {
	jwtService := service.NewJWTService(
		"test-secret",
		-1*time.Second,
	)

	token, err := jwtService.GenerateToken(
		10,
		5,
		"staff01",
	)

	if err != nil {
		t.Fatalf(
			"failed to generate token: %v",
			err,
		)
	}

	_, err = jwtService.ValidateToken(token)

	if !errors.Is(
		err,
		service.ErrInvalidToken,
	) {
		t.Fatalf(
			"expected ErrInvalidToken for expired token, got %v",
			err,
		)
	}
}
func TestJWTService_TamperedToken(t *testing.T) {
	jwtService := service.NewJWTService(
		"test-secret",
		time.Hour,
	)

	token, err := jwtService.GenerateToken(
		10,
		5,
		"staff01",
	)

	if err != nil {
		t.Fatalf(
			"failed to generate token: %v",
			err,
		)
	}

	parts := strings.Split(token, ".")

	if len(parts) != 3 {
		t.Fatalf(
			"expected JWT with 3 parts, got %d",
			len(parts),
		)
	}

	// เปลี่ยน signature โดยเจตนา
	if parts[2][0] == 'a' {
		parts[2] = "b" + parts[2][1:]
	} else {
		parts[2] = "a" + parts[2][1:]
	}

	tamperedToken := strings.Join(
		parts,
		".",
	)

	_, err = jwtService.ValidateToken(
		tamperedToken,
	)

	if !errors.Is(
		err,
		service.ErrInvalidToken,
	) {
		t.Fatalf(
			"expected ErrInvalidToken for tampered token, got %v",
			err,
		)
	}
}

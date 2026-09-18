package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	StaffID    uint   `json:"staff_id"`
	HospitalID uint   `json:"hospital_id"`
	Username   string `json:"username"`

	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateToken(
		staffID uint,
		hospitalID uint,
		username string,
	) (string, error)

	ValidateToken(
		tokenString string,
	) (*JWTClaims, error)
}

type jwtService struct {
	secret        []byte
	tokenDuration time.Duration // ระยะเวลาที่โทเค็น JWT มีผลบังคับใช้ (เช่น 24 * time.Hour)
}

func NewJWTService(secret string, tokenDuration time.Duration) JWTService {
	return &jwtService{
		secret:        []byte(secret),
		tokenDuration: tokenDuration,
	}
}

func (s *jwtService) GenerateToken(
	staffID uint,
	hospitalID uint,
	username string,
) (string, error) {

	now := time.Now()

	claims := JWTClaims{
		StaffID:    staffID,
		HospitalID: hospitalID,
		Username:   username,

		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(now),

			ExpiresAt: jwt.NewNumericDate(
				now.Add(s.tokenDuration),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := token.SignedString(s.secret)

	if err != nil {
		return "", fmt.Errorf(
			"failed to sign JWT token: %w",
			err,
		)
	}

	return signedToken, nil
}

func (s *jwtService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, ErrInvalidToken
			}

			return s.secret, nil
		},
		jwt.WithValidMethods(
			[]string{
				jwt.SigningMethodHS256.Alg(),
			},
		),
	)

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)

	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

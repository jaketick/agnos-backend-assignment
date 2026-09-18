package service

import "errors"

var (
	ErrInvalidStaffInput = errors.New(
		"invalid staff input",
	)

	ErrHospitalNotFound = errors.New(
		"hospital not found",
	)

	ErrStaffAlreadyExists = errors.New(
		"staff already exists",
	)

	ErrInvalidCredentials = errors.New(
		"invalid credentials",
	)

	ErrInvalidToken = errors.New(
		"invalid token",
	)

	ErrInvalidDateOfBirth = errors.New(
		"invalid date of birth",
	)

	ErrHospitalUnavailable = errors.New(
		"hospital service unavailable",
	)
)

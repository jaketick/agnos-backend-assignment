package client

import "errors"

var (
	ErrPatientNotFound      = errors.New("patient not found")
	ErrHospitalUnavailable  = errors.New("hospital service unavailable")
	ErrInvalidHospitalReply = errors.New("invalid hospital response")
)

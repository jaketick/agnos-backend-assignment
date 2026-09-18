package client_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agnos-backend-assignment/internal/client"
)

func TestHospitalClient_SearchPatient_Success(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				if r.Method != http.MethodGet {
					t.Fatalf(
						"expected GET, got %s",
						r.Method,
					)
				}

				if r.URL.Path !=
					"/patient/search/1234567890123" {

					t.Fatalf(
						"unexpected path: %s",
						r.URL.Path,
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(
					[]byte(`{
						"first_name_th": "สมชาย",
						"middle_name_th": "",
						"last_name_th": "ใจดี",
						"first_name_en": "Somchai",
						"middle_name_en": "",
						"last_name_en": "Jaidee",
						"date_of_birth": "1990-01-20",
						"patient_hn": "HN0001",
						"national_id": "1234567890123",
						"passport_id": "AA123456",
						"phone_number": "0812345678",
						"email": "somchai@example.com",
						"gender": "M"
					}`),
				)
			},
		),
	)

	defer server.Close()
	hospitalClient :=
		client.NewHospitalClient(
			time.Second,
		)
	result, err :=
		hospitalClient.SearchPatient(
			context.Background(),
			server.URL,
			"1234567890123",
		)
	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result == nil {
		t.Fatal(
			"expected patient, got nil",
		)
	}

	if result.NationalID != "1234567890123" {
		t.Errorf(
			"unexpected national ID: %s",
			result.NationalID,
		)
	}

	if result.FirstNameEN != "Somchai" {
		t.Errorf(
			"expected Somchai, got %s",
			result.FirstNameEN,
		)
	}

	if result.PatientHN != "HN0001" {
		t.Errorf(
			"expected HN0001, got %s",
			result.PatientHN,
		)
	}
}

func TestHospitalClient_SearchPatient_NotFound(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				w.WriteHeader(
					http.StatusNotFound,
				)
			},
		),
	)

	defer server.Close()

	hospitalClient :=
		client.NewHospitalClient(
			time.Second,
		)

	_, err :=
		hospitalClient.SearchPatient(
			context.Background(),
			server.URL,
			"0000000000000",
		)

	if !errors.Is(
		err,
		client.ErrPatientNotFound,
	) {
		t.Fatalf(
			"expected ErrPatientNotFound, got %v",
			err,
		)
	}
}
func TestHospitalClient_SearchPatient_ServerError(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				w.WriteHeader(
					http.StatusInternalServerError,
				)
			},
		),
	)

	defer server.Close()

	hospitalClient :=
		client.NewHospitalClient(
			time.Second,
		)

	_, err :=
		hospitalClient.SearchPatient(
			context.Background(),
			server.URL,
			"1234567890123",
		)

	if !errors.Is(
		err,
		client.ErrHospitalUnavailable,
	) {
		t.Fatalf(
			"expected ErrHospitalUnavailable, got %v",
			err,
		)
	}
}
func TestHospitalClient_SearchPatient_Hospital5xx(
	t *testing.T,
) {

	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "500 internal server error",
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "502 bad gateway",
			statusCode: http.StatusBadGateway,
		},
		{
			name:       "503 service unavailable",
			statusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {

				server := httptest.NewServer(
					http.HandlerFunc(
						func(
							w http.ResponseWriter,
							r *http.Request,
						) {

							w.WriteHeader(
								tt.statusCode,
							)
						},
					),
				)

				defer server.Close()

				hospitalClient :=
					client.NewHospitalClient(
						time.Second,
					)

				_, err :=
					hospitalClient.SearchPatient(
						context.Background(),
						server.URL,
						"1234567890123",
					)

				if !errors.Is(
					err,
					client.ErrHospitalUnavailable,
				) {
					t.Fatalf(
						"expected ErrHospitalUnavailable, got %v",
						err,
					)
				}
			},
		)
	}
}
func TestHospitalClient_SearchPatient_InvalidJSON(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(
					[]byte(
						`{this-is-invalid-json`,
					),
				)
			},
		),
	)

	defer server.Close()

	hospitalClient :=
		client.NewHospitalClient(
			time.Second,
		)

	_, err :=
		hospitalClient.SearchPatient(
			context.Background(),
			server.URL,
			"1234567890123",
		)

	if !errors.Is(
		err,
		client.ErrInvalidHospitalReply,
	) {
		t.Fatalf(
			"expected ErrInvalidHospitalReply, got %v",
			err,
		)
	}
}
func TestHospitalClient_SearchPatient_Timeout(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				select {

				case <-time.After(
					200 * time.Millisecond,
				):

					w.WriteHeader(
						http.StatusOK,
					)

				case <-r.Context().Done():

					return
				}
			},
		),
	)

	defer server.Close()

	hospitalClient :=
		client.NewHospitalClient(
			50 * time.Millisecond,
		)

	_, err :=
		hospitalClient.SearchPatient(
			context.Background(),
			server.URL,
			"1234567890123",
		)

	if !errors.Is(
		err,
		client.ErrHospitalUnavailable,
	) {
		t.Fatalf(
			"expected ErrHospitalUnavailable for timeout, got %v",
			err,
		)
	}
}
func TestHospitalClient_SearchPatient_ContextCancelled(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				<-r.Context().Done()
			},
		),
	)

	defer server.Close()

	hospitalClient :=
		client.NewHospitalClient(
			time.Second,
		)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cancel()

	_, err :=
		hospitalClient.SearchPatient(
			ctx,
			server.URL,
			"1234567890123",
		)

	if !errors.Is(
		err,
		client.ErrHospitalUnavailable,
	) {
		t.Fatalf(
			"expected ErrHospitalUnavailable, got %v",
			err,
		)
	}
}
func TestHospitalClient_SearchPatient_PassportID(
	t *testing.T,
) {

	server := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {

				expected :=
					"/patient/search/AA123456"

				if r.URL.Path != expected {
					t.Fatalf(
						"expected %s, got %s",
						expected,
						r.URL.Path,
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				_, _ = w.Write(
					[]byte(`{
						"patient_hn": "HN0002",
						"passport_id": "AA123456",
						"first_name_en": "John",
						"last_name_en": "Doe"
					}`),
				)
			},
		),
	)

	defer server.Close()

	hospitalClient :=
		client.NewHospitalClient(
			time.Second,
		)

	result, err :=
		hospitalClient.SearchPatient(
			context.Background(),
			server.URL,
			"AA123456",
		)

	if err != nil {
		t.Fatalf(
			"expected no error, got %v",
			err,
		)
	}

	if result.PassportID != "AA123456" {
		t.Fatalf(
			"expected AA123456, got %s",
			result.PassportID,
		)
	}
}

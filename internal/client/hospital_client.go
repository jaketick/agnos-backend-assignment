package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"agnos-backend-assignment/internal/dto"
)

type HospitalClient interface {
	SearchPatient(
		ctx context.Context,
		baseURL string,
		id string,
	) (*dto.HospitalPatientResponse, error)
}

type hospitalClient struct {
	httpClient *http.Client
}

func NewHospitalClient(
	timeout time.Duration,
) HospitalClient {

	return &hospitalClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *hospitalClient) SearchPatient(
	ctx context.Context,
	baseURL string,
	id string,
) (*dto.HospitalPatientResponse, error) {

	baseURL = strings.TrimRight(
		baseURL,
		"/",
	)

	endpoint := fmt.Sprintf(
		"%s/patient/search/%s",
		baseURL,
		url.PathEscape(id),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create hospital request: %w",
			err,
		)
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, ErrHospitalUnavailable
	}

	defer resp.Body.Close()
	// Example usage:
	// hospitalClient := NewHospitalClient(5 * time.Second)
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing response.

	case http.StatusNotFound:
		return nil, ErrPatientNotFound

	default:
		if resp.StatusCode >= http.StatusInternalServerError {
			return nil, ErrHospitalUnavailable
		}

		return nil, fmt.Errorf(
			"unexpected hospital status: %d",
			resp.StatusCode,
		)
	}

	body, err := io.ReadAll(
		io.LimitReader(
			resp.Body,
			1<<20,
		),
	)

	if err != nil {
		return nil, ErrInvalidHospitalReply
	}

	var patient dto.HospitalPatientResponse

	if err := json.Unmarshal(
		body,
		&patient,
	); err != nil {
		return nil, ErrInvalidHospitalReply
	}

	return &patient, nil
}

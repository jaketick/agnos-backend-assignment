package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agnos-backend-assignment/internal/middleware"
	"agnos-backend-assignment/internal/service"
	"agnos-backend-assignment/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_MissingAuthorizationHeader(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	validateCalled := false
	nextCalled := false

	jwtService := &testutil.MockJWTService{
		ValidateTokenFunc: func(
			tokenString string,
		) (*service.JWTClaims, error) {

			validateCalled = true

			return nil, nil
		},
	}

	router := gin.New()

	router.GET(
		"/protected",
		middleware.Auth(jwtService),
		func(c *gin.Context) {
			nextCalled = true

			c.JSON(
				http.StatusOK,
				gin.H{
					"message": "success",
				},
			)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status 401, got %d",
			recorder.Code,
		)
	}

	if validateCalled {
		t.Fatal(
			"JWT validation should not be called when authorization header is missing",
		)
	}

	if nextCalled {
		t.Fatal(
			"protected handler must not be called",
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"authorization header is required",
	) {
		t.Fatalf(
			"unexpected response: %s",
			recorder.Body.String(),
		)
	}
}
func TestAuthMiddleware_InvalidAuthorizationHeader(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing bearer prefix",
			header: "abcdefg",
		},
		{
			name:   "wrong authorization type",
			header: "Basic abcdefg",
		},
		{
			name:   "too many parts",
			header: "Bearer abc def",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {

				validateCalled := false
				nextCalled := false

				jwtService :=
					&testutil.MockJWTService{
						ValidateTokenFunc: func(
							tokenString string,
						) (*service.JWTClaims, error) {

							validateCalled = true

							return nil, nil
						},
					}

				router := gin.New()

				router.GET(
					"/protected",
					middleware.Auth(jwtService),
					func(c *gin.Context) {
						nextCalled = true

						c.Status(
							http.StatusOK,
						)
					},
				)

				req := httptest.NewRequest(
					http.MethodGet,
					"/protected",
					nil,
				)

				req.Header.Set(
					"Authorization",
					tt.header,
				)

				recorder :=
					httptest.NewRecorder()

				router.ServeHTTP(
					recorder,
					req,
				)

				if recorder.Code !=
					http.StatusUnauthorized {

					t.Fatalf(
						"expected 401, got %d",
						recorder.Code,
					)
				}

				if validateCalled {
					t.Fatal(
						"JWT validation should not be called for invalid authorization format",
					)
				}

				if nextCalled {
					t.Fatal(
						"protected handler must not be called",
					)
				}
			},
		)
	}
}

func TestAuthMiddleware_InvalidToken(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	validateCalled := false
	nextCalled := false

	jwtService := &testutil.MockJWTService{
		ValidateTokenFunc: func(
			tokenString string,
		) (*service.JWTClaims, error) {

			validateCalled = true

			if tokenString != "bad-token" {
				t.Fatalf(
					"expected bad-token, got %s",
					tokenString,
				)
			}

			return nil,
				service.ErrInvalidToken
		},
	}

	router := gin.New()

	router.GET(
		"/protected",
		middleware.Auth(jwtService),
		func(c *gin.Context) {
			nextCalled = true

			c.Status(
				http.StatusOK,
			)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer bad-token",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code !=
		http.StatusUnauthorized {

		t.Fatalf(
			"expected 401, got %d",
			recorder.Code,
		)
	}

	if !validateCalled {
		t.Fatal(
			"expected JWT validation to be called",
		)
	}

	if nextCalled {
		t.Fatal(
			"protected handler must not be called for invalid token",
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"invalid or expired token",
	) {
		t.Fatalf(
			"unexpected response: %s",
			recorder.Body.String(),
		)
	}
}

func TestAuthMiddleware_ValidToken(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	jwtService := &testutil.MockJWTService{
		ValidateTokenFunc: func(
			tokenString string,
		) (*service.JWTClaims, error) {

			if tokenString != "valid-token" {
				t.Fatalf(
					"expected valid-token, got %s",
					tokenString,
				)
			}

			return &service.JWTClaims{
				StaffID:    10,
				HospitalID: 5,
				Username:   "staff01",
			}, nil
		},
	}

	nextCalled := false

	router := gin.New()

	router.GET(
		"/protected",
		middleware.Auth(jwtService),
		func(c *gin.Context) {
			nextCalled = true

			staffID, ok :=
				middleware.GetStaffID(c)

			if !ok {
				t.Fatal(
					"staff ID not found in context",
				)
			}

			hospitalID, ok :=
				middleware.GetHospitalID(c)

			if !ok {
				t.Fatal(
					"hospital ID not found in context",
				)
			}

			username, exists := c.Get(
				middleware.ContextUsername,
			)

			if !exists {
				t.Fatal(
					"username not found in context",
				)
			}

			if staffID != 10 {
				t.Fatalf(
					"expected staff ID 10, got %d",
					staffID,
				)
			}

			if hospitalID != 5 {
				t.Fatalf(
					"expected hospital ID 5, got %d",
					hospitalID,
				)
			}

			if username != "staff01" {
				t.Fatalf(
					"expected staff01, got %v",
					username,
				)
			}

			c.JSON(
				http.StatusOK,
				gin.H{
					"message": "success",
				},
			)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"Bearer valid-token",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d: %s",
			recorder.Code,
			recorder.Body.String(),
		)
	}

	if !nextCalled {
		t.Fatal(
			"expected protected handler to be called",
		)
	}
}

func TestAuthMiddleware_BearerCaseInsensitive(
	t *testing.T,
) {
	gin.SetMode(gin.TestMode)

	jwtService := &testutil.MockJWTService{
		ValidateTokenFunc: func(
			tokenString string,
		) (*service.JWTClaims, error) {

			return &service.JWTClaims{
				StaffID:    1,
				HospitalID: 1,
				Username:   "staff01",
			}, nil
		},
	}

	router := gin.New()

	router.GET(
		"/protected",
		middleware.Auth(jwtService),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/protected",
		nil,
	)

	req.Header.Set(
		"Authorization",
		"bearer valid-token",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			recorder.Code,
		)
	}
}

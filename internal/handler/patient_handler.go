package handler

import (
	"errors"
	"net/http"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/middleware"
	"agnos-backend-assignment/internal/service"

	"github.com/gin-gonic/gin"
)

type PatientHandler struct {
	patientService service.PatientService
}

func NewPatientHandler(
	patientService service.PatientService,
) *PatientHandler {

	return &PatientHandler{
		patientService: patientService,
	}
}

func (h *PatientHandler) Search(
	c *gin.Context,
) {

	hospitalID, ok :=
		middleware.GetHospitalID(c)

	if !ok {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "hospital context not found",
			},
		)

		return
	}

	var req dto.PatientSearchRequest

	if err := c.ShouldBindQuery(
		&req,
	); err != nil {

		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid search parameters",
			},
		)

		return
	}

	result, err := h.patientService.Search(
		c.Request.Context(),
		hospitalID,
		req,
	)

	if err != nil {

		switch {
		case errors.Is(
			err,
			service.ErrInvalidDateOfBirth,
		):
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "date_of_birth must use YYYY-MM-DD format",
				},
			)

		case errors.Is(
			err,
			service.ErrHospitalUnavailable,
		):
			c.JSON(
				http.StatusServiceUnavailable,
				gin.H{
					"error": "hospital service unavailable",
				},
			)

		default:
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"error": "internal server error",
				},
			)
		}

		return
	}

	c.JSON(
		http.StatusOK,
		result,
	)
}

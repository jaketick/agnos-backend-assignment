package handler

import (
	"errors"
	"net/http"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/service"

	"github.com/gin-gonic/gin"
)

type StaffHandler struct {
	staffService service.StaffService
}

func NewStaffHandler(
	staffService service.StaffService,
) *StaffHandler {
	return &StaffHandler{
		staffService: staffService,
	}
}

func (h *StaffHandler) Create(c *gin.Context) {
	var req dto.CreateStaffRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)

		return
	}

	result, err := h.staffService.Create(req)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStaffInput):
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"error": "username, password and hospital are required",
				},
			)

		case errors.Is(err, service.ErrHospitalNotFound):
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"error": "hospital not found",
				},
			)

		case errors.Is(err, service.ErrStaffAlreadyExists):
			c.JSON(
				http.StatusConflict,
				gin.H{
					"error": "staff already exists",
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
		http.StatusCreated,
		gin.H{
			"message": "staff created successfully",
			"data":    result,
		},
	)
}

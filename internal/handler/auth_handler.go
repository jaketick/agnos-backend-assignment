package handler

import (
	"errors"
	"net/http"

	"agnos-backend-assignment/internal/dto"
	"agnos-backend-assignment/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(
	authService service.AuthService,
) *AuthHandler {

	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"error": "invalid request",
			},
		)

		return
	}

	result, err := h.authService.Login(req)

	if err != nil {
		if errors.Is(
			err,
			service.ErrInvalidCredentials,
		) {
			c.JSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid username, password or hospital",
				},
			)

			return
		}

		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"error": "internal server error",
			},
		)

		return
	}

	c.JSON(
		http.StatusOK,
		result,
	)
}

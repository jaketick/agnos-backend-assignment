package handler

import (
	"net/http"

	"agnos-backend-assignment/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Profile(c *gin.Context) {
	staffID, exists := c.Get(
		middleware.ContextStaffID,
	)

	if !exists {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "staff context not found",
			},
		)

		return
	}

	hospitalID, exists := c.Get(
		middleware.ContextHospitalID,
	)

	if !exists {
		c.JSON(
			http.StatusUnauthorized,
			gin.H{
				"error": "hospital context not found",
			},
		)

		return
	}

	username, _ := c.Get(
		middleware.ContextUsername,
	)

	c.JSON(
		http.StatusOK,
		gin.H{
			"staff_id":    staffID,
			"hospital_id": hospitalID,
			"username":    username,
		},
	)
}

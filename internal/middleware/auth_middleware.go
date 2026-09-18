package middleware

import (
	"net/http"
	"strings"

	"agnos-backend-assignment/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	ContextStaffID    = "staff_id"
	ContextHospitalID = "hospital_id"
	ContextUsername   = "username"
)

func Auth(
	jwtService service.JWTService,
) gin.HandlerFunc {

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "authorization header is required",
				},
			)

			return
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {

			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid authorization header",
				},
			)

			return
		}

		claims, err := jwtService.ValidateToken(
			parts[1],
		)

		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"error": "invalid or expired token",
				},
			)

			return
		}

		c.Set(
			ContextStaffID,
			claims.StaffID,
		)

		c.Set(
			ContextHospitalID,
			claims.HospitalID,
		)

		c.Set(
			ContextUsername,
			claims.Username,
		)

		c.Next()
	}
}

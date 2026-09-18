package middleware

import "github.com/gin-gonic/gin"

func GetHospitalID(
	c *gin.Context,
) (uint, bool) {

	value, exists := c.Get(
		ContextHospitalID,
	)

	if !exists {
		return 0, false
	}

	hospitalID, ok := value.(uint)

	return hospitalID, ok
}

func GetStaffID(
	c *gin.Context,
) (uint, bool) {

	value, exists := c.Get(
		ContextStaffID,
	)

	if !exists {
		return 0, false
	}

	staffID, ok := value.(uint)

	return staffID, ok
}

func GetUsername(
	c *gin.Context,
) (string, bool) {

	value, exists := c.Get(
		ContextUsername,
	)

	if !exists {
		return "", false
	}

	username, ok := value.(string)

	return username, ok
}

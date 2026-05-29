package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseID extracts an int64 ID from the URL path
func ParseID(c *gin.Context, paramName string) (int64, error) {
	val := c.Param(paramName)
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil || id <= 0 {
		return 0, err
	}
	return id, nil
}

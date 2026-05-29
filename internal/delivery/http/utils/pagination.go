package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/mafi020/ecom-golang/internal/utils"
)

// ParseQueryParams is a helper for simple lists that don't need extra filters
func ParseQueryParams(c *gin.Context) utils.QueryParams {
	p := utils.QueryParams{
		Page:      1,
		Limit:     10,
		SortOrder: "desc",
		SortBy:    "created_at",
	}

	_ = c.ShouldBindQuery(&p)
	p.SetDefaults()
	return p
}

package request

import "github.com/mafi020/ecom-golang/internal/utils"

type GetUsersRequest struct {
	utils.QueryParams
	Role string `form:"role"`
}

func (r *GetUsersRequest) SetDefaults() {
	// This handles Page, Limit, SortBy ("created_at"), and SortOrder ("desc")
	r.QueryParams.SetDefaults()
}

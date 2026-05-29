package request

type CategoryRequest struct {
	Name     string `json:"name" binding:"required,min=3"`
	ParentID *int64 `json:"parent_id" binding:"numeric,omitempty,gt=0"`
}

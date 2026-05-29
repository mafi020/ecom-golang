package entity

import (
	"time"

	"github.com/mafi020/ecom-golang/internal/utils"
)

type Category struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	ParentID *int64 `json:"parent_id,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Products []Product `json:"products,omitempty"`
}

// entity/category.go

type GetCategoriesParams struct {
	utils.QueryParams
	ParentID int64 `form:"parent_id"`
}

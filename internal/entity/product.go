package entity

import (
	"time"

	"github.com/mafi020/ecom-golang/internal/utils"
)

type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CategoryID  int64     `json:"category_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	//  setup relationship with category
	Category *Category `json:"category,omitempty"`
}

func NewProduct(name string, description *string, price float64, stock int, categoryID int64) *Product {
	return &Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		CategoryID:  categoryID,
	}
}

type UpdateProductInput struct {
	Name        *string
	Description *string
	Price       *float64
	Stock       *int
	CategoryID  *int64
}

type GetProductsParams struct {
	utils.QueryParams
	CategoryID int64 `form:"category_id"`
}

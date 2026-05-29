package utils

type QueryParams struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search    string `form:"search"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
}

// Offset calculates the database offset
func (p QueryParams) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// SetDefaults ensures all fields have safe values for the DB
func (p *QueryParams) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit <= 0 {
		p.Limit = 10
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortOrder == "" {
		p.SortOrder = "desc"
	}
}

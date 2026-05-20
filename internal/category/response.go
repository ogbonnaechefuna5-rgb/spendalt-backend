package category

import "github.com/moninte/backend/internal/core"

// CategoryListResponse wraps a paginated list of categories.
type CategoryListResponse struct {
	Categories []*Category `json:"categories"`
	core.PageMeta
}

// BreakdownListResponse wraps a paginated category breakdown.
type BreakdownListResponse struct {
	Breakdown []*CategoryBreakdown `json:"breakdown"`
	core.PageMeta
}

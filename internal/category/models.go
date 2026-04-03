package category

import "github.com/spendalt/backend/internal/core"

type Category struct {
	core.BaseModel
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

type CategoryBreakdown struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}
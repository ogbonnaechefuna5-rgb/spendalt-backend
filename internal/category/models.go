package category

import "github.com/moninte/backend/internal/core"

type Category struct {
	core.BaseModel
	Name     string   `json:"name"`
	Icon     string   `json:"icon"`
	Color    string   `json:"color"`
	Keywords []string `json:"keywords"`
}

type CategoryBreakdown struct {
	Category string  `json:"category"`
	Count    int     `json:"count"`
	Total    float64 `json:"total"`
}
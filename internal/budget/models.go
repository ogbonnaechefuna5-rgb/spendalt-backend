package budget

import "github.com/moninte/backend/internal/core"

type Budget struct {
	core.UserScoped
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Period   string  `json:"period"`
}

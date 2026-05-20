package budget

// CreateBudgetRequest is the body for POST /budgets.
type CreateBudgetRequest struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Period   string  `json:"period"`
}

// UpdateBudgetRequest is the body for PUT /budgets/:id.
type UpdateBudgetRequest struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Period   string  `json:"period"`
}

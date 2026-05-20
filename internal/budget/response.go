package budget

import "github.com/moninte/backend/internal/core"

// BudgetResponse wraps a single budget for POST /budgets.
type BudgetResponse struct {
	Budget *Budget `json:"budget"`
}

// BudgetListResponse wraps a paginated list of budgets for GET /budgets.
type BudgetListResponse struct {
	Budgets []*Budget `json:"budgets"`
	core.PageMeta
}

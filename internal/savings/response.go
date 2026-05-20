package savings

import "github.com/moninte/backend/internal/core"

// GoalResponse wraps a single goal for POST /savings.
type GoalResponse struct {
	Goal *SavingsGoal `json:"goal"`
}

// GoalListResponse wraps a paginated list of goals.
type GoalListResponse struct {
	Goals []*SavingsGoal `json:"goals"`
	core.PageMeta
}

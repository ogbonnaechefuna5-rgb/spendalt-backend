package savings

import "time"

// CreateGoalRequest is the body for POST /savings.
type CreateGoalRequest struct {
	Name         string     `json:"name"`
	TargetAmount float64    `json:"target_amount"`
	Deadline     *time.Time `json:"deadline"`
}

// UpdateProgressRequest is the body for PUT /savings/:id/progress.
type UpdateProgressRequest struct {
	Amount float64 `json:"amount"`
}

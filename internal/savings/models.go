package savings

import "time"

type SavingsGoal struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Name          string     `json:"name"`
	TargetAmount  float64    `json:"target_amount"`
	CurrentAmount float64    `json:"current_amount"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
}


type SavingsResponse struct {
	TotalSaved  float64        `json:"totalSaved"`
	MonthlyGain float64        `json:"monthlyGain"`
	Goals       []*SavingsGoal `json:"goals"`
}

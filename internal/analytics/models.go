package analytics

type CategoryAmount struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
}

type Insights struct {
	TotalSpending float64          `json:"total_spending"`
	TotalIncome   float64          `json:"total_income"`
	Categories    []CategoryAmount `json:"categories"`
}

type WeeklyTrend struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type HealthInsight struct {
	Category    string `json:"category"`
	Status      string `json:"status"`
	Score       int    `json:"score"`
	Description string `json:"description"`
}

type HealthScore struct {
	Score           int             `json:"score"`
	Grade           string          `json:"grade"`
	Percentile      int             `json:"percentile"`
	Insights        []HealthInsight `json:"insights"`
	Recommendations []string        `json:"recommendations"`
}

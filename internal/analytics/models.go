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

// Composite response for the mobile analytics screen
type CategoryBreakdown struct {
	Name    string  `json:"name"`
	Amount  float64 `json:"amount"`
	Percent int     `json:"percent"`
}

type DayAmount struct {
	Day    string  `json:"day"`
	Amount float64 `json:"amount"`
}

type MerchantSummary struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Visits int    `json:"visits"`
}

type AnalyticsResponse struct {
	TotalSpend       float64             `json:"totalSpend"`
	TotalSpendChange int                 `json:"totalSpendChange"`
	Categories       []CategoryBreakdown `json:"categories"`
	Weekly           []DayAmount         `json:"weekly"`
	Merchants        []MerchantSummary   `json:"merchants"`
}

package analytics

// InsightsResponse wraps monthly insights for GET /analytics/insights.
type InsightsResponse struct {
	Insights *Insights `json:"insights"`
}

// WeeklyTrendResponse wraps the weekly trend for GET /analytics/weekly-trend.
type WeeklyTrendResponse struct {
	Trend []*WeeklyTrend `json:"trend"`
}

// HealthScoreResponse wraps the health score for GET /health/score.
type HealthScoreResponse struct {
	Health *HealthScore `json:"health"`
}

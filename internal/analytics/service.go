package analytics

type Service interface {
	GetInsights(userID string) (*Insights, error)
	GetWeeklyTrend(userID string) ([]*WeeklyTrend, error)
	GetHealthScore(userID string) (*HealthScore, error)
}

type service struct{ repo Repository }

func NewService(repo Repository) Service { return &service{repo: repo} }

func (s *service) GetInsights(userID string) (*Insights, error) {
	return s.repo.GetMonthlyInsights(userID)
}

func (s *service) GetWeeklyTrend(userID string) ([]*WeeklyTrend, error) {
	return s.repo.GetWeeklyTrend(userID)
}

func (s *service) GetHealthScore(userID string) (*HealthScore, error) {
	ins, err := s.repo.GetMonthlyInsights(userID)
	if err != nil {
		return nil, err
	}
	savingsRatio := 0.0
	if ins.TotalIncome > 0 {
		savingsRatio = ((ins.TotalIncome - ins.TotalSpending) / ins.TotalIncome) * 100
	}
	savingsScore := int(fmin(savingsRatio*5, 100))
	disciplineScore := 100
	if ins.TotalIncome > 0 {
		spendRatio := ins.TotalSpending / ins.TotalIncome
		if spendRatio > 0.8 {
			disciplineScore = int(fmax(0, 100-(spendRatio-0.8)*500))
		}
	}
	overall := (savingsScore + disciplineScore) / 2
	grade := "POOR"
	switch {
	case overall >= 80:
		grade = "EXCELLENT"
	case overall >= 65:
		grade = "GOOD"
	case overall >= 50:
		grade = "FAIR"
	}
	recs := []string{}
	if savingsScore < 50 {
		recs = append(recs, "Try to save at least 20% of your income each month.")
	}
	if disciplineScore < 50 {
		recs = append(recs, "Your spending is high relative to income. Review your budget.")
	}
	if len(recs) == 0 {
		recs = append(recs, "Great job! Keep maintaining your financial discipline.")
	}
	return &HealthScore{
		Score:      overall,
		Grade:      grade,
		Percentile: overall,
		Insights: []HealthInsight{
			{Category: "Savings Ratio", Status: gradeLabel(savingsScore), Score: savingsScore, Description: "Percentage of income saved each month"},
			{Category: "Spending Discipline", Status: gradeLabel(disciplineScore), Score: disciplineScore, Description: "How well you control discretionary spending"},
		},
		Recommendations: recs,
	}, nil
}

func gradeLabel(score int) string {
	switch {
	case score >= 80:
		return "Excellent"
	case score >= 65:
		return "Good"
	case score >= 50:
		return "Fair"
	default:
		return "Needs Work"
	}
}

func fmin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func fmax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

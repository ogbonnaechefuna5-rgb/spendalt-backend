package analytics

import "github.com/spendalt/backend/internal/common"

type Repository interface {
	GetMonthlyInsights(userID string) (*Insights, error)
	GetWeeklyTrend(userID string) ([]*WeeklyTrend, error)
}

type repository struct{ db common.DB }

func NewRepository(db common.DB) Repository { return &repository{db: db} }

func (r *repository) GetMonthlyInsights(userID string) (*Insights, error) {
	ins := &Insights{Categories: []CategoryAmount{}}
	err := r.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN transaction_type = 'debit' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN transaction_type = 'credit' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1
		  AND DATE_TRUNC('month', transaction_date) = DATE_TRUNC('month', NOW())`,
		userID,
	).Scan(&ins.TotalSpending, &ins.TotalIncome)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(`
		SELECT COALESCE(category, 'Other'), SUM(amount)
		FROM transactions
		WHERE user_id = $1 AND transaction_type = 'debit'
		  AND DATE_TRUNC('month', transaction_date) = DATE_TRUNC('month', NOW())
		GROUP BY category ORDER BY SUM(amount) DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ca CategoryAmount
		if err := rows.Scan(&ca.Category, &ca.Amount); err != nil {
			return nil, err
		}
		ins.Categories = append(ins.Categories, ca)
	}
	return ins, nil
}

func (r *repository) GetWeeklyTrend(userID string) ([]*WeeklyTrend, error) {
	rows, err := r.db.Query(`
		SELECT TO_CHAR(d::date, 'YYYY-MM-DD'), COALESCE(SUM(t.amount), 0)
		FROM generate_series(NOW() - INTERVAL '6 days', NOW(), INTERVAL '1 day') AS d
		LEFT JOIN transactions t
		       ON t.user_id = $1
		      AND t.transaction_type = 'debit'
		      AND DATE_TRUNC('day', t.transaction_date) = DATE_TRUNC('day', d)
		GROUP BY d ORDER BY d`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var trend []*WeeklyTrend
	for rows.Next() {
		wt := &WeeklyTrend{}
		if err := rows.Scan(&wt.Date, &wt.Amount); err != nil {
			return nil, err
		}
		trend = append(trend, wt)
	}
	return trend, nil
}

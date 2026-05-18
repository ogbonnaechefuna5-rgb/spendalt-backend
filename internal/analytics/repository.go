package analytics

import "github.com/spendalt/backend/internal/common"

type Repository interface {
	GetMonthlyInsights(userID string) (*Insights, error)
	GetWeeklyTrend(userID string) ([]*WeeklyTrend, error)
	GetComposite(userID, period string) (*AnalyticsResponse, error)
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


func (r *repository) GetComposite(userID, period string) (*AnalyticsResponse, error) {
	interval := "7 days"
	switch period {
	case "month":
		interval = "30 days"
	case "year":
		interval = "365 days"
	}

	resp := &AnalyticsResponse{}

	// Total spend this period
	r.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM transactions
		WHERE user_id=$1 AND transaction_type='debit'
		AND transaction_date >= NOW()-$2::interval`, userID, interval,
	).Scan(&resp.TotalSpend)

	// Previous period for change
	var prev float64
	r.db.QueryRow(`
		SELECT COALESCE(SUM(amount),0) FROM transactions
		WHERE user_id=$1 AND transaction_type='debit'
		AND transaction_date >= NOW()-($2::interval * 2)
		AND transaction_date < NOW()-$2::interval`, userID, interval,
	).Scan(&prev)
	if prev > 0 {
		resp.TotalSpendChange = int(((resp.TotalSpend - prev) / prev) * 100)
	}

	// Category breakdown
	catRows, err := r.db.Query(`
		SELECT COALESCE(category,'Other'), SUM(amount)
		FROM transactions WHERE user_id=$1 AND transaction_type='debit'
		AND transaction_date >= NOW()-$2::interval
		GROUP BY category ORDER BY SUM(amount) DESC`, userID, interval)
	if err == nil {
		defer catRows.Close()
		var cats []CategoryBreakdown
		var total float64
		for catRows.Next() {
			var cb CategoryBreakdown
			if catRows.Scan(&cb.Name, &cb.Amount) == nil {
				total += cb.Amount
				cats = append(cats, cb)
			}
		}
		for i := range cats {
			if total > 0 {
				cats[i].Percent = int((cats[i].Amount / total) * 100)
			}
		}
		resp.Categories = cats
	}

	// Weekly trend (last 7 days with day names)
	weekRows, err := r.db.Query(`
		SELECT TO_CHAR(d::date, 'Dy'), COALESCE(SUM(t.amount),0)
		FROM generate_series(NOW()-INTERVAL '6 days', NOW(), INTERVAL '1 day') AS d
		LEFT JOIN transactions t ON t.user_id=$1 AND t.transaction_type='debit'
			AND DATE_TRUNC('day', t.transaction_date)=DATE_TRUNC('day', d)
		GROUP BY d ORDER BY d`, userID)
	if err == nil {
		defer weekRows.Close()
		for weekRows.Next() {
			var da DayAmount
			if weekRows.Scan(&da.Day, &da.Amount) == nil {
				resp.Weekly = append(resp.Weekly, da)
			}
		}
	}

	// Top merchants
	merchRows, err := r.db.Query(`
		SELECT COALESCE(merchant_name,'Unknown'), COALESCE(SUM(amount),0)::int, COUNT(*)::int
		FROM transactions WHERE user_id=$1 AND transaction_type='debit'
		AND transaction_date >= NOW()-$2::interval
		AND merchant_name IS NOT NULL AND merchant_name != ''
		GROUP BY merchant_name ORDER BY SUM(amount) DESC LIMIT 5`, userID, interval)
	if err == nil {
		defer merchRows.Close()
		for merchRows.Next() {
			var m MerchantSummary
			if merchRows.Scan(&m.Name, &m.Amount, &m.Visits) == nil {
				resp.Merchants = append(resp.Merchants, m)
			}
		}
	}

	return resp, nil
}

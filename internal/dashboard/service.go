package dashboard

import (
	"github.com/moninte/backend/internal/common"
)

type Service interface {
	GetDashboard(userID string) (*DashboardResponse, error)
}

type service struct {
	db common.DB
}

func NewService(db common.DB) Service { return &service{db: db} }

func (s *service) GetDashboard(userID string) (*DashboardResponse, error) {
	resp := &DashboardResponse{}

	// Net worth & banks from linked_accounts
	rows, err := s.db.Query(
		`SELECT bank_name, COALESCE(balance, 0) FROM linked_accounts WHERE user_id = $1 AND status = 'active'`, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var bal float64
			if rows.Scan(&name, &bal) == nil {
				resp.Banks = append(resp.Banks, name)
				resp.NetWorth += bal
			}
		}
	}

	// This month spending & income
	s.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN transaction_type='debit' THEN amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN transaction_type='credit' THEN amount ELSE 0 END),0)
		FROM transactions WHERE user_id=$1
		AND DATE_TRUNC('month', transaction_date)=DATE_TRUNC('month', NOW())`, userID,
	).Scan(&resp.ThisMonth, &resp.Income)

	// Previous month for change calculation
	var prevSpend, prevIncome float64
	s.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN transaction_type='debit' THEN amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN transaction_type='credit' THEN amount ELSE 0 END),0)
		FROM transactions WHERE user_id=$1
		AND DATE_TRUNC('month', transaction_date)=DATE_TRUNC('month', NOW()-INTERVAL '1 month')`, userID,
	).Scan(&prevSpend, &prevIncome)

	resp.MonthChange = common.PctChange(resp.ThisMonth, prevSpend)
	resp.IncomeChange = common.PctChange(resp.Income, prevIncome)

	if resp.Income > 0 {
		resp.SavingsPct = int(((resp.Income - resp.ThisMonth) / resp.Income) * 100)
	}
	var prevSavingsPct int
	if prevIncome > 0 {
		prevSavingsPct = int(((prevIncome - prevSpend) / prevIncome) * 100)
	}
	resp.SavingsChange = resp.SavingsPct - prevSavingsPct
	// 7-day spending trend
	trendRows, err := s.db.Query(`
		SELECT COALESCE(SUM(t.amount),0)
		FROM generate_series(NOW()-INTERVAL '6 days', NOW(), INTERVAL '1 day') AS d
		LEFT JOIN transactions t ON t.user_id=$1 AND t.transaction_type='debit'
			AND DATE_TRUNC('day', t.transaction_date)=DATE_TRUNC('day', d)
		GROUP BY d ORDER BY d`, userID)
	if err == nil {
		defer trendRows.Close()
		for trendRows.Next() {
			var amt float64
			if trendRows.Scan(&amt) == nil {
				resp.SpendingTrend = append(resp.SpendingTrend, amt)
			}
		}
	}

	// Recent transactions (last 5)
	txRows, err := s.db.Query(`
		SELECT COALESCE(merchant_name,'Unknown'), COALESCE(category,'Other'),
			   amount, transaction_type, transaction_date
		FROM transactions WHERE user_id=$1
		ORDER BY transaction_date DESC LIMIT 5`, userID)
	if err == nil {
		defer txRows.Close()
		for txRows.Next() {
			var t TxItem
			var txType string
			if txRows.Scan(&t.Merchant, &t.Category, &t.Amount, &txType, &t.Time) == nil {
				if txType == "debit" {
					t.Amount = -t.Amount
				}
				resp.Transactions = append(resp.Transactions, t)
			}
		}
	}

	// Budget progress (top 3)
	budgetRows, err := s.db.Query(`
		SELECT b.category, b.amount,
			COALESCE((SELECT SUM(t.amount) FROM transactions t
				WHERE t.user_id=$1 AND t.transaction_type='debit'
				AND LOWER(t.category)=LOWER(b.category)
				AND DATE_TRUNC('month', t.transaction_date)=DATE_TRUNC('month', NOW())),0)
		FROM budgets b WHERE b.user_id=$1
		ORDER BY b.created_at LIMIT 3`, userID)
	if err == nil {
		defer budgetRows.Close()
		for budgetRows.Next() {
			var b BudgetItem
			if budgetRows.Scan(&b.Category, &b.Total, &b.Spent) == nil {
				resp.Budgets = append(resp.Budgets, b)
			}
		}
	}

	// AI insight placeholder
	resp.AiInsight = "Review your spending patterns to find savings opportunities."

	return resp, nil
}


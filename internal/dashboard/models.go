package dashboard

import "time"

type DashboardResponse struct {
	NetWorth      float64      `json:"netWorth"`
	Banks         []string     `json:"banks"`
	ThisMonth     float64      `json:"thisMonth"`
	Income        float64      `json:"income"`
	SavingsPct    int          `json:"savingsPct"`
	MonthChange   int          `json:"monthChange"`
	IncomeChange  int          `json:"incomeChange"`
	SavingsChange int          `json:"savingsChange"`
	SpendingTrend []float64    `json:"spendingTrend"`
	Transactions  []TxItem     `json:"transactions"`
	Budgets       []BudgetItem `json:"budgets"`
	AiInsight     string       `json:"aiInsight"`
}

type TxItem struct {
	Merchant string    `json:"merchant"`
	Category string    `json:"category"`
	Amount   float64   `json:"amount"`
	Time     time.Time `json:"time"`
}

type BudgetItem struct {
	Category string  `json:"category"`
	Spent    float64 `json:"spent"`
	Total    float64 `json:"total"`
}

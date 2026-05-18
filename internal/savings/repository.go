package savings

import (
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
)

type Repository interface {
	Create(g *SavingsGoal) error
	GetByUserID(userID string, limit, offset int) ([]*SavingsGoal, error)
	UpdateProgress(id, userID string, amount float64) error
	Delete(id, userID string) error
	GetSummary(userID string) (totalSaved float64, monthlyGain float64, err error)
}

type repository struct{ db common.DB }

func NewRepository(db common.DB) Repository { return &repository{db: db} }

func (r *repository) Create(g *SavingsGoal) error {
	return r.db.QueryRow(
		`INSERT INTO savings_goals (user_id, name, target_amount, deadline)
		 VALUES ($1, $2, $3, $4) RETURNING id, current_amount, status, created_at`,
		g.UserID, g.Name, g.TargetAmount, g.Deadline,
	).Scan(&g.ID, &g.CurrentAmount, &g.Status, &g.CreatedAt)
}

func (r *repository) GetByUserID(userID string, limit, offset int) ([]*SavingsGoal, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, name, target_amount, current_amount, deadline, status, created_at
		 FROM savings_goals WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return core.ScanRows(rows, func(g *SavingsGoal) []interface{} {
		return []interface{}{&g.ID, &g.UserID, &g.Name, &g.TargetAmount, &g.CurrentAmount, &g.Deadline, &g.Status, &g.CreatedAt}
	})
}

func (r *repository) UpdateProgress(id, userID string, amount float64) error {
	res, err := r.db.Exec(
		`UPDATE savings_goals
		 SET current_amount = LEAST(current_amount + $1, target_amount),
		     status = CASE WHEN current_amount + $1 >= target_amount THEN 'completed' ELSE status END
		 WHERE id = $2 AND user_id = $3`,
		amount, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(id, userID string) error {
	res, err := r.db.Exec(`DELETE FROM savings_goals WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}


func (r *repository) GetSummary(userID string) (float64, float64, error) {
	var totalSaved float64
	r.db.QueryRow(`SELECT COALESCE(SUM(current_amount),0) FROM savings_goals WHERE user_id=$1`, userID).Scan(&totalSaved)

	// Monthly gain: difference between current amounts now vs 30 days ago (approximated by goals created this month)
	var monthlyGain float64
	r.db.QueryRow(`
		SELECT COALESCE(SUM(current_amount),0) FROM savings_goals
		WHERE user_id=$1 AND created_at >= NOW()-INTERVAL '30 days'`, userID).Scan(&monthlyGain)

	return totalSaved, monthlyGain, nil
}

package budget

import (
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
)

type Repository interface {
	Create(b *Budget) error
	GetByUserID(userID string, limit, offset int) ([]*Budget, error)
	Update(b *Budget) error
	Delete(id, userID string) error
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(b *Budget) error {
	return r.db.QueryRow(
		`INSERT INTO budgets (user_id, category, amount, period) VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		b.UserID, b.Category, b.Amount, b.Period,
	).Scan(&b.ID, &b.CreatedAt)
}

func (r *repository) GetByUserID(userID string, limit, offset int) ([]*Budget, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, category, amount, period, created_at
		 FROM budgets WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(b *Budget) []interface{} {
		return []interface{}{&b.ID, &b.UserID, &b.Category, &b.Amount, &b.Period, &b.CreatedAt}
	})
}

func (r *repository) Update(b *Budget) error {
	_, err := r.db.Exec(
		`UPDATE budgets SET category = $1, amount = $2, period = $3 WHERE id = $4 AND user_id = $5`,
		b.Category, b.Amount, b.Period, b.ID, b.UserID,
	)
	return err
}

func (r *repository) Delete(id, userID string) error {
	_, err := r.db.Exec(`DELETE FROM budgets WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

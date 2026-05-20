package budget

import (
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
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
	res, err := r.db.Exec(
		`UPDATE budgets SET category = $1, amount = $2, period = $3 WHERE id = $4 AND user_id = $5`,
		b.Category, b.Amount, b.Period, b.ID, b.UserID,
	)
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
	res, err := r.db.Exec(`DELETE FROM budgets WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

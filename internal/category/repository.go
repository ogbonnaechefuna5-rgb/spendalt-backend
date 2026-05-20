package category

import (
	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
)

type Repository interface {
	GetAll(limit, offset int) ([]*Category, error)
	GetBreakdownByUserID(userID string, limit, offset int) ([]*CategoryBreakdown, error)
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll(limit, offset int) ([]*Category, error) {
	query := `SELECT id, name, keywords FROM categories ORDER BY name LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(cat *Category) []interface{} {
		return []interface{}{&cat.ID, &cat.Name, &cat.Keywords}
	})
}

func (r *repository) GetBreakdownByUserID(userID string, limit, offset int) ([]*CategoryBreakdown, error) {
	query := `SELECT COALESCE(category, 'Other'), COUNT(*), SUM(amount) as total
			  FROM transactions WHERE user_id = $1 AND transaction_type = 'debit'
			  GROUP BY category ORDER BY total DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(cb *CategoryBreakdown) []interface{} {
		return []interface{}{&cb.Category, &cb.Count, &cb.Total}
	})
}

package category

import (
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
)

type Repository interface {
	GetAll() ([]*Category, error)
	GetBreakdownByUserID(userID int) ([]*CategoryBreakdown, error)
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetAll() ([]*Category, error) {
	query := `SELECT id, name, keywords FROM categories ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return core.ScanRows(rows, func(cat *Category) []interface{} {
		return []interface{}{&cat.ID, &cat.Name, &cat.Keywords}
	})
}

func (r *repository) GetBreakdownByUserID(userID int) ([]*CategoryBreakdown, error) {
	query := `SELECT category, SUM(amount) as total 
			  FROM transactions WHERE user_id = $1 AND type = 'debit' 
			  GROUP BY category ORDER BY total DESC`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return core.ScanRows(rows, func(cb *CategoryBreakdown) []interface{} {
		return []interface{}{&cb.Category, &cb.Total}
	})
}
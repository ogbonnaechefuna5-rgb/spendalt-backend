package core

import "github.com/spendalt/backend/internal/common"

// Repository provides base CRUD operations for a domain model T.
// Domain repositories embed this and supply table name + scan func.
type Repository[T any] struct {
	DB    common.DB
	Table string
	Scan  func(*T) []interface{}
}

func (r *Repository[T]) GetByID(id int) (*T, error) {
	rows, err := r.DB.Query("SELECT * FROM "+r.Table+" WHERE id = $1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}
	results, err := ScanRows(rows, r.Scan)
	if err != nil || len(results) == 0 {
		return nil, err
	}
	return results[0], nil
}

func (r *Repository[T]) GetAllByUserID(userID int) ([]*T, error) {
	rows, err := r.DB.Query("SELECT * FROM "+r.Table+" WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, err
	}
	return ScanRows(rows, r.Scan)
}

func (r *Repository[T]) Delete(id int) error {
	_, err := r.DB.Exec("DELETE FROM "+r.Table+" WHERE id = $1", id)
	return err
}

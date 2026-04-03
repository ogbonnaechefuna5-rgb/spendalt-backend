package transaction

import (
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
)

type Repository interface {
	Create(tx *Transaction) error
	GetByUserID(userID int, limit, offset int) ([]*Transaction, error)
	GetByFingerprint(fingerprint string) (*Transaction, error)
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(tx *Transaction) error {
	query := `INSERT INTO transactions (user_id, amount, type, merchant, category, description, fingerprint, transaction_date) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
	return r.db.QueryRow(query, tx.UserID, tx.Amount, tx.Type, tx.Merchant, tx.Category, tx.Description, tx.Fingerprint, tx.TransactionDate).
		Scan(&tx.ID, &tx.CreatedAt)
}

func (r *repository) GetByUserID(userID int, limit, offset int) ([]*Transaction, error) {
	query := `SELECT id, user_id, amount, type, merchant, category, description, fingerprint, transaction_date, created_at 
			  FROM transactions WHERE user_id = $1 ORDER BY transaction_date DESC LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return core.ScanRows(rows, func(tx *Transaction) []interface{} {
		return []interface{}{&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Merchant, &tx.Category, &tx.Description, &tx.Fingerprint, &tx.TransactionDate, &tx.CreatedAt}
	})
}

func (r *repository) GetByFingerprint(fingerprint string) (*Transaction, error) {
	tx := &Transaction{}
	query := `SELECT id, user_id, amount, type, merchant, category, description, fingerprint, transaction_date, created_at 
			  FROM transactions WHERE fingerprint = $1`
	err := r.db.QueryRow(query, fingerprint).Scan(
		&tx.ID, &tx.UserID, &tx.Amount, &tx.Type, &tx.Merchant, &tx.Category, &tx.Description, &tx.Fingerprint, &tx.TransactionDate, &tx.CreatedAt,
	)
	return tx, err
}
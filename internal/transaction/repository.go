package transaction

import (
	"github.com/spendalt/backend/internal/common"
	"github.com/spendalt/backend/internal/core"
)

type Repository interface {
	// raw_transactions
	CreateRaw(r *RawTransaction) error
	GetRawByID(id string) (*RawTransaction, error)
	GetUnprocessed(limit int) ([]*RawTransaction, error)
	MarkProcessed(id string) error

	// transactions
	Create(tx *Transaction) error
	GetByUserID(userID string, limit, offset int) ([]*Transaction, error)
	GetByFingerprint(fingerprint string) (*Transaction, error)
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateRaw(raw *RawTransaction) error {
	return r.db.QueryRow(
		`INSERT INTO raw_transactions (user_id, source, raw_text)
		 VALUES ($1, $2, $3) RETURNING id, created_at`,
		raw.UserID, raw.Source, raw.RawText,
	).Scan(&raw.ID, &raw.CreatedAt)
}

func (r *repository) GetRawByID(id string) (*RawTransaction, error) {
	raw := &RawTransaction{}
	err := r.db.QueryRow(
		`SELECT id, user_id, source, raw_text, processed, created_at
		 FROM raw_transactions WHERE id = $1`,
		id,
	).Scan(&raw.ID, &raw.UserID, &raw.Source, &raw.RawText, &raw.Processed, &raw.CreatedAt)
	return raw, err
}
func (r *repository) GetUnprocessed(limit int) ([]*RawTransaction, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, source, raw_text, processed, created_at
		 FROM raw_transactions WHERE processed = FALSE
		 ORDER BY created_at ASC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(r *RawTransaction) []interface{} {
		return []interface{}{&r.ID, &r.UserID, &r.Source, &r.RawText, &r.Processed, &r.CreatedAt}
	})
}

func (r *repository) MarkProcessed(id string) error {
	_, err := r.db.Exec(`UPDATE raw_transactions SET processed = TRUE WHERE id = $1`, id)
	return err
}

func (r *repository) Create(tx *Transaction) error {
	return r.db.QueryRow(
		`INSERT INTO transactions (user_id, amount, transaction_type, merchant_name, category, description, fingerprint, transaction_date, source)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id, created_at`,
		tx.UserID, tx.Amount, tx.Type, tx.Merchant, tx.Category, tx.Description, tx.Fingerprint, tx.TransactionDate, tx.Source,
	).Scan(&tx.ID, &tx.CreatedAt)
}

func (r *repository) GetByUserID(userID string, limit, offset int) ([]*Transaction, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, amount, transaction_type,
		        COALESCE(merchant_name, ''), COALESCE(category, ''),
		        COALESCE(description, ''), COALESCE(fingerprint, ''),
		        COALESCE(source, 'manual'), transaction_date, created_at
		 FROM transactions WHERE user_id = $1 ORDER BY transaction_date DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(tx *Transaction) []interface{} {
		return []interface{}{
			&tx.ID, &tx.UserID, &tx.Amount, &tx.Type,
			&tx.Merchant, &tx.Category, &tx.Description, &tx.Fingerprint,
			&tx.Source, &tx.TransactionDate, &tx.CreatedAt,
		}
	})
}

func (r *repository) GetByFingerprint(fingerprint string) (*Transaction, error) {
	tx := &Transaction{}
	err := r.db.QueryRow(
		`SELECT id, user_id, amount, transaction_type,
		        COALESCE(merchant_name, ''), COALESCE(category, ''),
		        COALESCE(description, ''), COALESCE(fingerprint, ''),
		        COALESCE(source, 'manual'), transaction_date, created_at
		 FROM transactions WHERE fingerprint = $1`,
		fingerprint,
	).Scan(
		&tx.ID, &tx.UserID, &tx.Amount, &tx.Type,
		&tx.Merchant, &tx.Category, &tx.Description, &tx.Fingerprint,
		&tx.Source, &tx.TransactionDate, &tx.CreatedAt,
	)
	return tx, err
}

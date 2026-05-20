package transaction

import "github.com/moninte/backend/internal/core"

// IngestSMSResponse is the body returned by POST /transactions/ingest/sms.
type IngestSMSResponse struct {
	Message string `json:"message"`
	ID      string `json:"id"`
}

// TransactionResponse wraps a single transaction for POST /transactions/ingest/manual.
type TransactionResponse struct {
	Transaction *Transaction `json:"transaction"`
}

// TransactionListResponse wraps a paginated list for GET /transactions.
type TransactionListResponse struct {
	Transactions []*Transaction `json:"transactions"`
	core.PageMeta
}

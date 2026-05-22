package transaction

import (
	"context"
	"log"
	"time"

	"github.com/moninte/backend/internal/core"
)

const (
	workerBatchSize    = 50
	workerPollInterval = 5 * time.Second
)

// Worker processes raw_transactions asynchronously.
// SMS ingestion writes to raw_transactions and enqueues the ID here.
// The worker loop also polls the DB on an interval to catch any records
// that were missed (e.g. after a restart).
type Worker struct {
	repo        Repository
	categorizer Categorizer
	queue       chan string // raw_transaction IDs
}

func NewWorker(repo Repository, categorizer Categorizer) *Worker {
	return &Worker{
		repo:        repo,
		categorizer: categorizer,
		queue:       make(chan string, 512),
	}
}

// Enqueue adds a raw transaction ID to the processing queue.
// Non-blocking: if the buffer is full the record will be picked up by the poll loop.
func (w *Worker) Enqueue(id string) {
	select {
	case w.queue <- id:
	default:
		log.Printf("[worker] queue full, raw tx %s will be picked up by poll", id)
	}
}

// Run starts the worker. Call in a goroutine; cancel ctx to stop.
func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(workerPollInterval)
	defer ticker.Stop()

	for {
		select {
		case id := <-w.queue:
			w.processID(ctx, id)
		case <-ticker.C:
			w.poll(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// poll fetches any unprocessed rows missed by the channel (e.g. after restart).
func (w *Worker) poll(ctx context.Context) {
	raws, err := w.repo.GetUnprocessed(workerBatchSize)
	if err != nil {
		log.Printf("[worker] poll error: %v", err)
		return
	}
	for _, raw := range raws {
		if ctx.Err() != nil {
			return
		}
		w.process(ctx, raw)
	}
}

func (w *Worker) processID(ctx context.Context, id string) {
	raw, err := w.repo.GetRawByID(id)
	if err != nil || raw.Processed {
		return
	}
	w.process(ctx, raw)
}

func (w *Worker) process(ctx context.Context, raw *RawTransaction) {
	if ctx.Err() != nil {
		return
	}

	amount, txType, merchant, reference, description, txTime, err := parseSMSWithMeta(raw.RawText)
	if err != nil {
		log.Printf("[worker] parse failed for raw %s: %v — marking processed", raw.ID, err)
		_ = w.repo.MarkProcessed(raw.ID)
		return
	}

	if txTime.IsZero() {
		txTime = raw.CreatedAt
	}

	balance := parseBalance(raw.RawText)
	fingerprint := generateFingerprint(raw.UserID, amount, merchant, reference, txTime)
	if _, err := w.repo.GetByFingerprint(fingerprint); err == nil {
		_ = w.repo.MarkProcessed(raw.ID)
		return
	}

	rawID := raw.ID
	tx := &Transaction{
		UserScoped:       core.UserScoped{UserID: raw.UserID},
		RawTransactionID: &rawID,
		Amount:           amount,
		Type:             txType,
		Merchant:         merchant,
		Description:      description,
		Category:         w.categorizer.Categorize(merchant),
		Fingerprint:      fingerprint,
		Source:           "sms",
		TransactionDate:  txTime,
		BalanceAfter:     balance,
	}

	if err := w.repo.Create(tx); err != nil {
		log.Printf("[worker] create failed for raw %s: %v", raw.ID, err)
		return
	}

	if err := w.repo.MarkProcessed(raw.ID); err != nil {
		log.Printf("[worker] mark processed failed for raw %s: %v", raw.ID, err)
	}
}

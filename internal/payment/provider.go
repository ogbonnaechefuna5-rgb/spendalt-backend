package payment

import "time"

const EventChargeSuccess = "charge.success"

// WebhookEvent is the normalised event all providers produce.
type WebhookEvent struct {
	Type      string
	UserID    string
	PlanID    string
	Reference string
	Amount    float64
	PeriodEnd *time.Time
}

// TransactionResult is returned by VerifyTransaction.
type TransactionResult struct {
	Reference string
	Amount    float64
	Status    string // "success", "failed", "pending"
	Metadata  map[string]string
}

// Provider is the interface every payment processor must implement.
type Provider interface {
	Name() string
	HandleWebhook(payload []byte, signature string) (*WebhookEvent, error)
	VerifyTransaction(reference string) (*TransactionResult, error)
}

// Registry holds all registered providers keyed by name.
type Registry struct {
	providers map[string]Provider
}

func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]Provider)}
}

func (r *Registry) Register(p Provider) {
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

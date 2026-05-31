package subscription

import (
	"database/sql"
	"errors"
	"time"

	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
)

type Repository interface {
	GetActiveByUserID(userID string) (*Subscription, error)
	HasEntitlement(userID, feature string) (bool, error)
	Activate(userID, planID, provider, providerRef string, periodEnd *time.Time) error
	Cancel(userID string) error
	GetPlans() ([]*Plan, error)
}

type repository struct{ db common.DB }

func NewRepository(db common.DB) Repository { return &repository{db: db} }

func (r *repository) GetActiveByUserID(userID string) (*Subscription, error) {
	s := &Subscription{}
	p := &Plan{}
	err := r.db.QueryRow(`
		SELECT s.id, s.user_id, s.plan_id, s.status, COALESCE(s.provider,''),
		       COALESCE(s.provider_reference,''), s.current_period_start,
		       s.current_period_end, s.cancelled_at, s.created_at,
		       p.name, p.description, p.price_ngn, p.interval
		FROM subscriptions s
		JOIN plans p ON p.id = s.plan_id
		WHERE s.user_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC LIMIT 1`, userID,
	).Scan(
		&s.ID, &s.UserID, &s.PlanID, &s.Status, &s.Provider,
		&s.ProviderReference, &s.CurrentPeriodStart,
		&s.CurrentPeriodEnd, &s.CancelledAt, &s.CreatedAt,
		&p.Name, &p.Description, &p.PriceNGN, &p.Interval,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.ID = s.PlanID
	s.Plan = p
	return s, nil
}

func (r *repository) HasEntitlement(userID, feature string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM subscriptions s
		JOIN plan_entitlements pe ON pe.plan_id = s.plan_id
		WHERE s.user_id = $1
		  AND s.status = 'active'
		  AND (s.current_period_end IS NULL OR s.current_period_end > NOW())
		  AND pe.feature = $2`, userID, feature,
	).Scan(&count)
	return count > 0, err
}

func (r *repository) Activate(userID, planID, provider, providerRef string, periodEnd *time.Time) error {
	// Expire any existing active subscription first
	_, err := r.db.Exec(`
		UPDATE subscriptions SET status = 'expired'
		WHERE user_id = $1 AND status = 'active'`, userID)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
		INSERT INTO subscriptions (user_id, plan_id, status, provider, provider_reference, current_period_end)
		VALUES ($1, $2, 'active', $3, $4, $5)`,
		userID, planID, provider, providerRef, periodEnd)
	return err
}

func (r *repository) Cancel(userID string) error {
	now := time.Now()
	res, err := r.db.Exec(`
		UPDATE subscriptions SET status = 'cancelled', cancelled_at = $1
		WHERE user_id = $2 AND status = 'active'`, now, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *repository) GetPlans() ([]*Plan, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.name, p.description, p.price_ngn, p.interval,
		       COALESCE(array_agg(pe.feature) FILTER (WHERE pe.feature IS NOT NULL), ARRAY[]::text[])
		FROM plans p
		LEFT JOIN plan_entitlements pe ON pe.plan_id = p.id
		WHERE p.is_active = TRUE
		GROUP BY p.id ORDER BY p.price_ngn ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return core.ScanRows(rows, func(p *Plan) []interface{} {
		return []interface{}{&p.ID, &p.Name, &p.Description, &p.PriceNGN, &p.Interval, &p.Features}
	})
}

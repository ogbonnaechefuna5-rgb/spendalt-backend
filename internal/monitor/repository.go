package monitor

import (
	"context"
	"log"

	"github.com/moninte/backend/internal/common"
)

// Repository persists and queries request logs (SRP — only storage concern).
type Repository interface {
	Save(log *RequestLog)
	GetByUserID(userID string, limit int) ([]*RequestLog, error)
}

type repository struct {
	db   common.DB
	queue chan *RequestLog
}

// NewRepository returns a Repository that writes asynchronously via a
// buffered channel so the HTTP response is never blocked by a DB write.
func NewRepository(db common.DB, ctx context.Context) Repository {
	r := &repository{db: db, queue: make(chan *RequestLog, 512)}
	go r.drain(ctx)
	return r
}

// Save enqueues a log entry — non-blocking, drops if queue is full.
func (r *repository) Save(entry *RequestLog) {
	select {
	case r.queue <- entry:
	default:
		log.Println("[monitor] queue full, dropping request log")
	}
}

func (r *repository) drain(ctx context.Context) {
	for {
		select {
		case entry := <-r.queue:
			r.insert(entry)
		case <-ctx.Done():
			// Flush remaining entries before exit
			for {
				select {
				case entry := <-r.queue:
					r.insert(entry)
				default:
					return
				}
			}
		}
	}
}

func (r *repository) GetByUserID(userID string, limit int) ([]*RequestLog, error) {
	rows, err := r.db.Query(`
		SELECT id, request_id, user_id, method, path, status, latency_ms, error,
		       device_id, device_type, os, app_version, user_agent,
		       ip, forwarded_for, real_ip, host, protocol, tls,
		       origin, referer, accept_language, created_at
		FROM request_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []*RequestLog
	for rows.Next() {
		e := &RequestLog{}
		if err := rows.Scan(
			&e.ID, &e.RequestID, &e.UserID, &e.Method, &e.Path, &e.Status, &e.LatencyMs, &e.Error,
			&e.DeviceID, &e.DeviceType, &e.OS, &e.AppVersion, &e.UserAgent,
			&e.IP, &e.ForwardedFor, &e.RealIP, &e.Host, &e.Protocol, &e.TLS,
			&e.Origin, &e.Referer, &e.AcceptLanguage, &e.CreatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, e)
	}
	return logs, nil
}

func (r *repository) insert(e *RequestLog) {
	_, err := r.db.Exec(`
		INSERT INTO request_logs (
			request_id, user_id, method, path, status, latency_ms, error,
			device_id, device_type, os, app_version, user_agent,
			ip, forwarded_for, real_ip, host, protocol, tls,
			origin, referer, accept_language
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,
			$8,$9,$10,$11,$12,
			$13,$14,$15,$16,$17,$18,
			$19,$20,$21
		)`,
		e.RequestID, e.UserID, e.Method, e.Path, e.Status, e.LatencyMs, e.Error,
		e.DeviceID, e.DeviceType, e.OS, e.AppVersion, e.UserAgent,
		e.IP, e.ForwardedFor, e.RealIP, e.Host, e.Protocol, e.TLS,
		e.Origin, e.Referer, e.AcceptLanguage,
	)
	if err != nil {
		log.Println("[monitor] failed to persist request log:", err)
	}
}

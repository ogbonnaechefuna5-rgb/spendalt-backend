package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const tokenTTL = 2 * time.Hour

type TokenStore interface {
	Revoke(tokenID string) error
	IsRevoked(tokenID string) (bool, error)
}

// ── Redis store ──────────────────────────────────────────────────────────────

type redisTokenStore struct {
	client *redis.Client
}

func NewTokenStore(redisURL string) (TokenStore, error) {
	if redisURL == "" {
		return NewMemoryTokenStore(), nil
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return NewMemoryTokenStore(), nil
	}
	client := redis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("[warn] Redis unavailable, using in-memory token store: %v\n", err)
		return NewMemoryTokenStore(), nil
	}
	return &redisTokenStore{client: client}, nil
}

func (s *redisTokenStore) Revoke(tokenID string) error {
	return s.client.Set(context.Background(), revokedKey(tokenID), 1, tokenTTL).Err()
}

func (s *redisTokenStore) IsRevoked(tokenID string) (bool, error) {
	val, err := s.client.Exists(context.Background(), revokedKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

func revokedKey(tokenID string) string {
	return "revoked:" + tokenID
}

// ── In-memory store (fallback when Redis is unavailable) ─────────────────────

type memoryTokenStore struct {
	mu      sync.RWMutex
	revoked map[string]time.Time
}

func NewMemoryTokenStore() TokenStore {
	s := &memoryTokenStore{revoked: make(map[string]time.Time)}
	go s.cleanup()
	return s
}

func (s *memoryTokenStore) Revoke(tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[tokenID] = time.Now().Add(tokenTTL)
	return nil
}

func (s *memoryTokenStore) IsRevoked(tokenID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exp, ok := s.revoked[tokenID]
	if !ok {
		return false, nil
	}
	return time.Now().Before(exp), nil
}

func (s *memoryTokenStore) cleanup() {
	for range time.Tick(10 * time.Minute) {
		s.mu.Lock()
		for id, exp := range s.revoked {
			if time.Now().After(exp) {
				delete(s.revoked, id)
			}
		}
		s.mu.Unlock()
	}
}

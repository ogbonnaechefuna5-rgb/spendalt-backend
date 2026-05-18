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
	GetCounter(key string) (int, error)
	IncrCounter(key string, ttl time.Duration) error
	DeleteCounter(key string) error
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

func (s *redisTokenStore) GetCounter(key string) (int, error) {
	val, err := s.client.Get(context.Background(), "counter:"+key).Int()
	if err != nil {
		return 0, nil // key not found = 0 attempts
	}
	return val, nil
}

func (s *redisTokenStore) IncrCounter(key string, ttl time.Duration) error {
	ctx := context.Background()
	k := "counter:" + key
	pipe := s.client.Pipeline()
	pipe.Incr(ctx, k)
	pipe.Expire(ctx, k, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *redisTokenStore) DeleteCounter(key string) error {
	return s.client.Del(context.Background(), "counter:"+key).Err()
}

func revokedKey(tokenID string) string {
	return "revoked:" + tokenID
}

// ── In-memory store (fallback when Redis is unavailable) ─────────────────────

type memoryTokenStore struct {
	mu      sync.RWMutex
	revoked  map[string]time.Time
	counters map[string]memCounter
}

type memCounter struct {
	count int
	exp   time.Time
}

func NewMemoryTokenStore() TokenStore {
	s := &memoryTokenStore{
		revoked:  make(map[string]time.Time),
		counters: make(map[string]memCounter),
	}
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

func (s *memoryTokenStore) GetCounter(key string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.counters[key]
	if !ok || time.Now().After(c.exp) {
		return 0, nil
	}
	return c.count, nil
}

func (s *memoryTokenStore) IncrCounter(key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.counters[key]
	if time.Now().After(c.exp) {
		c = memCounter{exp: time.Now().Add(ttl)}
	}
	c.count++
	s.counters[key] = c
	return nil
}

func (s *memoryTokenStore) DeleteCounter(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.counters, key)
	return nil
}

func (s *memoryTokenStore) cleanup() {
	for range time.Tick(10 * time.Minute) {
		s.mu.Lock()
		for id, exp := range s.revoked {
			if time.Now().After(exp) {
				delete(s.revoked, id)
			}
		}
		for k, c := range s.counters {
			if time.Now().After(c.exp) {
				delete(s.counters, k)
			}
		}
		s.mu.Unlock()
	}
}

package auth_test

import (
	"testing"
	"time"

	"github.com/moninte/backend/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenStore_RevokeAndIsRevoked(t *testing.T) {
	s := auth.NewMemoryTokenStore()

	revoked, err := s.IsRevoked("tok1")
	require.NoError(t, err)
	assert.False(t, revoked)

	require.NoError(t, s.Revoke("tok1"))

	revoked, err = s.IsRevoked("tok1")
	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestTokenStore_Counter(t *testing.T) {
	s := auth.NewMemoryTokenStore()

	count, err := s.GetCounter("key1")
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	require.NoError(t, s.IncrCounter("key1", time.Minute))
	require.NoError(t, s.IncrCounter("key1", time.Minute))

	count, err = s.GetCounter("key1")
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	require.NoError(t, s.DeleteCounter("key1"))

	count, err = s.GetCounter("key1")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestTokenStore_CounterExpiry(t *testing.T) {
	s := auth.NewMemoryTokenStore()

	require.NoError(t, s.IncrCounter("expkey", 1*time.Millisecond))
	time.Sleep(5 * time.Millisecond)

	count, err := s.GetCounter("expkey")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

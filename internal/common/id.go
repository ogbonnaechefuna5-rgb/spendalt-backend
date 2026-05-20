package common

import "github.com/google/uuid"

// NewID returns a new UUID v7 string.
// v7 is time-ordered, making it safe to use as a primary key and for log correlation.
func NewID() string {
	id, err := uuid.NewV7()
	if err != nil {
		// Fallback to v4 if the system clock is unavailable (extremely rare)
		return uuid.NewString()
	}
	return id.String()
}

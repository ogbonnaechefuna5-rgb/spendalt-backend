package monitor

import (
	"log/slog"
	"os"
)

// Logger is the abstraction both middleware and httpclient depend on (DIP).
type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

type slogLogger struct{ l *slog.Logger }

func NewLogger() Logger {
	return &slogLogger{slog.New(slog.NewJSONHandler(os.Stdout, nil))}
}

func (s *slogLogger) Info(msg string, args ...any)  { s.l.Info(msg, args...) }
func (s *slogLogger) Error(msg string, args ...any) { s.l.Error(msg, args...) }

package common

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	Close() error
	Ping() error
}

type PostgresDB struct {
	*sql.DB
}

func NewPostgresDB(databaseURL string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	for i := range 10 {
		if err = db.Ping(); err == nil {
			return &PostgresDB{DB: db}, nil
		}
		fmt.Printf("[db] waiting for postgres (attempt %d/10): %v\n", i+1, err)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to ping database: %w", err)
}
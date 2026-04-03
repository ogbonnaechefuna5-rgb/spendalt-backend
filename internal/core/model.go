package core

import (
	"database/sql"
	"time"
)

type BaseModel struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserScoped struct {
	BaseModel
	UserID int `json:"user_id"`
}

func ScanRows[T any](rows *sql.Rows, scan func(*T) []interface{}) ([]*T, error) {
	defer rows.Close()
	var results []*T
	for rows.Next() {
		item := new(T)
		if err := rows.Scan(scan(item)...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

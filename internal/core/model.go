package core

import "time"

type BaseModel struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserScoped struct {
	BaseModel
	UserID string `json:"user_id"`
}

package user

import "github.com/spendalt/backend/internal/core"

type User struct {
	core.BaseModel
	Phone        string `json:"phone"`
	Email        string `json:"email,omitempty"`
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	PasswordHash string `json:"-"`
}
package user

import "github.com/spendalt/backend/internal/common"

type Repository interface {
	Create(user *User) error
	GetByEmail(email string) (*User, error)
	GetByID(id int) (*User, error)
	Update(user *User) error
	Delete(id int) error
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *User) error {
	query := `INSERT INTO users (email, password_hash, first_name, last_name, phone) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRow(query, user.Email, user.PasswordHash, user.FirstName, user.LastName, user.Phone).
		Scan(&user.ID, &user.CreatedAt)
}

func (r *repository) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash, first_name, last_name, phone, created_at 
			  FROM users WHERE email = $1`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, 
		&user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt,
	)
	return user, err
}

func (r *repository) GetByID(id int) (*User, error) {
	user := &User{}
	query := `SELECT id, email, password_hash, first_name, last_name, phone, created_at 
			  FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, 
		&user.FirstName, &user.LastName, &user.Phone, &user.CreatedAt,
	)
	return user, err
}

func (r *repository) Update(user *User) error {
	query := `UPDATE users SET first_name = $1, last_name = $2, phone = $3 WHERE id = $4`
	_, err := r.db.Exec(query, user.FirstName, user.LastName, user.Phone, user.ID)
	return err
}

func (r *repository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
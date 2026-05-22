package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/moninte/backend/internal/common"
	"github.com/moninte/backend/internal/core"
)

type Repository interface {
	Create(user *User) error
	GetByEmail(email string) (*User, error)
	GetByPhone(phone string) (*User, error)
	GetByID(id string) (*User, error)
	Update(user *User) error
	UpdateAvatar(userID, avatarURL string) error
	Delete(id string) error
	GetPreferences(userID string) (*UserPreferences, error)
	SavePreferences(p *UserPreferences) error
	GetLinkedAccounts(userID string, limit, offset int) ([]*LinkedAccount, error)
	RemoveLinkedAccount(userID, accountID string) error
	SyncLinkedAccount(userID, accountID string) error
	GetSessions(userID string, limit, offset int) ([]*UserSession, error)
	RevokeAllSessions(userID string) error
	RevokeSession(userID, sessionID string) error
	CreateSession(session *UserSession) error
}

type repository struct {
	db common.DB
}

func NewRepository(db common.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(user *User) error {
	var email *string
	if user.Email != "" {
		email = &user.Email
	}
	query := `INSERT INTO users (email, password_hash, first_name, middle_name, last_name, phone)
			  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	return r.db.QueryRow(query, email, user.PasswordHash, user.FirstName, user.MiddleName, user.LastName, user.Phone).
		Scan(&user.ID, &user.CreatedAt)
}

func (r *repository) GetByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, COALESCE(email,''), password_hash, first_name, COALESCE(middle_name,''), last_name, phone, COALESCE(avatar_url,''), created_at
			  FROM users WHERE LOWER(email) = LOWER($1)`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.MiddleName, &user.LastName, &user.Phone, &user.AvatarURL, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	return user, err
}

func (r *repository) GetByPhone(phone string) (*User, error) {
	user := &User{}
	query := `SELECT id, COALESCE(email,''), password_hash, first_name, COALESCE(middle_name,''), last_name, phone, COALESCE(avatar_url,''), created_at
			  FROM users WHERE phone = $1`
	err := r.db.QueryRow(query, phone).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.MiddleName, &user.LastName, &user.Phone, &user.AvatarURL, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	return user, err
}

func (r *repository) GetByID(id string) (*User, error) {
	user := &User{}
	query := `SELECT id, COALESCE(email,''), password_hash, first_name, COALESCE(middle_name,''), last_name, phone, COALESCE(avatar_url,''), created_at
			  FROM users WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.MiddleName, &user.LastName, &user.Phone, &user.AvatarURL, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, core.ErrNotFound
	}
	return user, err
}

func (r *repository) Update(user *User) error {
	query := `UPDATE users SET first_name=$1, middle_name=$2, last_name=$3, phone=$4, password_hash=$5 WHERE id=$6`
	_, err := r.db.Exec(query, user.FirstName, user.MiddleName, user.LastName, user.Phone, user.PasswordHash, user.ID)
	return err
}

func (r *repository) UpdateAvatar(userID, avatarURL string) error {
	_, err := r.db.Exec(`UPDATE users SET avatar_url=$1 WHERE id=$2`, avatarURL, userID)
	return err
}

func (r *repository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *repository) GetPreferences(userID string) (*UserPreferences, error) {
	p := &UserPreferences{}
	err := r.db.QueryRow(
		`SELECT user_id, sms_detection, analytics, partner_offers,
		        transaction_alerts, budget_warnings, ai_insights,
		        weekly_report, savings_reminders, promotions,
		        hide_balances, crash_reports
		 FROM user_preferences WHERE user_id = $1`,
		userID,
	).Scan(&p.UserID, &p.SMSDetection, &p.Analytics, &p.PartnerOffers,
		&p.TransactionAlerts, &p.BudgetWarnings, &p.AIInsights,
		&p.WeeklyReport, &p.SavingsReminders, &p.Promotions,
		&p.HideBalances, &p.CrashReports)
	if err != nil {
		return &UserPreferences{
			UserID: userID, SMSDetection: true, Analytics: true,
			TransactionAlerts: true, BudgetWarnings: true, AIInsights: true,
			SavingsReminders: true, CrashReports: true,
		}, nil
	}
	return p, nil
}

func (r *repository) SavePreferences(p *UserPreferences) error {
	_, err := r.db.Exec(
		`INSERT INTO user_preferences
		 (user_id, sms_detection, analytics, partner_offers,
		  transaction_alerts, budget_warnings, ai_insights,
		  weekly_report, savings_reminders, promotions,
		  hide_balances, crash_reports, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())
		 ON CONFLICT (user_id) DO UPDATE SET
		  sms_detection=$2, analytics=$3, partner_offers=$4,
		  transaction_alerts=$5, budget_warnings=$6, ai_insights=$7,
		  weekly_report=$8, savings_reminders=$9, promotions=$10,
		  hide_balances=$11, crash_reports=$12, updated_at=NOW()`,
		p.UserID, p.SMSDetection, p.Analytics, p.PartnerOffers,
		p.TransactionAlerts, p.BudgetWarnings, p.AIInsights,
		p.WeeklyReport, p.SavingsReminders, p.Promotions,
		p.HideBalances, p.CrashReports,
	)
	return err
}

func (r *repository) GetLinkedAccounts(userID string, limit, offset int) ([]*LinkedAccount, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, bank_name, account_type, account_number, balance, status, last_sync, created_at
		 FROM linked_accounts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []*LinkedAccount
	for rows.Next() {
		a := &LinkedAccount{}
		if err := rows.Scan(&a.ID, &a.UserID, &a.BankName, &a.AccountType, &a.AccountNumber, &a.Balance, &a.Status, &a.LastSync, &a.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *repository) RemoveLinkedAccount(userID, accountID string) error {
	_, err := r.db.Exec(`DELETE FROM linked_accounts WHERE id = $1 AND user_id = $2`, accountID, userID)
	return err
}

func (r *repository) SyncLinkedAccount(userID, accountID string) error {
	_, err := r.db.Exec(
		`UPDATE linked_accounts SET last_sync = NOW(), status = 'active' WHERE id = $1 AND user_id = $2`,
		accountID, userID,
	)
	return err
}

func (r *repository) GetSessions(userID string, limit, offset int) ([]*UserSession, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, token_jti,
		        COALESCE(device, ''), COALESCE(device_type, ''), COALESCE(os, ''),
		        COALESCE(app_version, ''), COALESCE(ip_address, ''),
		        created_at, expires_at
		 FROM user_sessions WHERE user_id = $1 AND expires_at > NOW() ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []*UserSession
	for rows.Next() {
		s := &UserSession{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.TokenJTI, &s.Device, &s.DeviceType, &s.OS, &s.AppVersion, &s.IPAddress, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *repository) RevokeAllSessions(userID string) error {
	_, err := r.db.Exec(`DELETE FROM user_sessions WHERE user_id = $1`, userID)
	return err
}

func (r *repository) RevokeSession(userID, sessionID string) error {
	_, err := r.db.Exec(`DELETE FROM user_sessions WHERE id = $1 AND user_id = $2`, sessionID, userID)
	return err
}

func (r *repository) CreateSession(session *UserSession) error {
	_, err := r.db.Exec(
		`INSERT INTO user_sessions (user_id, token_jti, device, device_type, os, app_version, ip_address, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		session.UserID, session.TokenJTI, session.Device, session.DeviceType, session.OS, session.AppVersion, session.IPAddress, session.ExpiresAt,
	)
	return err
}

func (r *repository) GetSessionByJTI(jti string) (*UserSession, error) {
	s := &UserSession{}
	err := r.db.QueryRow(
		`SELECT id, user_id, token_jti, device, ip_address, created_at, expires_at
		 FROM user_sessions WHERE token_jti = $1`, jti,
	).Scan(&s.ID, &s.UserID, &s.TokenJTI, &s.Device, &s.IPAddress, &s.CreatedAt, &s.ExpiresAt)
	return s, err
}

func (r *repository) DeleteSessionByJTI(jti string) error {
	_, err := r.db.Exec(`DELETE FROM user_sessions WHERE token_jti = $1`, jti)
	return err
}

func (r *repository) DeleteExpiredSessions() error {
	_, err := r.db.Exec(`DELETE FROM user_sessions WHERE expires_at < NOW()`)
	return err
}

var _ = time.Now // keep time import used

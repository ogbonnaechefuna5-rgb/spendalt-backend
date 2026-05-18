// Package lang centralises every user-facing string so they can be updated,
// translated, or A/B tested without touching business logic.
package lang

// ── Auth ─────────────────────────────────────────────────────────────────────

const (
	ErrInvalidBody    = "invalid request body"
	ErrRefreshRequired = "refresh_token required"

	// Registration
	ErrFirstNameRequired  = "first name is required"
	ErrLastNameRequired   = "last name is required"
	ErrPhoneRequired      = "phone number is required"
	ErrPhoneInvalid       = "please enter a valid phone number"
	ErrPasswordRequired   = "password is required"
	ErrPasswordTooShort   = "password must be at least 8 characters"
	ErrEmailInvalid       = "please enter a valid email address"
	ErrPhoneTaken         = "an account with this phone number already exists"
	ErrEmailTaken         = "an account with this email already exists"
	ErrPasswordHash       = "failed to process password"
	ErrCreateAccount      = "failed to create account, please try again"

	// Login
	ErrIdentifierRequired  = "phone number or email is required"
	ErrAccountLocked       = "account temporarily locked due to too many failed attempts, try again later"
	ErrInvalidCredentials  = "incorrect phone number or password"
	ErrInvalidRefreshToken = "invalid or expired refresh token"
)

// ── User / Profile ────────────────────────────────────────────────────────────

const (
	ErrCurrentPasswordRequired = "current password is required"
	ErrNewPasswordRequired     = "new password is required"
	ErrNewPasswordTooShort     = "new password must be at least 8 characters"
	ErrPasswordSameAsOld       = "new password must be different from your current password"
	ErrCurrentPasswordWrong    = "current password is incorrect"
)

// ── Transactions ──────────────────────────────────────────────────────────────

const (
	ErrSMSTextRequired  = "sms_text is required"
	ErrSMSTextTooLong   = "sms_text is too long"
	ErrAmountRequired   = "amount must be greater than 0"
	ErrAmountTooLarge   = "amount exceeds maximum allowed value"
	ErrTypeRequired     = "type is required"
	ErrTypeInvalid      = "type must be 'debit' or 'credit'"
	ErrMerchantRequired = "merchant is required"
	ErrMerchantTooLong  = "merchant name is too long"
	ErrDescriptionTooLong = "description is too long"
	ErrDuplicateTransaction = "duplicate transaction"
)

// ── Budget ────────────────────────────────────────────────────────────────────

const (
	ErrBudgetCategoryRequired = "category is required"
	ErrBudgetAmountRequired   = "amount must be greater than 0"
	ErrBudgetPeriodRequired   = "period is required"
	ErrBudgetPeriodInvalid    = "period must be one of: daily, weekly, monthly, yearly"
)

// ── Savings ───────────────────────────────────────────────────────────────────

const (
	ErrGoalNameRequired   = "goal name is required"
	ErrGoalNameTooLong    = "goal name is too long"
	ErrTargetRequired     = "target amount must be greater than 0"
	ErrTargetTooLarge     = "target amount exceeds maximum allowed value"
)

// ── Core / Generic ────────────────────────────────────────────────────────────

const (
	ErrNotFound    = "resource not found"
	ErrInternal    = "internal server error"
	ErrUnauthorized = "unauthorized"
	ErrInvalidToken = "invalid token"
	ErrTokenRevoked = "token has been revoked"
)

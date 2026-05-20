package user

import "github.com/moninte/backend/internal/core"

// ProfileResponse wraps the user profile for GET /user/profile.
type ProfileResponse struct {
	User *User `json:"user"`
}

// PreferencesResponse wraps preferences for GET /user/preferences.
type PreferencesResponse struct {
	Preferences *UserPreferences `json:"preferences"`
}

// LinkedAccountsResponse wraps a paginated list of linked accounts.
type LinkedAccountsResponse struct {
	Accounts []*LinkedAccount `json:"accounts"`
	core.PageMeta
}

// SessionsResponse wraps a paginated list of sessions.
type SessionsResponse struct {
	Sessions []*UserSession `json:"sessions"`
	core.PageMeta
}

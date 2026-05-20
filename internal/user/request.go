package user

// UpdateProfileRequest is the body for PUT /user/profile.
type UpdateProfileRequest struct {
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Phone      string `json:"phone"`
}

// ChangePasswordRequest is the body for POST /user/change-password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// SavePreferencesRequest is the body for PUT /user/preferences.
type SavePreferencesRequest struct {
	SMSDetection      bool `json:"sms_detection"`
	Analytics         bool `json:"analytics"`
	PartnerOffers     bool `json:"partner_offers"`
	TransactionAlerts bool `json:"transaction_alerts"`
	BudgetWarnings    bool `json:"budget_warnings"`
	AIInsights        bool `json:"ai_insights"`
	WeeklyReport      bool `json:"weekly_report"`
	SavingsReminders  bool `json:"savings_reminders"`
	Promotions        bool `json:"promotions"`
	HideBalances      bool `json:"hide_balances"`
	CrashReports      bool `json:"crash_reports"`
}

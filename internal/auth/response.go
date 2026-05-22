package auth

// UserPayload is the user object embedded in auth responses.
type UserPayload struct {
	ID         string `json:"id"`
	Email      string `json:"email,omitempty"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name,omitempty"`
	LastName   string `json:"last_name"`
}

// SignupResponse is the body returned by POST /auth/signup.
type SignupResponse struct {
	Message string      `json:"message"`
	User    UserPayload `json:"user"`
}

// LoginResponse is the body returned by POST /auth/login.
type LoginResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserPayload `json:"user"`
}

// RefreshResponse is the body returned by POST /auth/refresh.
type RefreshResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// OIDCResponse is the body returned by POST /auth/oidc.
type OIDCResponse struct {
	Token        string      `json:"token"`
	RefreshToken string      `json:"refresh_token"`
	User         UserPayload `json:"user"`
}

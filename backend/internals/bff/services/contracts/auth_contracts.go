package contracts

// AuthUser summarizes the authenticated caller for the frontend session state.
type AuthUser struct {
	ID       string
	Username string
	Email    string
}

// AuthActiveProject summarizes the active tenant context returned after login.
type AuthActiveProject struct {
	ID   string
	Name string
	Role string
}

// AuthSessionResponse is the BFF service-layer login contract.
type AuthSessionResponse struct {
	AccessToken   string
	ExpiresIn     int
	RefreshAt     int
	CSRFToken     string
	User          AuthUser
	ActiveProject *AuthActiveProject
}

// RefreshSessionResponse is the BFF service-layer refresh contract.
type RefreshSessionResponse struct {
	AccessToken string
	ExpiresIn   int
	RefreshAt   int
	CSRFToken   string
}

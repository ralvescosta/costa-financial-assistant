package views

// LoginInput carries the username/password credentials for the login flow.
type LoginInput struct {
	Body struct {
		Username string `json:"username" doc:"Seeded username used to sign in" validate:"required,min=3,max=100"`
		Password string `json:"password" doc:"Seeded password used to sign in" validate:"required,min=8,max=255"`
	}
}

// RefreshInput carries the optional bearer/cookie session token headers for refresh.
type RefreshInput struct {
	Authorization string `header:"Authorization" doc:"Optional Bearer access token"`
	Cookie        string `header:"Cookie" doc:"Session cookie header"`
}

// UserSummaryResponse is the authenticated user shape returned to the frontend.
type UserSummaryResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// ProjectSummaryResponse is the active project context returned after login.
type ProjectSummaryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// LoginSuccessData is the session payload consumed by the frontend auth hook.
type LoginSuccessData struct {
	ExpiresIn     int                     `json:"expiresIn"`
	RefreshAt     int                     `json:"refreshAt"`
	CSRFToken     string                  `json:"csrfToken"`
	User          UserSummaryResponse     `json:"user"`
	ActiveProject *ProjectSummaryResponse `json:"activeProject,omitempty"`
}

// LoginResponse is the JSON envelope returned by POST /api/auth/login.
type LoginResponse struct {
	StatusCode int              `json:"statusCode"`
	Data       LoginSuccessData `json:"data"`
}

// LoginOutput returns the session envelope and sets the HTTP-only session cookie.
type LoginOutput struct {
	SetCookie string        `header:"Set-Cookie"`
	Body      LoginResponse `json:"body"`
}

// RefreshSuccessData is the session metadata returned after token rotation.
type RefreshSuccessData struct {
	ExpiresIn int    `json:"expiresIn"`
	RefreshAt int    `json:"refreshAt"`
	CSRFToken string `json:"csrfToken"`
}

// RefreshResponse is the JSON envelope returned by POST /api/auth/refresh.
type RefreshResponse struct {
	StatusCode int                `json:"statusCode"`
	Data       RefreshSuccessData `json:"data"`
}

// RefreshOutput returns the rotated session metadata and cookie header.
type RefreshOutput struct {
	SetCookie string          `header:"Set-Cookie"`
	Body      RefreshResponse `json:"body"`
}

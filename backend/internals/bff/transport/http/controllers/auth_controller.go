package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	controllermappers "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/controllers/mappers"
	views "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/views"
)

const sessionCookieName = "cfa_session"

// AuthController handles the login/refresh HTTP endpoints for the BFF.
type AuthController struct {
	BaseController
	svc bffinterfaces.AuthService
}

// NewAuthController constructs an AuthController.
func NewAuthController(logger *zap.Logger, validate *validator.Validate, svc bffinterfaces.AuthService) *AuthController {
	return &AuthController{BaseController: BaseController{logger: logger, validate: validate}, svc: svc}
}

// HandleLogin validates the submitted credentials and returns a session envelope.
func (c *AuthController) HandleLogin(ctx context.Context, input *views.LoginInput) (*views.LoginOutput, error) {
	if err := c.validateInput(input); err != nil {
		return nil, err
	}
	username, password := controllermappers.ToLoginCredentials(input)
	resp, err := c.svc.Login(ctx, username, password)
	if err != nil {
		return nil, c.grpcToHumaError(err, "login failed")
	}
	cookie := buildSessionCookie(resp.AccessToken, resp.ExpiresIn)
	return controllermappers.ToLoginOutput(resp, cookie), nil
}

// HandleRefresh rotates the current session token using the existing cookie or bearer token.
func (c *AuthController) HandleRefresh(ctx context.Context, input *views.RefreshInput) (*views.RefreshOutput, error) {
	token := extractSessionToken(input)
	if token == "" {
		return nil, huma.Error401Unauthorized("missing session token")
	}
	resp, err := c.svc.Refresh(ctx, token)
	if err != nil {
		return nil, c.grpcToHumaError(err, "refresh failed")
	}
	cookie := buildSessionCookie(resp.AccessToken, resp.ExpiresIn)
	return controllermappers.ToRefreshOutput(resp, cookie), nil
}

func buildSessionCookie(token string, expiresIn int) string {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
		MaxAge:   expiresIn,
		Expires:  time.Now().UTC().Add(time.Duration(expiresIn) * time.Second),
	}
	return cookie.String()
}

func extractSessionToken(input *views.RefreshInput) string {
	if input == nil {
		return ""
	}
	if authHeader := strings.TrimSpace(input.Authorization); authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	for _, part := range strings.Split(input.Cookie, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, value, found := strings.Cut(part, "=")
		if found && strings.TrimSpace(name) == sessionCookieName {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

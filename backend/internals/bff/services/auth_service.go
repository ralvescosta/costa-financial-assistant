package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"go.uber.org/zap"

	bffinterfaces "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/interfaces"
	bffcontracts "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/services/contracts"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

const authRefreshLeadSeconds = 300

// AuthServiceImpl implements bffinterfaces.AuthService using the identity gRPC client.
type AuthServiceImpl struct {
	logger         *zap.Logger
	identityClient bffinterfaces.IdentityClient
}

// NewAuthService constructs an AuthServiceImpl.
func NewAuthService(logger *zap.Logger, identityClient bffinterfaces.IdentityClient) bffinterfaces.AuthService {
	return &AuthServiceImpl{logger: logger, identityClient: identityClient}
}

// Login validates credentials via the identity service and returns the session payload expected by the frontend.
func (s *AuthServiceImpl) Login(ctx context.Context, username, password string) (*bffcontracts.AuthSessionResponse, error) {
	resp, err := s.identityClient.AuthenticateUser(ctx, &identityv1.AuthenticateUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		s.logger.Error("auth_svc: login failed", zap.String("username", username), zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	return mapAuthSessionResponse(resp.GetToken(), resp.GetExpiresAt(), resp.GetClaims(), resp.GetProjectName())
}

// Refresh validates the current session token and returns a rotated session payload.
func (s *AuthServiceImpl) Refresh(ctx context.Context, token string) (*bffcontracts.RefreshSessionResponse, error) {
	resp, err := s.identityClient.RefreshSession(ctx, &identityv1.RefreshSessionRequest{Token: token})
	if err != nil {
		s.logger.Error("auth_svc: refresh failed", zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return nil, appErr
		}
		return nil, apperrors.TranslateError(err, "service")
	}

	expiresIn, refreshAt := sessionTiming(resp.GetExpiresAt())
	csrfToken, err := newCSRFToken()
	if err != nil {
		s.logger.Error("auth_svc: csrf generation failed during refresh", zap.Error(err))
		return nil, apperrors.TranslateError(err, "service")
	}

	return &bffcontracts.RefreshSessionResponse{
		AccessToken: resp.GetToken(),
		ExpiresIn:   expiresIn,
		RefreshAt:   refreshAt,
		CSRFToken:   csrfToken,
	}, nil
}

func mapAuthSessionResponse(token string, expiresAt int64, claims *identityv1.JwtClaims, projectName string) (*bffcontracts.AuthSessionResponse, error) {
	csrfToken, err := newCSRFToken()
	if err != nil {
		return nil, apperrors.TranslateError(err, "service")
	}

	expiresIn, refreshAt := sessionTiming(expiresAt)
	response := &bffcontracts.AuthSessionResponse{
		AccessToken: token,
		ExpiresIn:   expiresIn,
		RefreshAt:   refreshAt,
		CSRFToken:   csrfToken,
	}
	if claims != nil {
		response.User = bffcontracts.AuthUser{
			ID:       claims.GetSubject(),
			Username: claims.GetUsername(),
			Email:    claims.GetEmail(),
		}
		if claims.GetProjectId() != "" || projectName != "" {
			response.ActiveProject = &bffcontracts.AuthActiveProject{
				ID:   claims.GetProjectId(),
				Name: projectName,
				Role: claims.GetRole(),
			}
		}
	}
	return response, nil
}

func sessionTiming(expiresAt int64) (int, int) {
	expiresIn := int(expiresAt - time.Now().UTC().Unix())
	if expiresIn < 60 {
		expiresIn = 60
	}
	refreshAt := expiresIn - authRefreshLeadSeconds
	if refreshAt < 60 {
		refreshAt = expiresIn / 2
		if refreshAt < 60 {
			refreshAt = 60
		}
	}
	return expiresIn, refreshAt
}

func newCSRFToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

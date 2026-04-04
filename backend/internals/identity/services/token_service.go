package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	identityrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/repositories"
	apperrors "github.com/ralvescosta/costa-financial-assistant/backend/pkgs/errors"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

const (
	tokenTTL = 24 * time.Hour
	keyID    = "bootstrap-key-v1"

	defaultBootstrapUserID       = "00000000-0000-0000-0000-000000000001"
	defaultBootstrapProjectID    = "00000000-0000-0000-0000-000000000010"
	defaultBootstrapUsername     = "ralvescosta"
	defaultBootstrapEmail        = "ralvescosta@local.dev"
	defaultBootstrapRole         = "write"
	defaultBootstrapProjectName  = "Costa Financial Assistant"
	defaultBootstrapPasswordHash = "$2a$10$AjPfEDzY4NI/NhnKuN9UEu6X6J6zRUNO2e79dfh3E1VbdkIpYHzcy"
)

// TokenServiceIface is the narrow interface consumed by the gRPC server.
// Pointer policy: JWT/JWKS struct payloads are exposed through pointer signatures.
type TokenServiceIface interface {
	AuthenticateUser(ctx context.Context, username, password string) (string, int64, *identityv1.JwtClaims, string, error)
	RefreshSession(ctx context.Context, token string) (string, int64, *identityv1.JwtClaims, string, error)
	IssueBootstrapToken(ctx context.Context, userID, projectID, role string) (string, int64, error)
	ValidateToken(ctx context.Context, token string) (bool, *identityv1.JwtClaims, error)
	GetJwksMetadata(ctx context.Context) (*identityv1.JwksMetadata, error)
}

// bootstrapAuthLookup resolves the seeded owner account used by the login flow.
type bootstrapAuthLookup interface {
	FindBootstrapUser(ctx context.Context, username string) (*identityrepo.BootstrapAuthRecord, error)
}

// TokenService handles JWT signing and JWKS exposition for the identity service.
type TokenService struct {
	authRepo bootstrapAuthLookup
	key      *rsa.PrivateKey
	logger   *zap.Logger
}

// NewTokenService constructs a TokenService with the fallback seeded bootstrap identity.
func NewTokenService(key *rsa.PrivateKey, logger *zap.Logger) TokenServiceIface {
	return &TokenService{authRepo: fallbackBootstrapRepo{}, key: key, logger: logger}
}

// NewTokenServiceWithRepository constructs a TokenService backed by the identity database.
func NewTokenServiceWithRepository(repo bootstrapAuthLookup, key *rsa.PrivateKey, logger *zap.Logger) TokenServiceIface {
	if repo == nil {
		repo = fallbackBootstrapRepo{}
	}
	return &TokenService{authRepo: repo, key: key, logger: logger}
}

// AuthenticateUser validates the seeded bootstrap credentials and returns a signed session.
func (s *TokenService) AuthenticateUser(ctx context.Context, username, password string) (string, int64, *identityv1.JwtClaims, string, error) {
	if username == "" || password == "" {
		return "", 0, nil, "", apperrors.NewWithCategory("username and password are required", apperrors.CategoryValidation)
	}

	record, err := s.authRepo.FindBootstrapUser(ctx, username)
	if err != nil {
		if appErr := apperrors.AsAppError(err); appErr != nil && appErr.Category == apperrors.CategoryNotFound {
			s.logger.Warn("bootstrap login rejected: user not found", zap.String("username", username))
			return "", 0, nil, "", apperrors.NewWithCategory("invalid credentials", apperrors.CategoryAuth)
		}
		s.logger.Error("bootstrap login lookup failed", zap.String("username", username), zap.Error(err))
		if appErr := apperrors.AsAppError(err); appErr != nil {
			return "", 0, nil, "", appErr
		}
		return "", 0, nil, "", apperrors.TranslateError(err, "service")
	}

	if record == nil || record.PasswordHash == "" {
		s.logger.Error("bootstrap login failed: credentials not seeded", zap.String("username", username))
		return "", 0, nil, "", apperrors.NewWithCategory("bootstrap credentials are not configured", apperrors.CategoryDependencyDB)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)); err != nil {
		s.logger.Warn("bootstrap login rejected: password mismatch", zap.String("username", username))
		return "", 0, nil, "", apperrors.NewWithCategory("invalid credentials", apperrors.CategoryAuth)
	}

	claims := &identityv1.JwtClaims{
		Subject:   record.UserID,
		ProjectId: record.ProjectID,
		Role:      defaultIfEmpty(record.Role, defaultBootstrapRole),
		Email:     defaultIfEmpty(record.Email, defaultBootstrapEmail),
		Username:  defaultIfEmpty(record.Username, defaultBootstrapUsername),
	}

	token, expiresAt, signedClaims, err := s.issueSignedToken(claims)
	if err != nil {
		return "", 0, nil, "", err
	}

	projectName := defaultIfEmpty(record.ProjectName, defaultBootstrapProjectName)
	s.logger.Info("bootstrap login authenticated",
		zap.String("user_id", signedClaims.GetSubject()),
		zap.String("project_id", signedClaims.GetProjectId()),
		zap.String("username", signedClaims.GetUsername()))
	return token, expiresAt, signedClaims, projectName, nil
}

// RefreshSession validates the existing token and reissues a new signed session.
func (s *TokenService) RefreshSession(ctx context.Context, token string) (string, int64, *identityv1.JwtClaims, string, error) {
	if token == "" {
		return "", 0, nil, "", apperrors.NewWithCategory("session token is required", apperrors.CategoryValidation)
	}

	valid, claims, err := s.ValidateToken(ctx, token)
	if err != nil {
		return "", 0, nil, "", err
	}
	if !valid || claims == nil {
		return "", 0, nil, "", apperrors.NewWithCategory("session expired", apperrors.CategoryAuth)
	}

	refreshedToken, expiresAt, signedClaims, err := s.issueSignedToken(&identityv1.JwtClaims{
		Subject:   claims.GetSubject(),
		ProjectId: claims.GetProjectId(),
		Role:      claims.GetRole(),
		Email:     claims.GetEmail(),
		Username:  claims.GetUsername(),
	})
	if err != nil {
		return "", 0, nil, "", err
	}

	return refreshedToken, expiresAt, signedClaims, defaultBootstrapProjectName, nil
}

// IssueBootstrapToken signs and returns a JWT for the given bootstrap context.
func (s *TokenService) IssueBootstrapToken(_ context.Context, userID, projectID, role string) (string, int64, error) {
	claims := &identityv1.JwtClaims{
		Subject:   userID,
		ProjectId: projectID,
		Role:      role,
		Email:     defaultBootstrapEmail,
		Username:  defaultBootstrapUsername,
	}

	signed, expiresAt, _, err := s.issueSignedToken(claims)
	if err != nil {
		return "", 0, err
	}

	s.logger.Info("bootstrap token issued",
		zap.String("user_id", userID),
		zap.String("project_id", projectID),
		zap.String("role", role),
	)
	return signed, expiresAt, nil
}

// ValidateToken parses and verifies the JWT, returning decoded claims on success.
func (s *TokenService) ValidateToken(_ context.Context, tokenStr string) (bool, *identityv1.JwtClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			s.logger.Debug("token validation rejected unexpected signing method", zap.Any("alg", t.Header["alg"]))
			return nil, apperrors.NewCatalogError(apperrors.ErrUnauthorized)
		}
		return &s.key.PublicKey, nil
	})
	if err != nil || !token.Valid {
		if err != nil {
			s.logger.Debug("token validation failed", zap.Error(err))
		}
		return false, nil, nil //nolint:nilerr // invalid token is not a processing error
	}

	mc, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil, nil
	}

	claims := &identityv1.JwtClaims{
		Subject:   stringClaim(mc, "sub"),
		ProjectId: stringClaim(mc, "project_id"),
		Role:      stringClaim(mc, "role"),
		IssuedAt:  int64Claim(mc, "iat"),
		ExpiresAt: int64Claim(mc, "exp"),
		Email:     stringClaim(mc, "email"),
		Username:  stringClaim(mc, "username"),
	}
	return true, claims, nil
}

// GetJwksMetadata returns the public JWKS representation of the service signing key.
func (s *TokenService) GetJwksMetadata(_ context.Context) (*identityv1.JwksMetadata, error) {
	pub := &s.key.PublicKey
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	return &identityv1.JwksMetadata{
		Keys: []*identityv1.JwksKey{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: keyID,
				Alg: "RS256",
				N:   n,
				E:   e,
			},
		},
	}, nil
}

func (s *TokenService) issueSignedToken(claims *identityv1.JwtClaims) (string, int64, *identityv1.JwtClaims, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(tokenTTL).Unix()

	signedClaims := &identityv1.JwtClaims{
		Subject:   claims.GetSubject(),
		ProjectId: claims.GetProjectId(),
		Role:      claims.GetRole(),
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt,
		Email:     claims.GetEmail(),
		Username:  claims.GetUsername(),
	}

	jwtClaims := jwt.MapClaims{
		"sub":        signedClaims.GetSubject(),
		"project_id": signedClaims.GetProjectId(),
		"role":       signedClaims.GetRole(),
		"iat":        signedClaims.GetIssuedAt(),
		"exp":        signedClaims.GetExpiresAt(),
		"email":      signedClaims.GetEmail(),
		"username":   signedClaims.GetUsername(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	t.Header["kid"] = keyID

	signedToken, err := t.SignedString(s.key)
	if err != nil {
		s.logger.Error("token signing failed", zap.String("user_id", claims.GetSubject()), zap.Error(err))
		return "", 0, nil, apperrors.TranslateError(err, "service")
	}

	return signedToken, expiresAt, signedClaims, nil
}

func stringClaim(mc jwt.MapClaims, key string) string {
	v, _ := mc[key].(string)
	return v
}

func int64Claim(mc jwt.MapClaims, key string) int64 {
	switch v := mc[key].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	}
	return 0
}

func defaultIfEmpty(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

type fallbackBootstrapRepo struct{}

func (fallbackBootstrapRepo) FindBootstrapUser(_ context.Context, username string) (*identityrepo.BootstrapAuthRecord, error) {
	if username != defaultBootstrapUsername {
		return nil, apperrors.NewCatalogError(apperrors.ErrResourceNotFound)
	}
	return &identityrepo.BootstrapAuthRecord{
		UserID:       defaultBootstrapUserID,
		ProjectID:    defaultBootstrapProjectID,
		Username:     defaultBootstrapUsername,
		Email:        defaultBootstrapEmail,
		PasswordHash: defaultBootstrapPasswordHash,
		Role:         defaultBootstrapRole,
		ProjectName:  defaultBootstrapProjectName,
	}, nil
}

// ensure rand is imported for RSA key generation in container.go
var _ = rand.Reader

package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

const (
	tokenTTL = 24 * time.Hour
	keyID    = "bootstrap-key-v1"
)

// TokenServiceIface is the narrow interface consumed by the gRPC server.
type TokenServiceIface interface {
	IssueBootstrapToken(ctx context.Context, userID, projectID, role string) (string, int64, error)
	ValidateToken(ctx context.Context, token string) (bool, *identityv1.JwtClaims, error)
	GetJwksMetadata(ctx context.Context) (*identityv1.JwksMetadata, error)
}

// TokenService handles JWT signing and JWKS exposition for the identity service.
type TokenService struct {
	key    *rsa.PrivateKey
	logger *zap.Logger
}

// NewTokenService constructs a TokenService with the provided RSA private key.
func NewTokenService(key *rsa.PrivateKey, logger *zap.Logger) TokenServiceIface {
	return &TokenService{key: key, logger: logger}
}

// IssueBootstrapToken signs and returns a JWT for the given bootstrap context.
func (s *TokenService) IssueBootstrapToken(_ context.Context, userID, projectID, role string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(tokenTTL)
	claims := jwt.MapClaims{
		"sub":        userID,
		"project_id": projectID,
		"role":       role,
		"iat":        now.Unix(),
		"exp":        exp.Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t.Header["kid"] = keyID

	signed, err := t.SignedString(s.key)
	if err != nil {
		s.logger.Error("token signing failed", zap.String("user_id", userID), zap.Error(err))
		return "", 0, fmt.Errorf("token service: sign: %w", err)
	}

	s.logger.Info("bootstrap token issued",
		zap.String("user_id", userID),
		zap.String("project_id", projectID),
		zap.String("role", role),
	)
	return signed, exp.Unix(), nil
}

// ValidateToken parses and verifies the JWT, returning decoded claims on success.
func (s *TokenService) ValidateToken(_ context.Context, tokenStr string) (bool, *identityv1.JwtClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return &s.key.PublicKey, nil
	})
	if err != nil || !token.Valid {
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

// ensure rand is imported for RSA key generation in container.go
var _ = rand.Reader

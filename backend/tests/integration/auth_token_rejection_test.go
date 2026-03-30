//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
)

func newTestLogger(t *testing.T) *zap.Logger {
	t.Helper()
	return zaptest.NewLogger(t)
}

// TestAuthTokenRejectionMatrix verifies that the token service rejects invalid,
// expired, and tampered tokens while accepting valid ones.
func TestAuthTokenRejectionMatrix(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	logger := newTestLogger(t)
	svc := services.NewTokenService(key, logger)

	// Issue a valid token to use as baseline
	validToken, _, err := svc.IssueBootstrapToken(
		context.Background(),
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000010",
		"write",
	)
	require.NoError(t, err)

	// Build an expired token signed with the correct key
	expiredClaims := jwt.MapClaims{
		"sub":        "user-1",
		"project_id": "proj-1",
		"role":       "read_only",
		"iat":        time.Now().Add(-48 * time.Hour).Unix(),
		"exp":        time.Now().Add(-24 * time.Hour).Unix(),
	}
	expiredTok := jwt.NewWithClaims(jwt.SigningMethodRS256, expiredClaims)
	expiredTok.Header["kid"] = "bootstrap-key-v1"
	expiredToken, err := expiredTok.SignedString(key)
	require.NoError(t, err)

	// Build a token signed with a different (wrong) key
	wrongKeyClaims := jwt.MapClaims{
		"sub":        "user-1",
		"project_id": "proj-1",
		"role":       "read_only",
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	}
	wrongKeyTok := jwt.NewWithClaims(jwt.SigningMethodRS256, wrongKeyClaims)
	wrongKeyTok.Header["kid"] = "bootstrap-key-v1"
	wrongKeyToken, err := wrongKeyTok.SignedString(otherKey)
	require.NoError(t, err)

	cases := []struct {
		name          string
		token         string
		expectValid   bool
		expectErrNil  bool
	}{
		{
			name:         "valid token is accepted",
			token:        validToken,
			expectValid:  true,
			expectErrNil: true,
		},
		{
			name:         "expired token is rejected",
			token:        expiredToken,
			expectValid:  false,
			expectErrNil: true,
		},
		{
			name:         "token signed with wrong key is rejected",
			token:        wrongKeyToken,
			expectValid:  false,
			expectErrNil: true,
		},
		{
			name:         "malformed token string is rejected",
			token:        "not.a.jwt",
			expectValid:  false,
			expectErrNil: true,
		},
		{
			name:         "empty token string is rejected",
			token:        "",
			expectValid:  false,
			expectErrNil: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			valid, claims, err := svc.ValidateToken(context.Background(), tc.token)

			if tc.expectErrNil {
				assert.NoError(t, err, fmt.Sprintf("case %q: unexpected error", tc.name))
			}
			assert.Equal(t, tc.expectValid, valid, fmt.Sprintf("case %q: valid mismatch", tc.name))
			if tc.expectValid {
				assert.NotNil(t, claims, fmt.Sprintf("case %q: claims should not be nil on success", tc.name))
			}
		})
	}
}

package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestTokenService_AuthenticateUser_ReturnsBootstrapSession(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	svc := NewTokenService(key, zaptest.NewLogger(t))
	token, expiresAt, claims, projectName, err := svc.AuthenticateUser(context.Background(), "ralvescosta", "mudar@1234")

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, expiresAt, int64(0))
	require.NotNil(t, claims)
	assert.Equal(t, "ralvescosta", claims.GetUsername())
	assert.Equal(t, "write", claims.GetRole())
	assert.NotEmpty(t, projectName)
}

func TestTokenService_AuthenticateUser_RejectsInvalidCredentials(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	svc := NewTokenService(key, zaptest.NewLogger(t))
	token, expiresAt, claims, projectName, err := svc.AuthenticateUser(context.Background(), "ralvescosta", "wrong-password")

	require.Error(t, err)
	assert.Empty(t, token)
	assert.Zero(t, expiresAt)
	assert.Nil(t, claims)
	assert.Empty(t, projectName)
}

func TestTokenService_RefreshSession_ReissuesTokenWithSessionClaims(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	svc := NewTokenService(key, zaptest.NewLogger(t))
	token, _, _, _, err := svc.AuthenticateUser(context.Background(), "ralvescosta", "mudar@1234")
	require.NoError(t, err)

	refreshed, expiresAt, claims, _, err := svc.RefreshSession(context.Background(), token)
	require.NoError(t, err)
	assert.NotEmpty(t, refreshed)
	assert.Greater(t, expiresAt, int64(0))
	require.NotNil(t, claims)
	assert.Equal(t, "ralvescosta", claims.GetUsername())
	assert.Equal(t, "ralvescosta@local.dev", claims.GetEmail())
}

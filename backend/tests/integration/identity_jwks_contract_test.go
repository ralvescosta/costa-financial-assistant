//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
)

// TestIdentityJWKSMetadataContract verifies that the identity token service
// returns a valid JWKS payload containing the expected key fields.
func TestIdentityJWKSMetadataContract(t *testing.T) {
	// Arrange
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "RSA key generation must succeed")

	logger := newTestLogger(t)
	svc := services.NewTokenService(key, logger)

	// Act
	jwks, err := svc.GetJwksMetadata(context.Background())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, jwks)
	assert.Len(t, jwks.Keys, 1, "JWKS should contain exactly one key")

	k := jwks.Keys[0]
	assert.Equal(t, "RSA", k.Kty)
	assert.Equal(t, "sig", k.Use)
	assert.Equal(t, "RS256", k.Alg)
	assert.NotEmpty(t, k.Kid, "kid must be set")
	assert.NotEmpty(t, k.N, "modulus must be base64url encoded")
	assert.NotEmpty(t, k.E, "exponent must be base64url encoded")
}

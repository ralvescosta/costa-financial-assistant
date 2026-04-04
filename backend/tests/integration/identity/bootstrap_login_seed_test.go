//go:build integration

package integration

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	identityrepo "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/repositories"
	identitysvc "github.com/ralvescosta/costa-financial-assistant/backend/internals/identity/services"
	integrationhelpers "github.com/ralvescosta/costa-financial-assistant/backend/tests/integration/helpers"
)

func TestBootstrapLoginSeedIntegration(t *testing.T) {
	// Given a fresh database with the onboarding + identity migrations applied.
	ctx := context.Background()
	resources, err := integrationhelpers.SetupPostgresSuite(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		resources.Close(context.Background())
	})

	require.NoError(t, integrationhelpers.RunMigrations(resources.DSN, "file://internals/onboarding/migrations/ddl"))
	require.NoError(t, integrationhelpers.RunMigrations(resources.DSN, "file://internals/identity/migrations/ddl"))

	logger := zaptest.NewLogger(t)
	repo := identityrepo.NewBootstrapAuthRepository(resources.DB, logger)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	svc := identitysvc.NewTokenServiceWithRepository(repo, key, logger)

	// When the seeded owner signs in with the documented bootstrap credentials.
	token, expiresAt, claims, projectName, err := svc.AuthenticateUser(ctx, "ralvescosta", "mudar@1234")

	// Then the login succeeds with a usable authenticated session envelope.
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, expiresAt, int64(0))
	require.NotNil(t, claims)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", claims.GetSubject())
	assert.Equal(t, "00000000-0000-0000-0000-000000000010", claims.GetProjectId())
	assert.Equal(t, "ralvescosta", claims.GetUsername())
	assert.Equal(t, "ralvescosta@local.dev", claims.GetEmail())
	assert.Equal(t, "Costa Financial Assistant", projectName)
}

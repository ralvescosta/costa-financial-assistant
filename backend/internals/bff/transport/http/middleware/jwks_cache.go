package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

const jwksCacheTTL = 5 * time.Minute

// JWKSCache fetches and caches RSA public keys from the identity gRPC service.
// Keys are refreshed on use when the cache is stale, with stale-while-revalidate semantics.
type JWKSCache struct {
	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	client    identityv1.IdentityServiceClient
	logger    *zap.Logger
}

// NewJWKSCache constructs a JWKSCache connected to the identity gRPC service.
func NewJWKSCache(logger *zap.Logger, cfg *configs.Config) (*JWKSCache, error) {
	conn, err := grpc.NewClient(
		cfg.Services.IdentityGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("jwks cache: dial identity service: %w", err)
	}

	return &JWKSCache{
		keys:   make(map[string]*rsa.PublicKey),
		client: identityv1.NewIdentityServiceClient(conn),
		logger: logger,
	}, nil
}

// GetKey returns the RSA public key for the given key ID.
// It refreshes the cache when stale, returning the previous value on refresh failure.
func (c *JWKSCache) GetKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	c.mu.RLock()
	key, ok := c.keys[kid]
	stale := time.Since(c.fetchedAt) > jwksCacheTTL
	c.mu.RUnlock()

	if ok && !stale {
		return key, nil
	}

	if err := c.refresh(ctx); err != nil {
		if ok {
			c.logger.Warn("JWKS refresh failed; using stale key",
				zap.String("kid", kid),
				zap.Error(err),
			)
			return key, nil
		}
		return nil, fmt.Errorf("jwks cache: refresh: %w", err)
	}

	c.mu.RLock()
	key, ok = c.keys[kid]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("jwks cache: key not found: %s", kid)
	}
	return key, nil
}

func (c *JWKSCache) refresh(ctx context.Context) error {
	resp, err := c.client.GetJwksMetadata(ctx, &identityv1.GetJwksMetadataRequest{})
	if err != nil {
		return fmt.Errorf("jwks cache: GetJwksMetadata: %w", err)
	}
	if resp.GetJwks() == nil {
		return fmt.Errorf("jwks cache: empty JWKS response")
	}

	newKeys := make(map[string]*rsa.PublicKey, len(resp.Jwks.Keys))
	for _, k := range resp.Jwks.Keys {
		pub, err := rsaPublicKeyFromJWK(k)
		if err != nil {
			c.logger.Warn("skipping malformed JWK", zap.String("kid", k.Kid), zap.Error(err))
			continue
		}
		newKeys[k.Kid] = pub
	}

	c.mu.Lock()
	c.keys = newKeys
	c.fetchedAt = time.Now()
	c.mu.Unlock()

	c.logger.Info("JWKS cache refreshed", zap.Int("key_count", len(newKeys)))
	return nil
}

func rsaPublicKeyFromJWK(k *identityv1.JwksKey) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{N: n, E: e}, nil
}

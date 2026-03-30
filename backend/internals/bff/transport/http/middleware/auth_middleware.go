package middleware

import (
	"context"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

// NewAuthMiddleware returns a Huma middleware that validates Bearer JWTs using the
// JWKS cache and stores the decoded claims in the request context.
func NewAuthMiddleware(cache *JWKSCache, logger *zap.Logger) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			unauthorizedResponse(ctx, "missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorizedResponse(ctx, "invalid Authorization header format")
			return
		}
		tokenStr := parts[1]

		kid, err := extractKID(tokenStr)
		if err != nil {
			logger.Debug("failed to extract kid from token", zap.Error(err))
			unauthorizedResponse(ctx, "malformed token")
			return
		}

		pubKey, err := cache.GetKey(context.Background(), kid)
		if err != nil {
			logger.Warn("JWKS public key lookup failed", zap.String("kid", kid), zap.Error(err))
			unauthorizedResponse(ctx, "unable to resolve signing key")
			return
		}

		claims, err := parseAndValidateToken(tokenStr, pubKey)
		if err != nil {
			logger.Debug("token validation failed", zap.Error(err))
			unauthorizedResponse(ctx, "invalid or expired token")
			return
		}

		newCtx := context.WithValue(ctx.Context(), ProjectContextKey, claims)
		ctx = huma.WithContext(ctx, newCtx)

		next(ctx)
	}
}

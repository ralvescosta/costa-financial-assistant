package middleware

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"

	identityv1 "github.com/ralvescosta/costa-financial-assistant/backend/protos/generated/identity/v1"
)

// allowedRoles maps roles to their precedence level (higher = more privilege).
var allowedRoles = map[string]int{
	"read_only": 1,
	"update":    2,
	"write":     3,
}

// NewProjectGuard returns a Huma middleware that enforces a minimum role requirement.
// Expects validated JWT claims in the request context (populated by NewAuthMiddleware).
func NewProjectGuard(minRole string, logger *zap.Logger) func(huma.Context, func(huma.Context)) {
	minLevel, ok := allowedRoles[minRole]
	if !ok {
		minLevel = 999 // fail closed for unknown roles
	}

	return func(ctx huma.Context, next func(huma.Context)) {
		claims, ok := ctx.Context().Value(ProjectContextKey).(*identityv1.JwtClaims)
		if !ok || claims == nil {
			logger.Warn("project guard: missing claims in context")
			forbiddenResponse(ctx, "authentication required")
			return
		}

		level, ok := allowedRoles[claims.Role]
		if !ok || level < minLevel {
			logger.Warn("project guard: insufficient role",
				zap.String("role", claims.Role),
				zap.String("required", minRole),
				zap.String("project_id", claims.ProjectId),
			)
			forbiddenResponse(ctx, "insufficient permissions")
			return
		}

		next(ctx)
	}
}

// ClaimsFromContext retrieves validated JWT claims from the request context.
// Returns nil if claims are not present (unauthenticated route).
func ClaimsFromContext(ctx context.Context) *identityv1.JwtClaims {
	v, _ := ctx.Value(ProjectContextKey).(*identityv1.JwtClaims)
	return v
}

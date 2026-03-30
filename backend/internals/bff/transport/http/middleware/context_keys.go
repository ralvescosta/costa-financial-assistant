package middleware

import (
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// contextKey is an unexported type so keys never collide with other packages.
type contextKey string

// ProjectContextKey is the request-context key under which validated JWT claims are stored.
const ProjectContextKey contextKey = "project_claims"

// writeErrResponse writes a minimal problem+json body directly to the huma.Context.
// Used in middleware where huma.API is not available.
func writeErrResponse(ctx huma.Context, status int, detail string) {
	body := fmt.Sprintf(`{"status":%d,"detail":%q}`, status, detail)
	ctx.SetHeader("Content-Type", "application/problem+json")
	ctx.SetStatus(status)
	_, _ = ctx.BodyWriter().Write([]byte(body))
}

// unauthorizedResponse sends HTTP 401.
func unauthorizedResponse(ctx huma.Context, detail string) {
	writeErrResponse(ctx, http.StatusUnauthorized, detail)
}

// forbiddenResponse sends HTTP 403.
func forbiddenResponse(ctx huma.Context, detail string) {
	writeErrResponse(ctx, http.StatusForbidden, detail)
}

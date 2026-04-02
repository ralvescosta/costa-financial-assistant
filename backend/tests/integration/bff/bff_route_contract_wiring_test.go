//go:build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	bfftransportroutes "github.com/ralvescosta/costa-financial-assistant/backend/internals/bff/transport/http/routes"
)

// TestBFFRouteContractWiring verifies that each route module constructor accepts
// the capability interface (not a concrete controller type) and produces a value
// that satisfies the routes.Route interface. This test catches interface drift
// early — if a controller method is renamed or its signature changes without
// updating the capability interface, this test will fail at compile time (via the
// var _ assertions in the route files) or at construction time here.
func TestBFFRouteContractWiring(t *testing.T) {
	logger := zap.NewNop()

	t.Run("DocumentsRoute accepts DocumentsCapability", func(t *testing.T) {
		var cap bfftransportroutes.DocumentsCapability = stubDocuments{}
		route := bfftransportroutes.NewDocumentsRoute(cap, logger)
		assert.Implements(t, (*bfftransportroutes.Route)(nil), route)
	})

	t.Run("ProjectsRoute accepts ProjectsCapability", func(t *testing.T) {
		var cap bfftransportroutes.ProjectsCapability = stubProjects{}
		route := bfftransportroutes.NewProjectsRoute(cap, logger)
		assert.Implements(t, (*bfftransportroutes.Route)(nil), route)
	})

	t.Run("SettingsRoute accepts SettingsCapability", func(t *testing.T) {
		var cap bfftransportroutes.SettingsCapability = stubSettings{}
		route := bfftransportroutes.NewSettingsRoute(cap, logger)
		assert.Implements(t, (*bfftransportroutes.Route)(nil), route)
	})

	t.Run("PaymentsRoute accepts PaymentsCapability", func(t *testing.T) {
		var cap bfftransportroutes.PaymentsCapability = stubPayments{}
		route := bfftransportroutes.NewPaymentsRoute(cap, logger)
		assert.Implements(t, (*bfftransportroutes.Route)(nil), route)
	})

	t.Run("ReconciliationRoute accepts ReconciliationCapability", func(t *testing.T) {
		var cap bfftransportroutes.ReconciliationCapability = stubReconciliation{}
		route := bfftransportroutes.NewReconciliationRoute(cap, logger)
		assert.Implements(t, (*bfftransportroutes.Route)(nil), route)
	})

	t.Run("HistoryRoute accepts HistoryCapability", func(t *testing.T) {
		var cap bfftransportroutes.HistoryCapability = stubHistory{}
		route := bfftransportroutes.NewHistoryRoute(cap, logger)
		assert.Implements(t, (*bfftransportroutes.Route)(nil), route)
	})
}

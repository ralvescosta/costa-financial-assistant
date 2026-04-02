//go:build integration

package integration

import (
	"os"
	"testing"
)

// TestMain for the root integration package is intentionally minimal.
// Segment-specific suites under bff/files/payments/cross_service own lifecycle.
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

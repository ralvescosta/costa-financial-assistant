//go:build integration

package helpers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// BDDScenario is the canonical scenario shape for integration tests.
type BDDScenario struct {
	Name   string
	Given  string
	When   string
	Then   string
	Assert func(t *testing.T)
}

// ValidateBDDScenario enforces mandatory BDD fields before scenario execution.
func ValidateBDDScenario(s BDDScenario) error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("scenario name is required")
	}
	if strings.TrimSpace(s.Given) == "" {
		return fmt.Errorf("scenario given is required")
	}
	if strings.TrimSpace(s.When) == "" {
		return fmt.Errorf("scenario when is required")
	}
	if strings.TrimSpace(s.Then) == "" {
		return fmt.Errorf("scenario then is required")
	}
	if s.Assert == nil {
		return fmt.Errorf("scenario assert function is required")
	}

	return nil
}

func TestValidateBDDScenario(t *testing.T) {
	t.Run("GivenMissingName WhenValidate ThenFails", func(t *testing.T) {
		err := ValidateBDDScenario(BDDScenario{
			Given:  "a precondition",
			When:   "an action",
			Then:   "an expectation",
			Assert: func(t *testing.T) {},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "name")
	})

	t.Run("GivenCompleteScenario WhenValidate ThenPasses", func(t *testing.T) {
		err := ValidateBDDScenario(BDDScenario{
			Name:   "GivenReadyState WhenExecute ThenSucceeds",
			Given:  "a ready state",
			When:   "execution occurs",
			Then:   "operation succeeds",
			Assert: func(t *testing.T) {},
		})
		require.NoError(t, err)
	})
}

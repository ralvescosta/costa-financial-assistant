//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOpenAPIOperationMetadataCompleteness ensures every Huma operation block
// in BFF controllers provides required OpenAPI metadata fields.
func TestOpenAPIOperationMetadataCompleteness(t *testing.T) {
	controllerGlob := filepath.Clean("../../internals/bff/transport/http/controllers/*_controller.go")
	files, err := filepath.Glob(controllerGlob)
	require.NoError(t, err)
	require.NotEmpty(t, files, "no controller files found using glob: %s", controllerGlob)

	required := []string{
		"OperationID:",
		"Summary:",
		"Description:",
		"Tags:",
		"Method:",
		"Path:",
	}

	totalOperations := 0
	missingByFile := map[string][]string{}

	for _, filePath := range files {
		contentBytes, readErr := os.ReadFile(filePath)
		require.NoError(t, readErr)
		content := string(contentBytes)

		segments := strings.Split(content, "huma.Operation{")
		if len(segments) <= 1 {
			continue
		}

		for idx := 1; idx < len(segments); idx++ {
			segment := segments[idx]
			end := strings.Index(segment, "}, func")
			if end == -1 {
				end = strings.Index(segment, "}, c.")
			}
			if end == -1 {
				continue
			}
			block := segment[:end]
			totalOperations++

			for _, field := range required {
				if !strings.Contains(block, field) {
					missingByFile[filePath] = append(missingByFile[filePath], field)
				}
			}
		}
	}

	require.Greater(t, totalOperations, 0, "expected at least one Huma operation block")
	assert.Empty(t, missingByFile, "all operation blocks must define required OpenAPI metadata")
}

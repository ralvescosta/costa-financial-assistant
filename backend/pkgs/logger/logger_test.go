package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/ralvescosta/costa-financial-assistant/backend/pkgs/configs"
)

func TestResolveLevel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     string
		want      zapcore.Level
		wantError bool
	}{
		{name: "GivenInfoLevelWhenResolveLevelThenReturnsInfo", input: "info", want: zap.InfoLevel},
		{name: "GivenTraceLevelWhenResolveLevelThenMapsToDebug", input: "trace", want: zap.DebugLevel},
		{name: "GivenUppercaseWarnLevelWhenResolveLevelThenReturnsWarn", input: "WARN", want: zap.WarnLevel},
		{name: "GivenInvalidLevelWhenResolveLevelThenReturnsError", input: "loud", wantError: true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Given a configured log level input.
			// Arrange
			levelInput := testCase.input

			// When the logger package resolves the zap level.
			// Act
			got, err := resolveLevel(levelInput)

			// Then the resolved level matches the expected severity or returns an error.
			// Assert
			if testCase.wantError {
				if err == nil {
					t.Fatalf("expected error for input %q", levelInput)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolve level returned error: %v", err)
			}

			if got != testCase.want {
				t.Fatalf("expected level %v, got %v", testCase.want, got)
			}
		})
	}
}

func TestIsProductionEnvironment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "GivenProductionWhenCheckingEnvironmentThenReturnsTrue", input: "production", want: true},
		{name: "GivenPrdWhenCheckingEnvironmentThenReturnsTrue", input: "prd", want: true},
		{name: "GivenStgWhenCheckingEnvironmentThenReturnsTrue", input: "stg", want: true},
		{name: "GivenDevelopmentWhenCheckingEnvironmentThenReturnsFalse", input: "development", want: false},
		{name: "GivenLocalWhenCheckingEnvironmentThenReturnsFalse", input: "local", want: false},
		{name: "GivenEmptyWhenCheckingEnvironmentThenReturnsFalse", input: "", want: false},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Given an application environment string.
			// Arrange
			env := testCase.input

			// When the logger package classifies the environment.
			// Act
			got := isProductionEnvironment(env)

			// Then the result matches the expected production/developer mode.
			// Assert
			if got != testCase.want {
				t.Fatalf("expected %t, got %t", testCase.want, got)
			}
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		cfg          *configs.Config
		debugEnabled bool
		infoEnabled  bool
		wantError    bool
	}{
		{
			name:         "GivenDevelopmentDebugConfigWhenCreatingLoggerThenDebugIsEnabled",
			cfg:          &configs.Config{ServiceName: "bills", Env: "development", LogLevel: "debug"},
			debugEnabled: true,
			infoEnabled:  true,
		},
		{
			name:         "GivenProductionInfoConfigWhenCreatingLoggerThenDebugIsDisabled",
			cfg:          &configs.Config{ServiceName: "bff", Env: "prd", LogLevel: "info"},
			debugEnabled: false,
			infoEnabled:  true,
		},
		{
			name:      "GivenInvalidLevelWhenCreatingLoggerThenReturnsError",
			cfg:       &configs.Config{ServiceName: "files", Env: "dev", LogLevel: "verbose"},
			wantError: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Given a runtime logging configuration.
			// Arrange
			cfg := testCase.cfg

			// When a logger is created from the shared factory.
			// Act
			logger, err := New(cfg)

			// Then the logger honors the configured level or returns a validation error.
			// Assert
			if testCase.wantError {
				if err == nil {
					t.Fatal("expected logger creation to fail")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error creating logger: %v", err)
			}

			defer func() { _ = logger.Sync() }()

			if logger == nil {
				t.Fatal("expected non-nil logger")
			}

			if got := logger.Core().Enabled(zap.DebugLevel); got != testCase.debugEnabled {
				t.Fatalf("expected debug enabled=%t, got %t", testCase.debugEnabled, got)
			}

			if got := logger.Core().Enabled(zap.InfoLevel); got != testCase.infoEnabled {
				t.Fatalf("expected info enabled=%t, got %t", testCase.infoEnabled, got)
			}
		})
	}
}

// Package otel provides OpenTelemetry bootstrap helpers shared across all services.
package otel

import (
	"context"
	"runtime/debug"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// RegisterServiceMetrics registers the standard per-service health and build-info OTel metrics.
//
// It creates two instruments on the global MeterProvider:
//   - <serviceName>_up (observable gauge, value always 1.0) — signals the service is running.
//   - build_info (observable gauge, value always 1.0) — carries go version and vcs revision as attributes.
//
// Both instruments are registered as asynchronous gauges so they are emitted on every collection cycle
// without additional caller code.
func RegisterServiceMetrics(serviceName string) error {
	meter := otel.Meter(serviceName)

	// Resolve build metadata from the embedded module information.
	goVersion := "unknown"
	vcsRevision := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		goVersion = info.GoVersion
		for _, s := range info.Settings {
			if s.Key == "vcs.revision" {
				vcsRevision = s.Value
			}
		}
	}

	// <service>_up gauge: always 1 while the process is running.
	upGaugeName := serviceName + "_up"
	_, err := meter.Float64ObservableGauge(
		upGaugeName,
		metric.WithDescription("1 if the service is running, 0 when shutting down"),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			o.Observe(1.0, metric.WithAttributes(attribute.String("service", serviceName)))
			return nil
		}),
	)
	if err != nil {
		return err
	}

	// build_info gauge: carries version metadata as attributes.
	_, err = meter.Float64ObservableGauge(
		"build_info",
		metric.WithDescription("Build metadata gauge with go_version and vcs_revision labels"),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			o.Observe(1.0, metric.WithAttributes(
				attribute.String("service", serviceName),
				attribute.String("go_version", goVersion),
				attribute.String("vcs_revision", vcsRevision),
			))
			return nil
		}),
	)
	return err
}

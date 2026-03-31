//go:build integration

package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBFFREDMetricsMiddlewareCoverage validates that otelecho middleware emits
// HTTP server metrics for successful and failing requests.
func TestBFFREDMetricsMiddlewareCoverage(t *testing.T) {
	ctx := context.Background()

	reader := sdkmetric.NewManualReader()
	meterProvider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prevMeterProvider := otel.GetMeterProvider()
	otel.SetMeterProvider(meterProvider)
	t.Cleanup(func() {
		otel.SetMeterProvider(prevMeterProvider)
		_ = meterProvider.Shutdown(context.Background())
	})

	prevTracerProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(noop.NewTracerProvider())
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTracerProvider)
	})

	e := echo.New()
	e.Use(otelecho.Middleware("bff-metrics-test"))
	e.GET("/ok", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/fail", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "boom")
	})

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)

	okResp, err := http.Get(srv.URL + "/ok")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, okResp.StatusCode)
	require.NoError(t, okResp.Body.Close())

	failResp, err := http.Get(srv.URL + "/fail")
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, failResp.StatusCode)
	require.NoError(t, failResp.Body.Close())

	var resourceMetrics metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &resourceMetrics))

	metricFound := false
	requestCountObserved := uint64(0)

	for _, scope := range resourceMetrics.ScopeMetrics {
		for _, metric := range scope.Metrics {
			if metric.Name != "http.server.request.duration" {
				continue
			}
			metricFound = true

			histogram, ok := metric.Data.(metricdata.Histogram[float64])
			require.True(t, ok, "expected histogram metric data for request duration")
			for _, dp := range histogram.DataPoints {
				requestCountObserved += dp.Count
			}
		}
	}

	assert.True(t, metricFound, "expected http.server.request.duration metric to be emitted")
	assert.GreaterOrEqual(t, requestCountObserved, uint64(2), "expected both success and error requests to be recorded")
}

package observe

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/exaring/otelpgx"
	"github.com/alternayte/forge/internal/config"
)

// Setup bootstraps the OpenTelemetry SDK, setting the global TracerProvider
// and MeterProvider. It returns a shutdown function that callers must invoke
// (usually via defer) to flush and stop all exporters gracefully.
//
// When cfg.OTLPEndpoint is non-empty, traces are exported via OTLP HTTP to
// that endpoint. Otherwise traces are written to stdout (pretty-printed) for
// local development visibility.
//
// Metrics are always exported via the Prometheus bridge so that the admin
// server /metrics endpoint can serve them (OTEL-02).
func Setup(ctx context.Context, cfg config.ObserveConfig, serviceName string) (shutdown func(context.Context) error, err error) {
	// --- Trace exporter ---
	var traceExporter sdktrace.SpanExporter
	if cfg.OTLPEndpoint != "" {
		traceExporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(cfg.OTLPEndpoint),
		)
		if err != nil {
			return nil, err
		}
	} else {
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
	}

	// --- Resource ---
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// --- TracerProvider ---
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// --- Prometheus metrics exporter ---
	promExp, err := promexporter.New()
	if err != nil {
		return nil, err
	}

	// --- MeterProvider ---
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// --- Shutdown function ---
	shutdown = func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}
		return mp.Shutdown(ctx)
	}

	return shutdown, nil
}

// NewHTTPHandler wraps the given http.Handler with OpenTelemetry HTTP
// instrumentation, creating spans for each inbound request (OTEL-01).
// The operation parameter becomes the span name prefix (e.g. "api" or "html").
func NewHTTPHandler(handler http.Handler, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}

// PGXTracer returns an otelpgx.Tracer that can be passed to pgxpool.Config as
// a tracer for automatic database query span creation (OTEL-01 DB traces).
func PGXTracer() *otelpgx.Tracer {
	return otelpgx.NewTracer()
}

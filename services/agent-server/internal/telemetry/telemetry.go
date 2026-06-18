package telemetry

import (
	"context"
	"log"

	"google.golang.org/adk/telemetry"
)

// InitADKTelemetry initializes the native ADK hierarchical tracing.
// It relies on standard OpenTelemetry environment variables (e.g., OTEL_EXPORTER_OTLP_ENDPOINT)
func InitADKTelemetry(ctx context.Context) (func(context.Context) error, error) {
	// Initialize ADK telemetry providers. 
	prov, err := telemetry.New(ctx)
	if err != nil {
		return nil, err
	}

	// Sets the global otel.TracerProvider and otel.MeterProvider
	prov.SetGlobalOtelProviders()
	log.Println("ADK Native Telemetry successfully initialized")

	return prov.Shutdown, nil
}

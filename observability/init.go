package observability

import (
	"context"
	"errors"
	"log"
	obmetrics "my-platform/observability/metrics"
	obtracing "my-platform/observability/tracing"

	"go.opentelemetry.io/otel"
	metric2 "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type (
	ObConfig struct {
		Name     string
		Endpoint string
	}
)

func Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}

func Meter(name string) metric2.Meter {
	return otel.GetMeterProvider().Meter(name)
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, cfg ObConfig) (shutdown func(context.Context) error, err error) {

	log.Println("Initializing OpenTelemetry SDK")

	// Verbose error handling
	handleErr := func(inErr error) {
		log.Printf("OpenTelemetry setup error: %v", inErr)
		err = errors.Join(inErr, shutdown(ctx))
	}

	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := obtracing.NewTraceProvider(ctx, cfg.Name, cfg.Endpoint)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := obmetrics.NewMeterProvider(ctx, cfg.Name, cfg.Endpoint)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

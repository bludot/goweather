package tracing

import (
	"context"
	"github.com/bludot/goweather/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
	"log"
)

type Provider struct {
	provider trace.TracerProvider
}

func TracerProvider(config *config.Config) (*Provider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.Tracing.URL)))
	if err != nil {
		log.Println("Failed to create the Jaeger exporter: ", err)
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.ServiceNameKey.String(config.AppConfig.Name),
			semconv.ServiceVersionKey.String(config.AppConfig.Version),
			semconv.DeploymentEnvironmentKey.String("dev"),
		)),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("goweather"),
			attribute.String("environment", "dev"),
			// attribute.Int64("ID", id),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// tracer = struct{ trace.Tracer }{otel.Tracer("goweather")}
	return &Provider{provider: tp}, nil
}

func (p Provider) Close(ctx context.Context) error {
	if prv, ok := p.provider.(*tracesdk.TracerProvider); ok {
		return prv.Shutdown(ctx)
	}

	return nil
}

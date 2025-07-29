package eotel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"google.golang.org/grpc"
)

func InitEOTEL(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	globalCfg = cfg

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("resource.New: %w", err)
	}

	if cfg.EnableTracing {
		tExp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(cfg.OtelCollector),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		)
		if err != nil {
			return nil, fmt.Errorf("trace exporter: %w", err)
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSpanProcessor(sdktrace.NewBatchSpanProcessor(tExp)),
		)
		otel.SetTracerProvider(tp)
	}

	if cfg.EnableMetrics {
		mExp, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithInsecure(),
			otlpmetricgrpc.WithEndpoint(cfg.OtelCollector),
			otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
		)
		if err != nil {
			return nil, fmt.Errorf("metric exporter: %w", err)
		}
		mp := sdkmetric.NewMeterProvider(
			sdkmetric.WithResource(res),
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(mExp)),
		)
		otel.SetMeterProvider(mp)
	}

	if cfg.EnableSentry {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			EnableTracing:    cfg.EnableTracing,
			TracesSampleRate: 1.0,
			Environment:      "production",
		})
		if err != nil {
			log.Printf("init Sentry error: %v", err)
		}
	}

	return func(ctx context.Context) error {
		sentry.Flush(2 * time.Second)
		return nil
	}, nil
}

package eotel

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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

var globalTracer trace.Tracer
var globalMeter metric.Meter

func InitEOTEL(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	globalCfg = cfg

	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceName(cfg.ServiceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("resource.New: %w", err)
	}

	// Init tracing
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
		globalTracer = tp.Tracer(cfg.ServiceName)
	} else {
		globalTracer = otel.GetTracerProvider().Tracer(cfg.ServiceName)
	}

	// Init metrics
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
		globalMeter = mp.Meter(cfg.ServiceName)
	} else {
		globalMeter = otel.GetMeterProvider().Meter(cfg.ServiceName)
	}

	// Init sentry
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

	// Graceful shutdown function
	return func(ctx context.Context) error {
		if cfg.EnableSentry {
			sentry.Flush(2 * time.Second)
		}
		return nil
	}, nil
}

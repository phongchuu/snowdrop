package opentelemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	otelTrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	snowdrop "internal.snowdrop/framework"
)

type SetupOtelSDKParams struct {
	fx.In
	Lifecycle fx.Lifecycle
	Config    snowdrop.ConfigManager
}

type SetupOtelSDKResult struct {
	fx.Out
	Logger         *slog.Logger
	Tracer         otelTrace.Tracer
	MetricProvider *metric.MeterProvider
}

func SetupOtelSDK(p SetupOtelSDKParams) (SetupOtelSDKResult, error) {
	const schemaName = "snowdrop.console"

	// Create resource with service information
	resource, err := createResource(p.Config)
	if err != nil {
		return SetupOtelSDKResult{}, err
	}

	// Set up OpenTelemetry components
	initializeTextMapPropagator()

	loggerProvider, err := newLoggerProvider(p.Lifecycle, resource)
	if err != nil {
		return SetupOtelSDKResult{}, err
	}

	logger := otelslog.NewLogger(
		schemaName,
		otelslog.WithSource(true),
		otelslog.WithLoggerProvider(loggerProvider),
	)

	tracerProvider, err := newTracerProvider(p.Lifecycle, resource)
	if err != nil {
		return SetupOtelSDKResult{}, err
	}

	tracer := tracerProvider.Tracer(schemaName)

	metricProvider, err := newMetricProvider(p.Lifecycle, resource)
	if err != nil {
		return SetupOtelSDKResult{}, err
	}

	// Start runtime metrics collection
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		logger.ErrorContext(
			context.TODO(),
			"otel runtime instrumentation failed:",
			slog.Any("error", err),
		)
	}

	return SetupOtelSDKResult{
		Logger:         logger,
		Tracer:         tracer,
		MetricProvider: metricProvider,
	}, nil
}

func createResource(config snowdrop.ConfigManager) (*resource.Resource, error) {
	return resource.New(
		context.Background(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceName("snowdrop.console"),
			semconv.ServiceVersion(config.GetAppVersion()),
			attribute.String("service.revision", config.GetAppRevision()),
		),
	)
}

func initializeTextMapPropagator() {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)
}

func newTracerProvider(
	lifecycle fx.Lifecycle,
	r *resource.Resource,
) (*trace.TracerProvider, error) {
	traceExporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(otlptracegrpc.WithInsecure()),
	)
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(r),
		trace.WithSampler(trace.AlwaysSample()),
	)
	lifecycle.Append(fx.StopHook(tracerProvider.Shutdown))
	otel.SetTracerProvider(tracerProvider)

	return tracerProvider, nil
}

func newMetricProvider(
	lifecycle fx.Lifecycle,
	r *resource.Resource,
) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(context.TODO(), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(r),
	)
	lifecycle.Append(fx.StopHook(meterProvider.Shutdown))
	otel.SetMeterProvider(meterProvider)

	return meterProvider, nil
}

func newLoggerProvider(lifecycle fx.Lifecycle, r *resource.Resource) (*log.LoggerProvider, error) {
	stdoutExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	logExporter, err := otlploggrpc.New(context.Background(), otlploggrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	stdoutProcesor := log.NewSimpleProcessor(
		stdoutExporter,
	)

	httpProcessor := log.NewBatchProcessor(
		logExporter,
	)

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(httpProcessor),
		log.WithProcessor(stdoutProcesor),
		log.WithResource(r),
	)
	lifecycle.Append(fx.StopHook(loggerProvider.Shutdown))
	global.SetLoggerProvider(loggerProvider)

	return loggerProvider, nil
}

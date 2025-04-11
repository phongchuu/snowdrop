package log

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
	Tracer otelTrace.Tracer
	Logger *slog.Logger
}

func SetupOtelSDK(p SetupOtelSDKParams) (SetupOtelSDKResult, error) {
	// Initialize basic components
	schemaName := "snowdrop.console"
	tracer := otel.Tracer(schemaName)
	logger := otelslog.NewLogger(schemaName, otelslog.WithSource(true))

	// Create resource with service information
	resource, err := createResource(p.Config)
	if err != nil {
		return SetupOtelSDKResult{}, err
	}

	// Set up OpenTelemetry components
	setupPropagator()

	if err := setupTracing(p.Lifecycle, resource); err != nil {
		return SetupOtelSDKResult{}, err
	}

	if err := setupMetrics(p.Lifecycle, resource); err != nil {
		return SetupOtelSDKResult{}, err
	}

	if err := setupLogging(p.Lifecycle, resource); err != nil {
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
		Tracer: tracer,
		Logger: logger,
	}, nil
}

func createResource(config snowdrop.ConfigManager) (*resource.Resource, error) {
	return resource.New(
		context.TODO(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceName("snowdrop.console"),
			semconv.ServiceVersion(config.GetAppVersion()),
			attribute.String("service.revision", config.GetAppRevision()),
		),
	)
}

func setupPropagator() {
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)
}

func setupTracing(lifecycle fx.Lifecycle, r *resource.Resource) error {
	traceExporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(otlptracegrpc.WithInsecure()),
	)
	if err != nil {
		return err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(r),
		trace.WithSampler(trace.AlwaysSample()),
	)
	lifecycle.Append(fx.StopHook(tracerProvider.Shutdown))
	otel.SetTracerProvider(tracerProvider)

	return nil
}

func setupMetrics(lifecycle fx.Lifecycle, r *resource.Resource) error {
	metricExporter, err := otlpmetricgrpc.New(context.TODO(), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(r),
	)
	lifecycle.Append(fx.StopHook(meterProvider.Shutdown))
	otel.SetMeterProvider(meterProvider)

	return nil
}

func setupLogging(lifecycle fx.Lifecycle, r *resource.Resource) error {
	stdoutExporter, err := stdoutlog.New(stdoutlog.WithPrettyPrint())
	if err != nil {
		return err
	}

	logExporter, err := otlploggrpc.New(context.Background(), otlploggrpc.WithInsecure())
	if err != nil {
		return err
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

	return nil
}

package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Name        string  `mapstructure:"name" json:"name" yaml:"name"`
	SampleRatio float64 `mapstructure:"sample_ratio" json:"sample_ratio" yaml:"sample_ratio"`
	Endpoint    string  `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	Insecure    bool    `mapstructure:"insecure" json:"insecure" yaml:"insecure"`
}

type Tracer interface {
	StartSpan(ctx context.Context, name string, args ...any) context.Context
	EndSpan(ctx context.Context)
}

type otelTracer struct {
	tracer trace.Tracer
}

type noopTracer struct{}

func (noopTracer) StartSpan(ctx context.Context, name string, args ...any) context.Context { return ctx }
func (noopTracer) EndSpan(ctx context.Context)                                            {}

var tracer Tracer = noopTracer{}
var provider *sdktrace.TracerProvider

func Init(cfg *Config) error {
	if cfg == nil || cfg.Endpoint == "" {
		tracer = noopTracer{}
		return nil
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	} else {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	client := otlptracegrpc.NewClient(opts...)
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return err
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.Name),
		),
		resource.WithHost(),
	)
	if err != nil {
		return err
	}

	provider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	tracer = otelTracer{tracer: provider.Tracer(cfg.Name)}
	return nil
}

func Shutdown(ctx context.Context) error {
	if provider == nil {
		return nil
	}
	return provider.Shutdown(ctx)
}

func StartSpan(ctx context.Context, name string, args ...any) context.Context {
	return tracer.StartSpan(ctx, name, args...)
}

func EndSpan(ctx context.Context) {
	tracer.EndSpan(ctx)
}

func Get() Tracer {
	return tracer
}

func WithLoggerErrorHandler(logger interface{}) {}

type queryTracer struct{}

func NewQueryTracer(t Tracer) any {
	return queryTracer{}
}

func (o otelTracer) StartSpan(ctx context.Context, name string, args ...any) context.Context {
	ctx, _ = o.tracer.Start(ctx, name)
	return ctx
}

func (otelTracer) EndSpan(ctx context.Context) {}

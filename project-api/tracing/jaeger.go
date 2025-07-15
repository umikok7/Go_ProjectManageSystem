package tracing

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"test.com/project-api/config"
)

func JaegerTraceProvider() (*sdktrace.TracerProvider, error) {
	// 获取配置
	jaegerConfig := config.C.JaegerConfig
	if !jaegerConfig.Enabled {
		// 如果禁用，返回一个空的 TracerProvider
		return sdktrace.NewTracerProvider(), nil
	}
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerConfig.Endpoint)))
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(jaegerConfig.ServiceName),
			semconv.DeploymentEnvironmentKey.String(jaegerConfig.Environment),
		)),
	)
	return tp, nil
}

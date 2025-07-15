package tracing

import (
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"test.com/project-user/config"
)

func JaegerTraceProvider() (*sdktrace.TracerProvider, error) {
	// 获取配置
	jaegerConfig := config.C.JaegerConfig
	if !jaegerConfig.Enabled {
		// 如果禁用，返回一个空的 TracerProvider
		return sdktrace.NewTracerProvider(), nil
	}
	// 创建 Jaeger Exporter（负责将追踪数据发送到 Jaeger）
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerConfig.Endpoint)))
	if err != nil {
		return nil, err
	}
	// 创建 TracerProvider（追踪数据的生产者）
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(jaegerConfig.ServiceName),           // 设置服务名
			semconv.DeploymentEnvironmentKey.String(jaegerConfig.Environment), // 设置环境
		)),
	)
	return tp, nil
}

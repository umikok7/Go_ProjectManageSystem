package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	srv "test.com/project-common"
	"test.com/project-user/config"
	"test.com/project-user/router"
	"test.com/project-user/tracing"
)

func main() {
	r := gin.Default()

	// jaeger相关
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	// 设置为全局 TracerProvider
	otel.SetTracerProvider(tp)
	// 设置追踪上下文传播器（用于服务间传递追踪信息）
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	r.Use(otelgin.Middleware("project-user")) // jaeger相关
	//从配置中读取日志配置，初始化日志
	config.C.InitZapLog()
	// 路由
	router.InitRouter(r)
	// 注册grpc服务
	gc := router.RegisterGrpc()
	// grpc服务注册到etcd当中
	router.RegisterEtcdServer()
	stop := func() {
		gc.Stop()
	}
	srv.Run(r, config.C.SC.Name, config.C.SC.Port, stop)
}

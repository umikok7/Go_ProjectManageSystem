package main

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	srv "test.com/project-common"
	"test.com/project-project/config"
	"test.com/project-project/router"
	"test.com/project-project/tracing"
)

func main() {
	r := gin.Default()
	// 添加这一句后bug得到修复
	config.C.InitZapLog()

	// jaeger相关
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	r.Use(otelgin.Middleware("project-project")) // jaeger相关

	// 路由
	router.InitRouter(r)
	// 初始化rpc调用
	router.InitUserRpc()
	// 注册grpc服务
	gc := router.RegisterGrpc()
	// grpc服务注册到etcd当中
	router.RegisterEtcdServer()
	// 初始化Kafka
	c := config.InitKafkaWriter()
	stop := func() {
		gc.Stop()
		c()
	}

	srv.Run(r, config.C.SC.Name, config.C.SC.Port, stop)
}

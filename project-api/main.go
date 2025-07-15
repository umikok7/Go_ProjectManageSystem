package main

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"log"
	"net/http"
	_ "test.com/project-api/api"
	"test.com/project-api/config"
	"test.com/project-api/router"
	"test.com/project-api/tracing"
	srv "test.com/project-common"
)

func main() {
	r := gin.Default()
	// jaeger相关
	tp, tpErr := tracing.JaegerTraceProvider()
	if tpErr != nil {
		log.Fatal(tpErr)
	}
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	//r.Use(midd.RequestLog())
	r.Use(otelgin.Middleware("project-api")) // jaeger相关
	r.StaticFS("/upload", http.Dir("upload"))
	//从配置中读取日志配置，初始化日志
	config.C.InitZapLog()
	// 路由
	router.InitRouter(r)
	// 开启pprof 默认的访问路径是/debug/pprof
	pprof.Register(r)
	srv.Run(r, config.C.SC.Name, config.C.SC.Port, nil)
}

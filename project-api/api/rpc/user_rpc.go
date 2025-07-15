package rpc

import (
	"log"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"test.com/project-api/config"
	"test.com/project-common/discovery"
	"test.com/project-common/logs"
	login "test.com/project-grpc/user/login"
)

var LoginServiceClient login.LoginServiceClient

// InitRpcUserClient etcd的客户端，etcd服务的消费者
func InitRpcUserClient() {
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG) // 创建etcd实例
	resolver.Register(etcdRegister)                                           // 注册到 gRPC 的全局解析器注册表

	conn, err := grpc.Dial("etcd:///user",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 创建不安全的凭证，通过Dial与gRPC服务器建立连接
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),       // 集成 OpenTelemetry 追踪功能
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	LoginServiceClient = login.NewLoginServiceClient(conn)

}

package rpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"test.com/project-common/discovery"
	"test.com/project-common/logs"
	login "test.com/project-grpc/user/login"
	"test.com/project-project/config"
)

var LoginServiceClient login.LoginServiceClient

// InitRpcUserClient etcd的客户端，etcd服务的消费者
func InitRpcUserClient() {
	// 添加这一段 防止忘记预先初始化log
	if logs.LG == nil {
		log.Println("WARNING: logs.LG is nil, initializing logger...")
		config.C.InitZapLog()
		if logs.LG == nil {
			log.Fatal("FATAL: Failed to initialize logger")
		}
	}
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG) // 创建etcd实例
	resolver.Register(etcdRegister)                                           // 注册到 gRPC 的全局解析器注册表

	conn, err := grpc.Dial(
		"etcd:///user",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 创建不安全的凭证，通过Dial与gRPC服务器建立连接
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),       // 集成 OpenTelemetry 追踪功能

	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	LoginServiceClient = login.NewLoginServiceClient(conn)

}

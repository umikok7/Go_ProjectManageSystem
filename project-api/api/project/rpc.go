package project

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"test.com/project-api/config"
	"test.com/project-common/discovery"
	"test.com/project-common/logs"
	"test.com/project-grpc/account"
	"test.com/project-grpc/auth"
	"test.com/project-grpc/department"
	"test.com/project-grpc/menu"
	"test.com/project-grpc/project"
	"test.com/project-grpc/task"
)

var ProjectServiceClient project.ProjectServiceClient
var TaskServiceClient task.TaskServiceClient
var AccountServiceClient account.AccountServiceClient
var DepartmentServiceClient department.DepartmentServiceClient
var AuthServiceClient auth.AuthServiceClient
var MenuServiceClient menu.MenuServiceClient

// InitRpcProjectClient etcd的客户端，etcd服务的消费者
func InitRpcProjectClient() {
	etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG) // 创建etcd实例
	resolver.Register(etcdRegister)                                           // 注册到 gRPC 的全局解析器注册表

	conn, err := grpc.Dial("etcd:///project",
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 创建不安全的凭证，通过Dial与gRPC服务器建立连接
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),       // 集成 OpenTelemetry 追踪功能
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	ProjectServiceClient = project.NewProjectServiceClient(conn)
	TaskServiceClient = task.NewTaskServiceClient(conn)
	AccountServiceClient = account.NewAccountServiceClient(conn)
	DepartmentServiceClient = department.NewDepartmentServiceClient(conn)
	AuthServiceClient = auth.NewAuthServiceClient(conn)
	MenuServiceClient = menu.NewMenuServiceClient(conn)
}

package router

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"test.com/project-common/discovery"
	"test.com/project-common/logs"
	"test.com/project-grpc/account"
	"test.com/project-grpc/auth"
	"test.com/project-grpc/department"
	"test.com/project-grpc/menu"
	"test.com/project-grpc/project"
	"test.com/project-grpc/task"
	"test.com/project-project/config"
	"test.com/project-project/internal/rpc"
	account_service_v1 "test.com/project-project/pkg/service/account.service.v1"
	auth_service_v1 "test.com/project-project/pkg/service/auth.service.v1"
	department_service_v1 "test.com/project-project/pkg/service/department.service.v1"
	menu_service_v1 "test.com/project-project/pkg/service/menu.service.v1"
	project_service_v1 "test.com/project-project/pkg/service/project.service.v1"
	task_service_v1 "test.com/project-project/pkg/service/task.service.v1"
)

// Router 接口
type Router interface {
	Router(r *gin.Engine)
}

type RegisterRouter struct {
}

func New() *RegisterRouter {
	return &RegisterRouter{}
}

func (*RegisterRouter) Router(ro Router, r *gin.Engine) {
	ro.Router(r)
}

var routers []Router

func InitRouter(r *gin.Engine) {
	//rg := New()
	//rg.Router(&user.RouterUser{}, r)

	// 通过这种方式添加路由很巧妙
	for _, ro := range routers {
		ro.Router(r)
	}
}

func Register(ro ...Router) {
	routers = append(routers, ro...)
}

type gRPCConfig struct {
	Addr         string
	RegisterFunc func(*grpc.Server)
}

func RegisterGrpc() *grpc.Server {
	c := gRPCConfig{
		Addr: config.C.GC.Addr,
		RegisterFunc: func(g *grpc.Server) {
			project.RegisterProjectServiceServer(g, project_service_v1.New())          // 注册LoginService服务
			task.RegisterTaskServiceServer(g, task_service_v1.New())                   // 注册TaskService服务
			account.RegisterAccountServiceServer(g, account_service_v1.New())          // 注册AccountService服务
			department.RegisterDepartmentServiceServer(g, department_service_v1.New()) // 注册DepartmentService服务
			auth.RegisterAuthServiceServer(g, auth_service_v1.New())                   // 注册AuthService服务
			menu.RegisterMenuServiceServer(g, menu_service_v1.New())                   // 注册MenuService服务
		}}
	// 创建拦截器
	//s := grpc.NewServer(interceptor.New().Cache())
	// 使用新的 StatsHandler API
	s := grpc.NewServer(
		grpc.StatsHandler(
			otelgrpc.NewServerHandler(),
		), // 追踪
		//grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		//	interceptor.New().CacheInterceptor(), // 业务逻辑拦截器处理缓存
		//)),
	)
	c.RegisterFunc(s)
	lis, err := net.Listen("tcp", c.Addr) // 创建监听
	if err != nil {
		log.Println("cannot listen")
	}
	go func() {
		err := s.Serve(lis) // 启动gRPC服务
		if err != nil {
			log.Println("serve started error", err)
			return
		}
	}()
	return s
}

// RegisterEtcdServer etcd的服务端
func RegisterEtcdServer() {
	//etcdRegister := discovery.NewResolver(config.C.EtcdConfig.Addrs, logs.LG)
	//resolver.Register(etcdRegister)
	// 从配置文件中获取服务注册所需的所有信息
	info := discovery.Server{
		Name:    config.C.GC.Name,
		Addr:    config.C.GC.Addr,
		Version: config.C.GC.Version,
		Weight:  config.C.GC.Weight,
	}
	r := discovery.NewRegister(config.C.EtcdConfig.Addrs, logs.LG)
	_, err := r.Register(info, 2) // 将自己注册到etcd当中去
	if err != nil {
		log.Fatalln(err)
	}
}

func InitUserRpc() {
	rpc.InitRpcUserClient()
}

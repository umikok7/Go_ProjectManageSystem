package router

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"log"
	"net"
	"test.com/project-common/discovery"
	"test.com/project-common/logs"
	"test.com/project-grpc/user/login"
	"test.com/project-user/config"
	loginServiceV1 "test.com/project-user/pkg/service/login.service.v1"
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
			login.RegisterLoginServiceServer(g, loginServiceV1.New()) // 注册LoginService服务
		}}
	// 创建拦截器
	s := grpc.NewServer(
		// 使用 StatsHandler 进行 gRPC 追踪
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

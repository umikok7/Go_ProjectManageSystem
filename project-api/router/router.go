package router

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
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

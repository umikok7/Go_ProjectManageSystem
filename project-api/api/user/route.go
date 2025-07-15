package user

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/midd"
	"test.com/project-api/api/rpc"
	"test.com/project-api/router"
)

func init() {
	log.Println("init user router")
	router.Register(&RouterUser{})
}

type RouterUser struct {
}

func (*RouterUser) Router(r *gin.Engine) {
	// 初始化gRPC的客户端连接
	rpc.InitRpcUserClient()
	h := New()
	r.POST("project/login/getCaptcha", h.getCaptcha)
	r.POST("project/login/register", h.register)
	r.POST("project/login", h.login)
	org := r.Group("/project/organization")
	org.Use(midd.TokenVerify())
	org.POST("/_getOrgList", h.myOrgList)
}

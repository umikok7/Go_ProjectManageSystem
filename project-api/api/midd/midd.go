package midd

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"test.com/project-api/api/rpc"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/user/login"
	"time"
)

// GetIp 获取ip函数
func GetIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}

func TokenVerify() func(ctx *gin.Context) {
	result := &common.Result{}
	return func(c *gin.Context) {
		// 1. 从handler中获取token
		token := c.GetHeader("Authorization")
		// 2. 调用user服务进行token认证
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		ip := GetIp(c)
		// 先去查询node表 确认不使用登录控制的接口，不做登录认证了
		response, err := rpc.LoginServiceClient.TokenVerify(ctx, &login.LoginMessage{Token: token, Ip: ip}) // 调用对应gRPC方法
		// 3. 处理结果，认证通过将信息放入gin的上下文，失败返回未登陆
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort()
			return
		}
		c.Set("memberId", response.Member.Id)
		c.Set("memberName", response.Member.Name)
		c.Set("organizationCode", response.Member.OrganizationCode)
		c.Next()
	}
}

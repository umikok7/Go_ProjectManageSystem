package user

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"test.com/project-api/api/rpc"
	"test.com/project-api/pkg/model/user"
	common "test.com/project-common"
	"test.com/project-common/errs"
	login "test.com/project-grpc/user/login"
	"time"
)

type HandlerUser struct {
}

func New() *HandlerUser {
	return &HandlerUser{}
}

func (h *HandlerUser) getCaptcha(ctx *gin.Context) {
	rsp := &common.Result{}
	// 1. 获取参数
	mobile := ctx.PostForm("mobile")
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	capchaRsp, err := rpc.LoginServiceClient.GetCaptcha(c, &login.CaptchaMessage{Mobile: mobile})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, rsp.Fail(code, msg))
		return
	}
	ctx.JSON(http.StatusOK, rsp.Success(capchaRsp.Code))
}

func (h *HandlerUser) register(ctx *gin.Context) {
	// 1. 接收参数 参数模型
	rsp := &common.Result{}
	var req user.RegisterReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, rsp.Fail(http.StatusBadRequest, "参数格式错误"))
		return
	}
	// 2. 校验参数 判断参数是否合法
	if err := req.Verify(); err != nil {
		ctx.JSON(http.StatusOK, rsp.Fail(http.StatusBadRequest, err.Error()))
		return
	}
	// 3. 调用user gRPC服务 获取响应
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &login.RegisterMessage{}
	err = copier.Copy(msg, req)
	if err != nil {
		ctx.JSON(http.StatusOK, rsp.Fail(http.StatusBadRequest, "copy错误"))
		return
	}
	_, err = rpc.LoginServiceClient.Register(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, rsp.Fail(code, msg))
		return
	}
	// 4. 返回结果
	ctx.JSON(http.StatusOK, rsp.Success(""))
}

// GetIp 获取ip函数
func GetIp(c *gin.Context) string {
	ip := c.ClientIP()
	if ip == "::1" {
		ip = "127.0.0.1"
	}
	return ip
}

func (h *HandlerUser) login(ctx *gin.Context) {
	// 1. 接收参数 参数模型
	rsp := &common.Result{}
	var req user.LoginReq
	err := ctx.ShouldBind(&req)
	if err != nil {
		ctx.JSON(http.StatusOK, rsp.Fail(http.StatusBadRequest, "参数格式错误"))
		return
	}
	// 2. 调用user grpc完成登陆
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &login.LoginMessage{}
	err = copier.Copy(msg, req) // 将req复制到msg当中
	if err != nil {
		ctx.JSON(http.StatusOK, rsp.Fail(http.StatusBadRequest, "copy错误"))
		return
	}
	msg.Ip = GetIp(ctx)
	loginRsp, err := rpc.LoginServiceClient.Login(c, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		ctx.JSON(http.StatusOK, rsp.Fail(code, msg))
		return
	}
	response := &user.LoginRsp{}
	err = copier.Copy(response, loginRsp)
	if err != nil {
		ctx.JSON(http.StatusOK, rsp.Fail(http.StatusBadRequest, "copy错误"))
		return
	}
	// 4. 返回结果
	ctx.JSON(http.StatusOK, rsp.Success(response))
}

func (h *HandlerUser) myOrgList(c *gin.Context) {
	result := &common.Result{}
	memberIdStr, _ := c.Get("memberId")
	memberId := memberIdStr.(int64)
	list, err2 := rpc.LoginServiceClient.MyOrgList(context.Background(), &login.UserMessage{MemId: memberId})
	if err2 != nil {
		code, msg := errs.ParseGrpcError(err2)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	if list.OrganizationList == nil {
		c.JSON(http.StatusOK, result.Success([]*user.OrganizationList{}))
		return
	}
	var orgs []*user.OrganizationList
	err2 = copier.Copy(&orgs, list.OrganizationList)
	c.JSON(http.StatusOK, result.Success(orgs))
}

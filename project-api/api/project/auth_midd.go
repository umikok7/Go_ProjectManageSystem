package project

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	common "test.com/project-common"
	"test.com/project-common/errs"
)

var ignores = []string{
	"project/login/register",
	"project/login",
	"project/login/getCaptcha",
	"project/organization",
	"project/auth/apply"}

func Auth() func(ctx *gin.Context) {
	return func(c *gin.Context) {
		result := common.Result{}
		uri := c.Request.RequestURI

		// 判断url是否包含在ignores里面
		for _, v := range ignores {
			if strings.Contains(uri, v) {
				c.Next() // 白名单直接放行
				return
			}
		}
		// 判断此uri是否在用户的授权列表中
		a := NewAuth()
		nodes, err := a.GetAuthNode(c)
		if err != nil {
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort()
			return
		}
		// 再次判断，检查是否有权限进行访问
		for _, v := range nodes {
			if strings.Contains(uri, v) {
				c.Next() // 有权限，放行
				return
			}
		}
		// 无权限
		c.JSON(http.StatusOK, result.Fail(403, "无权限操作"))
		c.Abort()
		return
	}
}

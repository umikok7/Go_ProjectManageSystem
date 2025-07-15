package project

import (
	"net/http"

	"github.com/gin-gonic/gin"
	common "test.com/project-common"
	"test.com/project-common/errs"
)

// ProjectAuth 该中间件用于确保只有项目成员才能操作项目相关功能，而对于私有项目，只有项目所有者才能进行操作
func ProjectAuth() func(ctx *gin.Context) {
	return func(c *gin.Context) {
		// 如果此用户不是项目成员，则不能操作此项目，直接返回无权限
		result := &common.Result{}
		// 如果不是该项目的成员，无权限查看项目和操作项目
		// 检查是否有projectCode和taskCode这两个参数,有的话则说明需要项目认证了
		isProjectAuth := false
		projectCode := c.PostForm("projectCode")
		if projectCode != "" {
			isProjectAuth = true
		}
		taskCode := c.PostForm("taskCode")
		if taskCode != "" {
			isProjectAuth = true
		}
		if isProjectAuth {
			p := New()
			pr, isMember, isOwner, err := p.FindProjectByMemberId(c.GetInt64("memberId"), projectCode, taskCode)
			if err != nil {
				code, msg := errs.ParseGrpcError(err)
				c.JSON(http.StatusOK, result.Fail(code, msg))
				c.Abort() // 停止执行下一个中间件或者handler函数
				return
			}
			if !isMember {
				c.JSON(http.StatusOK, result.Fail(403, "不是项目成员，无操作权限"))
				c.Abort()
				return
			}
			if pr.Private == 1 {
				//私有项目
				if isOwner {
					c.Next()
					return
				} else {
					c.JSON(http.StatusOK, result.Fail(403, "私有项目，无操作权限"))
					c.Abort()
					return
				}
			}
			// 公开项目并且是成员
			c.Next()
		} else {
			// 不需要进行验证的时候，直接放行
			c.Next()
		}
	}
}

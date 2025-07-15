package model

import "github.com/gin-gonic/gin"

type Page struct {
	Page     int64 `json:"page" form:"page"`
	PageSize int64 `json:"pageSize" form:"pageSize"`
}

func (p *Page) Bind(c *gin.Context) {
	_ = c.ShouldBind(&p) // 将 HTTP 请求中的参数自动绑定到 Page 结构体
	if p.Page == 0 {
		p.Page = 1
	}
	if p.PageSize == 0 {
		p.PageSize = 10
	}
}

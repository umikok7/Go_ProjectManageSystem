package model

import (
	"test.com/project-common/errs"
)

var (
	RedisError         = errs.NewError(999, "redis错误")
	DBError            = errs.NewError(998, "db错误")
	NoLogin            = errs.NewError(997, "未登陆")
	NoLegalMobile      = errs.NewError(10102001, "手机号不合法")
	CaptchaError       = errs.NewError(10102002, "验证码不正确")
	CaptchaNotExist    = errs.NewError(10102003, "验证码不存在")
	EmailExist         = errs.NewError(10102004, "邮箱已经存在")
	AccountExist       = errs.NewError(10102005, "账号已经存在")
	MobileExist        = errs.NewError(10102006, "手机号已经存在")
	AccountAndPwdError = errs.NewError(10102006, "账号密码不正确")
)

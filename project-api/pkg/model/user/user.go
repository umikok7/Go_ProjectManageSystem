package user

import (
	"errors"
	common "test.com/project-common"
)

type RegisterReq struct {
	Email     string `json:"email" form:"email"`
	Name      string `json:"name" form:"name"`
	Password  string `json:"password" form:"password"`
	Password2 string `json:"password2" form:"password2"`
	Mobile    string `json:"mobile" form:"mobile"`
	Captcha   string `json:"captcha" form:"captcha"`
}

func (r *RegisterReq) VerifyPassword() bool {
	return r.Password == r.Password2
}

// Verify 验证参数是否合法
func (r *RegisterReq) Verify() error {
	if !common.VerifyMobile(r.Mobile) {
		return errors.New("手机号格式不正确")
	}
	if !r.VerifyPassword() {
		return errors.New("两次输入的密码不一致")
	}
	if !common.VerifyEmailFormat(r.Email) {
		return errors.New("邮箱格式不正确")
	}
	return nil
}

type LoginReq struct {
	Account  string `json:"account" form:"account"`
	Password string `json:"password" form:"password"`
}

type LoginRsp struct {
	Member           Member             `json:"member"`
	TokenList        TokenList          `json:"tokenList"`
	OrganizationList []OrganizationList `json:"organizationList"`
}
type Member struct {
	Id               int64  `json:"id"`
	Name             string `json:"name"`
	Mobile           string `json:"mobile"`
	Status           int    `json:"status"`
	Code             string `json:"code"`
	CreateTime       string `json:"create_Time"`
	LastLoginTime    string `json:"last_login_time"`
	OrganizationCode string `json:"organization_code"`
}

type TokenList struct {
	AccessToken    string `json:"accessToken"`
	RefreshToken   string `json:"refreshToken"`
	TokenType      string `json:"tokenType"`
	AccessTokenExp int64  `json:"accessTokenExp"`
}

type OrganizationList struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Description string `json:"description"`
	OwnerCode   string `json:"owner_code"`
	MemberId    int64  `json:"memberId"`
	CreateTime  int64  `json:"create_Time"`
	Personal    int32  `json:"personal"`
	Address     string `json:"address"`
	Province    int32  `json:"province"`
	City        int32  `json:"city"`
	Area        int32  `json:"area"`
	Code        string `json:"code"`
}

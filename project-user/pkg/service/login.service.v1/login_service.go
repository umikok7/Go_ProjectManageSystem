package login_service_v1

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"log"
	"strconv"
	"strings"
	common "test.com/project-common"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-common/jwts"
	"test.com/project-common/tms"
	"test.com/project-grpc/user/login"
	"test.com/project-user/config"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/data/organization"
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/tran"
	"test.com/project-user/internal/repo"
	"test.com/project-user/pkg/model"
	"time"
)

type LoginService struct {
	login.UnimplementedLoginServiceServer
	cache            repo.Cache
	memberRepo       repo.MemberRepo
	organizationRepo repo.OrganizationRepo
	transaction      tran.Transaction
}

func New() *LoginService {
	return &LoginService{
		cache:            dao.Rc,
		memberRepo:       dao.NewMemberDao(),
		organizationRepo: dao.NewOrganizationDao(),
		transaction:      dao.NewTransaction(),
	}
}

// GetCaptcha 处理RPC请求
func (ls *LoginService) GetCaptcha(ctx context.Context, msg *login.CaptchaMessage) (*login.CaptchaResponse, error) {
	// 1. 获取参数
	mobile := msg.Mobile
	// 2. 校验参数
	if !common.VerifyMobile(mobile) {
		return nil, errs.GrpcError(model.NoLegalMobile)
	}
	// 3. 生成验证码 （随机四位1000-9999，或者六位100000-999999）
	code := "123456"
	// 4. 调用短信平台（第三方效率不可保证，所以放入go程中执行，接口可以快速响应）
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功，发送短信 info")
		// redis 假设后续缓存可能存在mysql当中，也可能存在mongo当中，也可能存在memcache当中
		// 5. 存储验证码到 redis 当中 过期时间15min
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second) // 创建一个带有2s超时时间的上下文
		defer cancel()                                                        // 确保函数结束时释放上下文资源避免资源泄漏
		err := ls.cache.Put(c, "REGISTER_"+mobile, code, 15*time.Minute)
		if err != nil {
			log.Printf("验证码存入reids出错，cause by： %v\n", err)
			//zap.L().Info("验证码存入redis出错，cause by： %v\n", err)
		}
	}()

	return &login.CaptchaResponse{Code: code}, nil
}

func (ls *LoginService) Register(ctx context.Context, msg *login.RegisterMessage) (*login.RegisterResponse, error) {
	c := context.Background()
	// 1. 获取参数，校验参数
	// 2. 校验验证码
	redisCode, err := ls.cache.Get(c, model.RegisterRedisKey+msg.Mobile)
	if err == redis.Nil {
		return nil, errs.GrpcError(model.CaptchaNotExist)
	}
	if err != nil {
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.RedisError)
	}
	if redisCode != msg.Captcha {
		return nil, errs.GrpcError(model.CaptchaError)
	}
	// 3. 校验业务逻辑（此处例如 邮箱、账号或者手机号等是否被注册过）
	exist, err := ls.memberRepo.GetMemberByEmail(c, msg.Email)
	if err != nil {
		zap.L().Error("Register db get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.EmailExist)
	}

	exist, err = ls.memberRepo.GetMemberByAccount(c, msg.Name)
	if err != nil {
		zap.L().Error("Register db get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.AccountExist)
	}

	exist, err = ls.memberRepo.GetMemberByMobile(c, msg.Mobile)
	if err != nil {
		zap.L().Error("Register db get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if exist {
		return nil, errs.GrpcError(model.MobileExist)
	}

	// 4. 执行业务 将数据存入member表，生成对应数据，存入组织表
	pwd := encrypts.Md5(msg.Password)
	mem := &member.Member{
		Account:       msg.Name,
		Password:      pwd,
		Name:          msg.Name,
		Mobile:        msg.Mobile,
		Email:         msg.Mobile,
		CreateTime:    time.Now().UnixMilli(),
		LastLoginTime: time.Now().UnixMilli(),
		Status:        model.Normal,
	}
	err = ls.transaction.Action(func(conn database.DbConn) error {
		err = ls.memberRepo.SaveMember(conn, ctx, mem)
		if err != nil {
			zap.L().Error("SaveMember get error", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		org := &organization.Organization{
			Name:       mem.Name + "个人组织",
			MemberId:   mem.Id,
			CreateTime: time.Now().UnixMilli(),
			Personal:   model.Personal,
			Avatar:     "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fc-ssl.dtstatic.com%2Fuploads%2Fblog%2F202103%2F31%2F20210331160001_9a852.thumb.1000_0.jpg&refer=http%3A%2F%2Fc-ssl.dtstatic.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1673017724&t=ced22fc74624e6940fd6a89a21d30cc5",
		}
		err = ls.organizationRepo.SaveOrganization(conn, ctx, org)
		if err != nil {
			zap.L().Error("register SaveOrganization db err", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})
	// ToDoList 生成一个账户 账户的授权角色是成员，新生成一个角色（如果没有），同时将此角色授权的node生成

	// 5. 返回响应
	return &login.RegisterResponse{}, err
}

func (ls *LoginService) Login(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	c := context.Background()
	// 1. 去数据库查询账号密码是否正确
	pwd := encrypts.Md5(msg.Password)                         // 通过Md5加密保证用户隐私
	mem, err := ls.memberRepo.FindMember(c, msg.Account, pwd) // 获取对应的用户
	if err != nil {
		zap.L().Error("FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if mem == nil {
		return nil, errs.GrpcError(model.AccountAndPwdError)
	}
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, mem)
	// 对数据库Id进行加密，采用AES进行可逆的加密，防遍历攻击
	memMsg.Code, _ = encrypts.EncryptInt64(mem.Id, model.AESKey)
	memMsg.LastLoginTime = tms.FormatByMill(mem.LastLoginTime)
	memMsg.CreateTime = tms.FormatByMill(mem.CreateTime)

	// 2. 根据用户id查询组织
	orgs, err := ls.organizationRepo.FindOrganizationByMemId(c, mem.Id) // 获取到对应的组织
	if err != nil {
		zap.L().Error("FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgsMessage []*login.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	for _, v := range orgsMessage {
		v.Code, _ = encrypts.EncryptInt64(v.Id, model.AESKey)
		v.OwnerCode = memMsg.Code
		o := organization.ToMap(orgs)[v.Id]
		v.CreateTime = tms.FormatByMill(o.CreateTime)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}

	// 3. 登陆成功用jwt生成token
	memIdStr := strconv.FormatInt(mem.Id, 10)
	exp := time.Duration(config.C.JwtConfig.AccessExp*3600*24) * time.Second
	rExp := time.Duration(config.C.JwtConfig.RefreshExp*3600*24) * time.Second
	token := jwts.CreateToken(memIdStr, exp, config.C.JwtConfig.AccessSecret,
		rExp, config.C.JwtConfig.RefreshSecret, msg.Ip)
	// 可以给token进行加密处理增加安全性
	tokenList := &login.TokenMessage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		TokenType:      "bearer",
		AccessTokenExp: token.AccessExp,
	}

	// 优化点 ：放入缓存中
	go func() {
		marshal, _ := json.Marshal(mem)
		ls.cache.Put(c, model.Member+"::"+memIdStr, string(marshal), exp)
		orgsJson, _ := json.Marshal(orgs)
		ls.cache.Put(c, model.MemberOrganization+"::"+memIdStr, string(orgsJson), exp)
	}()

	// 4. 返回响应
	return &login.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgsMessage,
		TokenList:        tokenList,
	}, nil
}

func (ls *LoginService) TokenVerify(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	token := msg.Token
	if strings.Contains(token, "bearer") {
		token = strings.ReplaceAll(token, "bearer ", "")
	}
	parseToken, err := jwts.ParseToken(token, config.C.JwtConfig.AccessSecret, msg.Ip)
	if err != nil {
		zap.L().Error("Login TokenVerify error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}

	// 从缓存中查询，如果没有，直接返回认证错误
	memJson, err := ls.cache.Get(context.Background(), model.Member+"::"+parseToken)
	if err != nil {
		zap.L().Error("Login TokenVerify cache get member error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	if memJson == "" {
		zap.L().Error("Login TokenVerify cache get member expire error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	memberById := &member.Member{}
	json.Unmarshal([]byte(memJson), memberById)
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, memberById)
	memMsg.Code, _ = encrypts.EncryptInt64(memberById.Id, model.AESKey)

	orgsJson, err := ls.cache.Get(context.Background(), model.MemberOrganization+"::"+parseToken)
	if err != nil {
		zap.L().Error("Login TokenVerify cache get organization error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	if orgsJson == "" {
		zap.L().Error("Login TokenVerify cache get organization expire error", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	var orgs []*organization.Organization
	json.Unmarshal([]byte(orgsJson), &orgs)
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	memMsg.CreateTime = tms.FormatByMill(memberById.CreateTime)
	return &login.LoginResponse{Member: memMsg}, nil
}

func (l *LoginService) MyOrgList(ctx context.Context, msg *login.UserMessage) (*login.OrgListResponse, error) {
	memId := msg.MemId
	orgs, err := l.organizationRepo.FindOrganizationByMemId(ctx, memId)
	if err != nil {
		zap.L().Error("MyOrgList FindOrganizationByMemId err", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgsMessage []*login.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	for _, org := range orgsMessage {
		org.Code, _ = encrypts.EncryptInt64(org.Id, model.AESKey)
	}
	return &login.OrgListResponse{OrganizationList: orgsMessage}, nil
}

func (ls *LoginService) FindMemInfoById(ctx context.Context, msg *login.UserMessage) (*login.MemberMessage, error) {
	memberById, err := ls.memberRepo.FindMemberById(context.Background(), msg.MemId)
	if err != nil {
		zap.L().Error("TokenVerify db FindMemById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	// 同样的进行加密
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, memberById)
	memMsg.Code, _ = encrypts.EncryptInt64(memberById.Id, model.AESKey)
	orgs, err := ls.organizationRepo.FindOrganizationByMemId(context.Background(), memberById.Id) // 获取到对应的组织
	if err != nil {
		zap.L().Error("TokenVerify db FindOrganizationByMemId error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if len(orgs) > 0 {
		memMsg.OrganizationCode, _ = encrypts.EncryptInt64(orgs[0].Id, model.AESKey)
	}
	memMsg.CreateTime = tms.FormatByMill(memberById.CreateTime)
	return memMsg, nil
}

func (ls *LoginService) FindMemInfoByIds(ctx context.Context, msg *login.UserMessage) (*login.MemberMessageList, error) {
	memberList, err := ls.memberRepo.FindMemberByIds(context.Background(), msg.MIds)
	if err != nil {
		zap.L().Error("FindMemInfoByIds db memberRepo.FindMemberByIds error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if memberList == nil || len(memberList) <= 0 {
		return &login.MemberMessageList{
			List: nil,
		}, nil
	}
	mMap := make(map[int64]*member.Member)
	for _, v := range memberList {
		mMap[v.Id] = v
	}
	var memMsgs []*login.MemberMessage
	copier.Copy(&memMsgs, memberList)
	for _, v := range memMsgs {
		m := mMap[v.Id]
		v.CreateTime = tms.FormatByMill(m.CreateTime)
		v.Code = encrypts.EncryptNoErr(v.Id)
	}
	return &login.MemberMessageList{
		List: memMsgs,
	}, nil

}

package auth_service_v1

import (
	"context"
	"github.com/jinzhu/copier"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-grpc/auth"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/database/tran"
	"test.com/project-project/internal/domain"
	"test.com/project-project/internal/repo"
)

type AuthService struct {
	auth.UnimplementedAuthServiceServer
	cache             repo.Cache
	transaction       tran.Transaction
	projectAuthDomain *domain.ProjectAuthDomain
}

func New() *AuthService {
	return &AuthService{
		cache:             dao.Rc,
		transaction:       dao.NewTransaction(),
		projectAuthDomain: domain.NewProjectAuthDomain(),
	}
}

func (a *AuthService) AuthList(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ListAuthMessage, error) {
	organizationCode := encrypts.DecryptNoErr(msg.OrganizationCode)
	list, total, err := a.projectAuthDomain.AuthListPage(organizationCode, msg.Page, msg.PageSize)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	var prList []*auth.ProjectAuth
	copier.Copy(&prList, list)
	return &auth.ListAuthMessage{
		List:  prList,
		Total: total,
	}, nil
}

func (a *AuthService) Apply(ctx context.Context, msg *auth.AuthReqMessage) (*auth.ApplyResponse, error) {
	if msg.Action == "getnode" {
		// 获取列表
		list, checkedList, err := a.projectAuthDomain.AllNodeAndAuth(msg.AuthId)
		if err != nil {
			return nil, errs.GrpcError(err)
		}
		var prList []*auth.ProjectNodeMessage
		copier.Copy(&prList, list)
		return &auth.ApplyResponse{
			List:        prList,
			CheckedList: checkedList,
		}, nil
	}
	if msg.Action == "save" {
		// 先删除project_auth_node表，再新增 （两个操作，需要事务）
		nodes := msg.Nodes
		// 先删再存
		authId := msg.AuthId
		err := a.transaction.Action(func(conn database.DbConn) error {
			err := a.projectAuthDomain.Save(conn, authId, nodes)
			return err
		})
		if err != nil {
			return nil, errs.GrpcError(err.(*errs.BError))
		}
	}
	return &auth.ApplyResponse{}, nil
}

func (a *AuthService) AuthNodesByMemberId(ctx context.Context, msg *auth.AuthReqMessage) (*auth.AuthNodesResponse, error) {
	list, err := a.projectAuthDomain.AuthNodes(msg.MemberId)
	if err != nil {
		return nil, errs.GrpcError(err)
	}
	return &auth.AuthNodesResponse{List: list}, nil
}

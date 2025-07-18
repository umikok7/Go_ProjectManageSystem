package domain

import (
	"context"
	"go.uber.org/zap"
	"strconv"
	"test.com/project-common/errs"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/repo"
	"test.com/project-project/pkg/model"
	"time"
)

type ProjectAuthDomain struct {
	userRpcDomain         *UserRpcDomain
	projectAuthRepo       repo.ProjectAuthRepo
	projectNodeDomain     *ProjectNodeDomain
	projectAuthNodeDomain *ProjectAuthNodeDomain
	accountDomain         *AccountDomain
}

func (d *ProjectAuthDomain) AuthList(orgCode int64) ([]*data.ProjectAuthDisplay, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, err := d.projectAuthRepo.FindAuthList(c, orgCode)
	if err != nil {
		zap.L().Error("project AuthList projectAuthRepo.FindAuthList error", zap.Error(err))
		return nil, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, nil
}

func (d *ProjectAuthDomain) AuthListPage(organizationCode int64, page int64, pageSize int64) ([]*data.ProjectAuthDisplay, int64, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, total, err := d.projectAuthRepo.FindAuthListPage(c, organizationCode, page, pageSize)
	if err != nil {
		zap.L().Error("project AuthListPage projectAuthRepo.FindAuthListPage error", zap.Error(err))
		return nil, 0, model.DBError
	}
	var pdList []*data.ProjectAuthDisplay
	for _, v := range list {
		display := v.ToDisplay()
		pdList = append(pdList, display)
	}
	return pdList, total, nil
}

func (d *ProjectAuthDomain) AllNodeAndAuth(authId int64) ([]*data.ProjectNodeAuthTree, []string, *errs.BError) {
	nodeList, err := d.projectNodeDomain.AllNodeList()
	if err != nil {
		return nil, nil, err
	}
	checkedList, err := d.projectAuthNodeDomain.AuthNodeList(authId)
	if err != nil {
		return nil, nil, err
	}
	list := data.ToAuthNodeTreeList(nodeList, checkedList)
	return list, checkedList, nil
}

func (d *ProjectAuthDomain) Save(conn database.DbConn, authId int64, nodes []string) *errs.BError {
	err := d.projectAuthNodeDomain.Save(conn, authId, nodes)
	if err != nil {
		return err
	}
	return nil
}

func (d *ProjectAuthDomain) AuthNodes(memberId int64) ([]string, *errs.BError) {
	account, err := d.accountDomain.FindAccount(memberId)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, model.ParamsError
	}
	authorize := account.Authorize
	authid, _ := strconv.ParseInt(authorize, 10, 64)
	authNodeList, err := d.projectAuthNodeDomain.AuthNodeList(authid)
	if err != nil {
		return nil, model.DBError
	}
	return authNodeList, nil
}

func NewProjectAuthDomain() *ProjectAuthDomain {
	return &ProjectAuthDomain{
		userRpcDomain:         NewUserRpcDomain(),
		projectAuthRepo:       dao.NewProjectAuthDao(),
		projectNodeDomain:     NewProjectNodeDomain(),
		projectAuthNodeDomain: NewProjectAuthNodeDomain(),
		accountDomain:         NewAccountDomain(),
	}
}

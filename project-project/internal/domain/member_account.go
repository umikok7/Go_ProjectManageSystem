package domain

import (
	"context"
	"fmt"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/repo"
	"test.com/project-project/pkg/model"
	"time"
)

type AccountDomain struct {
	accountRepo      repo.AccountRepo
	userRpcDomain    *UserRpcDomain
	departmentDomain *DepartmentDomain
}

func (d *AccountDomain) AccountList(organizationCode string,
	memberId int64,
	page int64,
	pageSize int64,
	departmentCode string,
	searchType int32) ([]*data.MemberAccountDisplay, int64, *errs.BError) {
	condition := ""
	organizationCodeId := encrypts.DecryptNoErr(organizationCode)
	departmentCodeId := encrypts.DecryptNoErr(departmentCode)
	switch searchType {
	case 1:
		condition = "status = 1" // 正在使用的
	case 2:
		condition = "department_code = NULL"
	case 3:
		condition = "status = 0" // 警用的
	case 4:
		condition = fmt.Sprintf("status = 1 and department_code = %d", departmentCodeId)
	default:
		condition = "status = 1"
	}
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, total, err := d.accountRepo.FindList(c, condition, organizationCodeId, departmentCodeId, page, pageSize)
	if err != nil {
		return nil, 0, model.DBError
	}
	var dList []*data.MemberAccountDisplay
	for _, v := range list {
		display := v.ToDisplay()
		memberInfo, _ := d.userRpcDomain.MemberInfo(c, v.MemberCode)
		display.Avatar = memberInfo.Avatar
		if v.DepartmentCode > 0 {
			// 如果有部门
			department, err := d.departmentDomain.FindDepartmentById(v.DepartmentCode)
			if err != nil {
				return nil, 0, err
			}
			display.Departments = department.Name
		}
		dList = append(dList, display)
	}
	return dList, total, nil
}

func (d *AccountDomain) FindAccount(memberId int64) (*data.MemberAccount, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberAccount, err := d.accountRepo.FindByMemberId(c, memberId)
	if err != nil {
		return nil, model.DBError
	}
	return memberAccount, nil
}

func NewAccountDomain() *AccountDomain {
	return &AccountDomain{
		userRpcDomain:    NewUserRpcDomain(),
		accountRepo:      dao.NewMemberAccountDao(),
		departmentDomain: NewDepartmentDomain(),
	}
}

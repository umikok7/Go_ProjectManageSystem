package repo

import (
	"context"
	"test.com/project-project/internal/data"
)

type AccountRepo interface {
	FindList(ctx context.Context, condition string, organizationCode int64, departmentCodeId int64, page int64, pageSize int64) ([]*data.MemberAccount, int64, error)
	FindByMemberId(ctx context.Context, memberId int64) (ma *data.MemberAccount, err error)
}

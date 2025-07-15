package domain

import (
	"context"
	"test.com/project-common/errs"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/database"
	"test.com/project-project/internal/repo"
	"test.com/project-project/pkg/model"
	"time"
)

type ProjectAuthNodeDomain struct {
	projectAuthNodeRepo repo.ProjectAuthNodeRepo
}

func (d *ProjectAuthNodeDomain) AuthNodeList(authId int64) ([]string, *errs.BError) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	list, err := d.projectAuthNodeRepo.FindNodeStringList(c, authId)
	if err != nil {
		return nil, model.DBError
	}
	return list, nil
}

func (d *ProjectAuthNodeDomain) Save(conn database.DbConn, authId int64, nodes []string) *errs.BError {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := d.projectAuthNodeRepo.DeleteByAuthId(c, conn, authId)
	if err != nil {
		return model.DBError
	}
	err = d.projectAuthNodeRepo.Save(c, conn, authId, nodes)
	if err != nil {
		return model.DBError
	}
	return nil
}

func NewProjectAuthNodeDomain() *ProjectAuthNodeDomain {
	return &ProjectAuthNodeDomain{
		projectAuthNodeRepo: dao.NewProjectAuthNodeDao(),
	}
}

package dao

import (
	"context"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database/gorms"
)

type ProjectNode struct {
	conn *gorms.GormConn
}

func (p *ProjectNode) FindAll(ctx context.Context) (pms []*data.ProjectNode, err error) {
	session := p.conn.Session(ctx)
	err = session.Model(&data.ProjectNode{}).Find(&pms).Error
	return
}

func NewProjectNodeDao() *ProjectNode {
	return &ProjectNode{
		conn: gorms.New(),
	}
}

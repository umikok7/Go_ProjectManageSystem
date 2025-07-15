package dao

import (
	"context"
	"test.com/project-project/internal/data"
	"test.com/project-project/internal/database/gorms"
)

type MenuDao struct {
	conn *gorms.GormConn
}

func (m MenuDao) FindMenus(ctx context.Context) (pms []*data.ProjectMenu, err error) {
	session := m.conn.Session(ctx)
	// pms 作为结果接收者，接收查询到的所有菜单数据
	err = session.Order("pid,sort asc, id asc").Find(&pms).Error
	return
}

func NewMenuDao() *MenuDao {
	return &MenuDao{
		conn: gorms.New(),
	}
}

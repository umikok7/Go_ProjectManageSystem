package dao

import (
	"context"
	"gorm.io/gorm"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/gorms"
)

type memberDao struct {
	conn *gorms.GormConn
}

func (m *memberDao) FindMemberByIds(background context.Context, ids []int64) (list []*member.Member, err error) {
	if len(ids) <= 0 {
		return nil, nil
	}
	err = m.conn.Session(background).Model(&member.Member{}).Where("id in (?)", ids).First(&list).Error
	return
}

func (m *memberDao) FindMemberById(ctx context.Context, id int64) (mem *member.Member, err error) {
	err = m.conn.Session(ctx).Where("id=?", id).First(&mem).Error
	return mem, err
}

func (m *memberDao) FindMember(ctx context.Context, account string, pwd string) (*member.Member, error) {
	var mem *member.Member
	err := m.conn.Session(ctx).Where("account=? and password=?", account, pwd).First(&mem).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return mem, err
}

func NewMemberDao() *memberDao {
	return &memberDao{
		conn: gorms.New(),
	}
}

func (m *memberDao) SaveMember(conn database.DbConn, ctx context.Context, mem *member.Member) error {
	//TODO implement me
	m.conn = conn.(*gorms.GormConn)
	return m.conn.Tx(ctx).Create(mem).Error
}

func (m *memberDao) GetMemberByEmail(ctx context.Context, email string) (bool, error) {
	//TODO implement me
	var count int64
	err := m.conn.Session(ctx).Model(&member.Member{}).Where("email=?", email).Count(&count).Error
	return count > 0, err
}

func (m *memberDao) GetMemberByAccount(ctx context.Context, account string) (bool, error) {
	//TODO implement me
	var count int64
	err := m.conn.Session(ctx).Model(&member.Member{}).Where("account=?", account).Count(&count).Error
	return count > 0, err
}

func (m *memberDao) GetMemberByMobile(ctx context.Context, mobile string) (bool, error) {
	//TODO implement me
	var count int64
	err := m.conn.Session(ctx).Model(&member.Member{}).Where("mobile=?", mobile).Count(&count).Error
	return count > 0, err
}

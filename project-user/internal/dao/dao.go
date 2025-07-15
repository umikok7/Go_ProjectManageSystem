package dao

import (
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/gorms"
)

type TransactionImpl struct {
	conn database.DbConn
}

func NewTransaction() *TransactionImpl {
	return &TransactionImpl{
		conn: gorms.NewTran(),
	}
}

func (t TransactionImpl) Action(f func(conn database.DbConn) error) error {
	//TODO implement me
	t.conn.Begin()
	err := f(t.conn)
	if err != nil {
		t.conn.Rollback()
		return err
	}
	t.conn.Commit()
	return nil
}

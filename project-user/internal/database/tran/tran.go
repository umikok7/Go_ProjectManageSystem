package tran

import "test.com/project-user/internal/database"

// Transaction 事务的操作一定与数据库有关 注入数据库的连接 gorm.db
type Transaction interface {
	Action(func(conn database.DbConn) error) error
}

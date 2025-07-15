package gorms

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"test.com/project-user/config"
)

var _db *gorm.DB

func init() {
	// 配置mysql连接参数，此处与docker开的数据库相对应
	username := config.C.MysqlConfig.Username
	password := config.C.MysqlConfig.Password
	host := config.C.MysqlConfig.Host
	port := config.C.MysqlConfig.Port
	Dbname := config.C.MysqlConfig.Db
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
	var err error
	_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("数据库连接失败， error = " + err.Error())
	}
}

func GetDB() *gorm.DB {
	return _db
}

type GormConn struct {
	db *gorm.DB // 普通连接 用于非事务操作
	tx *gorm.DB // 事务连接 用于事务操作
}

func (g *GormConn) Begin() {
	g.tx = GetDB().Begin()
}

func New() *GormConn {
	return &GormConn{
		db: GetDB(),
	}
}

func NewTran() *GormConn {
	return &GormConn{db: GetDB(), tx: GetDB()}
}

func (g *GormConn) Session(ctx context.Context) *gorm.DB {
	return g.db.Session(&gorm.Session{Context: ctx})
}

func (g *GormConn) Rollback() {
	g.tx.Rollback()
}

func (g *GormConn) Commit() {
	g.tx.Commit()
}

func (g *GormConn) Tx(ctx context.Context) *gorm.DB {
	return g.tx.WithContext(ctx) // 将上下文绑定到事务的连接
}

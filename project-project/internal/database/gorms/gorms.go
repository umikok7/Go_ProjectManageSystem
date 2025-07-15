package gorms

import (
	"context"
	"gorm.io/gorm"
)

var _db *gorm.DB

//func init() {
//	// 配置mysql读写分离
//	if config.C.DbConfig.Separation {
//		// 说明开启了读写分离
//		// Master操作
//		username := config.C.DbConfig.Master.Username
//		password := config.C.DbConfig.Master.Password
//		host := config.C.DbConfig.Master.Host
//		port := config.C.DbConfig.Master.Port
//		Dbname := config.C.DbConfig.Master.Db
//		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
//		var err error
//		_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//			Logger: logger.Default.LogMode(logger.Info),
//		})
//		if err != nil {
//			panic("数据库连接失败， error = " + err.Error())
//		}
//		// Slave操作
//		replicas := []gorm.Dialector{}
//		for _, v := range config.C.DbConfig.Slave {
//			username := v.Username
//			password := v.Password
//			host := v.Host
//			Dbname := v.Db
//			port := v.Port
//			dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
//			cfg := mysql.Config{
//				DSN: dsn,
//			}
//			replicas = append(replicas, mysql.New(cfg))
//		}
//		_db.Use(dbresolver.Register(dbresolver.Config{
//			Sources: []gorm.Dialector{mysql.New(mysql.Config{
//				DSN: dsn,
//			})}, //主库配置，用于写操作
//			Replicas: replicas,                  // 从库配置，用于读操作
//			Policy:   dbresolver.RandomPolicy{}, // 负载均衡策略
//		}).SetMaxIdleConns(10).SetMaxOpenConns(200)) // 连接池操作
//	} else {
//		// 无读写分离的配置
//		// 配置mysql连接参数，此处与docker开的数据库相对应
//		username := config.C.MysqlConfig.Username
//		password := config.C.MysqlConfig.Password
//		host := config.C.MysqlConfig.Host
//		port := config.C.MysqlConfig.Port
//		Dbname := config.C.MysqlConfig.Db
//		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", username, password, host, port, Dbname)
//		var err error
//		_db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
//			Logger: logger.Default.LogMode(logger.Info),
//		})
//		if err != nil {
//			panic("数据库连接失败， error = " + err.Error())
//		}
//	}
//}

func GetDB() *gorm.DB {
	return _db
}

func SetDB(db *gorm.DB) {
	_db = db
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

package config

import (
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"log"
	"os"
	"test.com/project-common/logs"
)

var C = InitConfig()

type Config struct {
	viper        *viper.Viper
	SC           *ServerConfig
	GC           *GrpcConfig
	EtcdConfig   *EtcdConfig
	MysqlConfig  *MysqlConfig
	JwtConfig    *JwtConfig
	JaegerConfig *JaegerConfig
}

type ServerConfig struct {
	Name string
	Port string
}

type GrpcConfig struct {
	Addr    string
	Name    string
	Version string
	Weight  int64
}

type EtcdConfig struct {
	Addrs []string
}

type MysqlConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Db       string
}

type JwtConfig struct {
	AccessExp     int
	RefreshExp    int
	AccessSecret  string
	RefreshSecret string
}

type JaegerConfig struct {
	Endpoint    string // Jaeger Collector 端点
	ServiceName string // 服务名称
	Environment string // 部署环境
	Enabled     bool   // 是否启用链路追踪
}

func InitConfig() *Config {
	conf := &Config{
		viper: viper.New(),
	}
	workdir, _ := os.Getwd()
	conf.viper.SetConfigName("config")
	conf.viper.SetConfigType("yaml")
	conf.viper.AddConfigPath("/Users/umikok/Desktop/GoLand/ms_project/project-user/config")
	conf.viper.AddConfigPath(workdir + "/config")
	log.Println("Config path:", workdir+"/config")

	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	// 调用对应的初始化
	conf.ReadServerConfig()
	conf.ReadGrpcConfig()
	conf.ReadEtcdConfig()
	conf.InitMysqlConfig()
	conf.InitJwtConfig()
	conf.ReadJaegerConfig()
	return conf
}

func (c *Config) InitZapLog() {
	//从配置中读取日志配置，初始化日志
	lc := &logs.LogConfig{
		DebugFileName: c.viper.GetString("zap.debugFileName"),
		InfoFileName:  c.viper.GetString("zap.infoFileName"),
		WarnFileName:  c.viper.GetString("zap.warnFileName"),
		MaxSize:       c.viper.GetInt("maxSize"),
		MaxAge:        c.viper.GetInt("maxAge"),
		MaxBackups:    c.viper.GetInt("maxBackups"),
	}
	err := logs.InitLogger(lc)
	if err != nil {
		log.Fatalln(err)
	}
}

func (c *Config) InitRedisConfig() *redis.Options {
	return &redis.Options{
		Addr:     c.viper.GetString("redis.host") + ":" + c.viper.GetString("redis.port"),
		Password: c.viper.GetString("redis.password"),
		DB:       c.viper.GetInt("redis.db"), // use default db
	}
}

func (c *Config) InitMysqlConfig() {
	mc := &MysqlConfig{
		Username: c.viper.GetString("mysql.username"),
		Password: c.viper.GetString("mysql.password"),
		Host:     c.viper.GetString("mysql.host"),
		Port:     c.viper.GetInt("mysql.port"),
		Db:       c.viper.GetString("mysql.db"),
	}
	c.MysqlConfig = mc
}

func (c *Config) ReadServerConfig() {
	sc := &ServerConfig{}
	sc.Name = c.viper.GetString("server.name")
	sc.Port = c.viper.GetString("server.port")
	c.SC = sc
}

func (c *Config) ReadGrpcConfig() {
	gc := &GrpcConfig{}
	gc.Name = c.viper.GetString("grpc.name")
	gc.Addr = c.viper.GetString("grpc.addr")
	gc.Version = c.viper.GetString("grpc.version")
	gc.Weight = c.viper.GetInt64("grpc.weight")
	c.GC = gc
}

func (c *Config) ReadEtcdConfig() {
	ec := &EtcdConfig{}
	var addrs []string
	err := c.viper.UnmarshalKey("etcd.addrs", &addrs)
	if err != nil {
		log.Fatalln(err)
	}
	ec.Addrs = addrs
	c.EtcdConfig = ec
}

func (c *Config) InitJwtConfig() {
	jwt := &JwtConfig{
		AccessExp:     c.viper.GetInt("jwt.accessExp"),
		RefreshExp:    c.viper.GetInt("jwt.refreshExp"),
		AccessSecret:  c.viper.GetString("jwt.accessSecret"),
		RefreshSecret: c.viper.GetString("jwt.refreshSecret"),
	}
	c.JwtConfig = jwt
}

func (c *Config) ReadJaegerConfig() {
	jc := &JaegerConfig{}
	jc.Endpoint = c.viper.GetString("jaeger.endpoint")
	jc.ServiceName = c.viper.GetString("jaeger.serviceName")
	jc.Environment = c.viper.GetString("jaeger.environment")
	jc.Enabled = c.viper.GetBool("jaeger.enabled")
	c.JaegerConfig = jc
}

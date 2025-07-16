package config

import (
	"bytes"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
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
	DbConfig     *DbConfig
	JaegerConfig *JaegerConfig
	KafkaConfig  *KafkaConfig
	MinioConfig  *MinioConfig
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

type DbConfig struct {
	Master     MysqlConfig
	Slave      []MysqlConfig
	Separation bool
}

type JaegerConfig struct {
	Endpoint    string // Jaeger Collector 端点
	ServiceName string // 服务名称
	Environment string // 部署环境
	Enabled     bool   // 是否启用链路追踪
}

type KafkaConfig struct {
	Addr  string
	Topic string
}

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
}

func InitConfig() *Config {
	conf := &Config{
		viper: viper.New(),
	}
	// 先从nacos读取配置，如果读取不到就在本地读取
	nacosClient := InitNacosClient()
	configYaml, err2 := nacosClient.confClient.GetConfig(vo.ConfigParam{
		DataId: "config.yaml",
		Group:  nacosClient.group,
	})
	if err2 != nil {
		log.Fatalln(err2)
	}
	// 启用监听，监听变化
	err2 = nacosClient.confClient.ListenConfig(vo.ConfigParam{
		DataId: "config.yaml",
		Group:  nacosClient.group,
		OnChange: func(namespace, group, dataId, data string) {
			log.Printf("load nacos config changed %s \n", data)
			err := conf.viper.ReadConfig(bytes.NewBuffer([]byte(data)))
			if err != nil {
				log.Printf("load nacos config changed, err: %s \n", err)
			}
			// 所有配置应该重新读取
			conf.ReLoadAllConfig()
		},
	})
	if err2 != nil {
		log.Fatalln(err2)
	}

	conf.viper.SetConfigType("yaml")
	if configYaml != "" {
		// 说明有nacos有值
		err := conf.viper.ReadConfig(bytes.NewBuffer([]byte(configYaml)))
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("load nacos config %s\n", configYaml)
	} else {
		// 说明在nacos中读取不到，需要进行本地读取
		workdir, _ := os.Getwd()
		conf.viper.SetConfigName("config")
		conf.viper.AddConfigPath("/Users/umikok/Desktop/GoLand/ms_project/project-project/config")
		conf.viper.AddConfigPath(workdir + "/config")
		log.Println("Config path:", workdir+"/config")

		err := conf.viper.ReadInConfig()
		if err != nil {
			log.Fatalln(err)
		}
	}
	// 调用对应的初始化
	conf.ReLoadAllConfig()
	return conf
}

func (c *Config) ReLoadAllConfig() {
	c.ReadServerConfig()
	c.InitZapLog()
	c.ReadGrpcConfig()
	c.ReadEtcdConfig()
	c.InitMysqlConfig()
	c.InitJwtConfig()
	c.InitDbConfig()
	c.ReadJaegerConfig()
	c.ReadKafkaConfig()
	c.ReadMinioConfig()
	// 重新创建相关的客户端
	c.ReConnRedis()
	c.ReConnMysql()
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

func (c *Config) InitDbConfig() {
	mc := DbConfig{}
	mc.Separation = c.viper.GetBool("db.separation")
	var slaves []MysqlConfig
	err := c.viper.UnmarshalKey("db.slave", &slaves)
	if err != nil {
		panic(err)
	}
	master := MysqlConfig{
		Username: c.viper.GetString("db.master.username"),
		Password: c.viper.GetString("db.master.password"),
		Host:     c.viper.GetString("db.master.host"),
		Port:     c.viper.GetInt("db.master.port"),
		Db:       c.viper.GetString("db.master.db"),
	}
	mc.Master = master
	mc.Slave = slaves
	c.DbConfig = &mc
}

func (c *Config) ReadJaegerConfig() {
	jc := &JaegerConfig{}
	jc.Endpoint = c.viper.GetString("jaeger.endpoint")
	jc.ServiceName = c.viper.GetString("jaeger.serviceName")
	jc.Environment = c.viper.GetString("jaeger.environment")
	jc.Enabled = c.viper.GetBool("jaeger.enabled")
	c.JaegerConfig = jc
}

func (c *Config) ReadKafkaConfig() {
	kc := &KafkaConfig{}
	kc.Addr = c.viper.GetString("kafka.addr")
	kc.Topic = c.viper.GetString("kafka.topic")
	c.KafkaConfig = kc
}

func (c *Config) ReadMinioConfig() {
	mc := &MinioConfig{}
	mc.AccessKey = c.viper.GetString("minio.accessKey")
	mc.SecretKey = c.viper.GetString("minio.secretKey")
	mc.Endpoint = c.viper.GetString("minio.endPoint")
	mc.BucketName = c.viper.GetString("minio.bucketName")
	c.MinioConfig = mc
}

package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"test.com/project-common/logs"
)

var C = InitConfig()

type Config struct {
	viper        *viper.Viper
	SC           *ServerConfig
	EtcdConfig   *EtcdConfig
	JaegerConfig *JaegerConfig
}

type ServerConfig struct {
	Name string
	Port string
}

type EtcdConfig struct {
	Addrs []string
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
	conf.viper.AddConfigPath("/Users/umikok/Desktop/GoLand/ms_project/project-api/config")
	conf.viper.AddConfigPath(workdir + "/config")
	log.Println("Config path:", workdir+"/config")

	err := conf.viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
	// 初始化
	conf.ReadServerConfig()
	conf.ReadEtcdConfig()
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

// 初始化部分

func (c *Config) ReadServerConfig() {
	sc := &ServerConfig{}
	sc.Name = c.viper.GetString("server.name")
	sc.Port = c.viper.GetString("server.port")
	c.SC = sc
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

func (c *Config) ReadJaegerConfig() {
	jc := &JaegerConfig{}
	jc.Endpoint = c.viper.GetString("jaeger.endpoint")
	jc.ServiceName = c.viper.GetString("jaeger.serviceName")
	jc.Environment = c.viper.GetString("jaeger.environment")
	jc.Enabled = c.viper.GetBool("jaeger.enabled")
	c.JaegerConfig = jc

}

package config

import (
	"fmt"
	"log"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type NacosClient struct {
	confClient config_client.IConfigClient
	group      string
}

func InitNacosClient() *NacosClient {
	//create clientConfig
	bootConf := InitBootstrap()
	// 使用时间戳创建唯一的缓存目录
	timestamp := time.Now().Format("20060102-150405")
	cacheDir := fmt.Sprintf("/tmp/nacos/cache-%s", timestamp)
	logDir := fmt.Sprintf("/tmp/nacos/log-%s", timestamp)
	clientConfig := constant.ClientConfig{
		NamespaceId:         bootConf.NacosConfig.Namespace, //we can create multiple clients with different namespaceId to support multiple namespace.When namespace is public, fill in the blank string here.
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              logDir,
		CacheDir:            cacheDir,
		LogLevel:            "debug",
	}
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      bootConf.NacosConfig.IpAddr,
			ContextPath: bootConf.NacosConfig.ContextPath,
			Port:        uint64(bootConf.NacosConfig.Port),
			Scheme:      bootConf.NacosConfig.Scheme,
		},
	}
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Fatalln(err)
	}
	nc := &NacosClient{
		confClient: configClient,
		group:      bootConf.NacosConfig.Group,
	}
	return nc
}

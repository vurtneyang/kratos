package nacos

import (
	"kratos/pkg/log"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
)

// AliNacosLinkConf 连接阿里云nacos的配置
type AliNacosLinkConf struct {
	Endpoint    string `toml:"endpoint"`
	NamespaceID string `toml:"namespace_id"`
	AccessKey   string `toml:"access_key"`
	SecretKey   string `toml:"secret_key"`
	DataID      string `toml:"data_id"`
	Group       string `toml:"group"`
	CacheDir    string `toml:"cache_dir"`
	LogPath     string `toml:"log_path"`
}

// NewNacosConfigClient 创建新Nacos客户端
func NewNacosConfigClient(linkConf *AliNacosLinkConf) (config_client.IConfigClient, error) {
	sc := []constant.ServerConfig{
		{
			IpAddr: linkConf.Endpoint,
			Port:   8848,
		},
	}
	clientConfig := constant.ClientConfig{
		NamespaceId: linkConf.NamespaceID,
		AccessKey:   linkConf.AccessKey,
		SecretKey:   linkConf.SecretKey,
		TimeoutMs:   5 * 1000,
		LogDir:      linkConf.LogPath,
		LogLevel:    "error",
		CacheDir:    linkConf.CacheDir,
	}

	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: sc,
		},
	)

	return configClient, err
}

func isPtr(i interface{}) bool {
	vi := reflect.ValueOf(i)
	return vi.Kind() == reflect.Ptr
}

// NewNacosConfig 从阿里云获取配置
func NewNacosConfig(linkConf *AliNacosLinkConf, projectConfigPtr interface{}) (interface{}, error) {

	if !isPtr(projectConfigPtr) {
		return nil,errors.New("need ptr config struct")
	}

	configClient, err := NewNacosConfigClient(linkConf)
	if err != nil {
		log.Error("init nacos config client error,err:%v", err)
		return nil, err
	}

	// -------- 获取配置,监听配置更新 --------
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: linkConf.DataID,
		Group:  linkConf.Group})

	if err != nil {
		log.Error("nacos get config content error,err:%v", err)
		return nil, err
	}

	_, err = toml.Decode(content, projectConfigPtr)
	if err != nil {
		log.Error("nacos toml decode failed,err:%v", err)
		return nil, err
	}

	log.Info("init nacos config success")

	// 监听配置
	err = configClient.ListenConfig(vo.ConfigParam{
		DataId: linkConf.DataID,
		Group:  linkConf.Group,
		OnChange: func(namespace, group, dataId, data string) {
			_, err = toml.Decode(data, projectConfigPtr)
			if err != nil {
				log.Error("update nacos config failed,err:%v,data:%+v", err, data)
			}
		},
	})
	if err != nil {
		log.Error("nacos listen failed,err: " + err.Error())
		return nil, err
	}

	return projectConfigPtr, nil
}
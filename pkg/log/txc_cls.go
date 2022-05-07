package log

import (
	"context"
	"errors"
	"fmt"
	sysLog "log"
	"time"

	"kratos/pkg/conf/paladin"

	cls "github.com/tencentcloud/tencentcloud-cls-sdk-go"
)

type txcCls struct {
	Config   *CLSConfig `toml:"TxcCLS"`
	producer *cls.AsyncProducerClient
}

type CLSConfig struct {
	// https://github.com/TencentCloud/tencentcloud-cls-sdk-go/blob/main/async_producer_client_config.go
	RecordLevel     Level  `toml:"RecordLevel"`
	Endpoint        string `toml:"Endpoint"`
	AccessKeyID     string `toml:"AccessKeyID"`
	AccessKeySecret string `toml:"AccessKeySecret"`
	TopicId         string `toml:"TopicId"`
}

func NewCLSCfg() (cfs *CLSConfig) {
	cfg := &txcCls{}
	paladin.Get("application.toml").UnmarshalTOML(&cfg)

	return cfg.Config
}

// NewCLS create a sls log handler
func NewCLS(cfg *CLSConfig) (*txcCls, error) {
	a := &txcCls{}
	if cfg == nil {
		return nil, errors.New("Configuration cannot be nil ")
	}
	clsConfig := cls.GetDefaultAsyncProducerClientConfig()

	if cfg.Endpoint == "" || cfg.AccessKeySecret == "" || cfg.AccessKeyID == "" || cfg.TopicId == "" {
		return nil, errors.New("Endpoint, AccessKeySecret, AccessKeyID TopicId cannot be empty ")
	}

	clsConfig.Endpoint = cfg.Endpoint
	clsConfig.AccessKeyID = cfg.AccessKeyID
	clsConfig.AccessKeySecret = cfg.AccessKeySecret

	a.Config = cfg
	pclient, err := cls.NewAsyncProducerClient(clsConfig)
	if err != nil {
		panic(err)
	}
	pclient.Start() // 启动producer实例
	a.producer = pclient
	return a, nil
}

// Log logging to cls
func (a *txcCls) Log(ctx context.Context, lv Level, args ...D) {
	if args == nil {
		return
	}
	if a.Config.RecordLevel != 0 {
		if a.Config.RecordLevel > lv {
			return
		}
	}
	d := toMap(args...)
	addExtraField(ctx, d)
	var log *cls.Log
	content := make(map[string]string)
	// create log
	for k, v := range d {
		content[k] = fmt.Sprintf("%v", v)
	}

	log = cls.NewCLSLog(time.Now().Unix(), content)
	err := a.producer.SendLog(a.Config.TopicId, log, nil)
	if err != nil {
		sysLog.Printf("log producer error(%v)", err)
	}
}

func (a *txcCls) SetFormat(string) {

}

// Close log handler
func (a *txcCls) Close() error {
	a.producer.Close(6000)

	return nil
}

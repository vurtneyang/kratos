package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/VurtneYang/kratos/pkg/conf/paladin"

	sls "github.com/aliyun/aliyun-log-go-sdk"
	"github.com/aliyun/aliyun-log-go-sdk/producer"
	"github.com/gogo/protobuf/proto"

	sysLog "log"
)

type aliSLS struct {
	Config   *SLSConfig	`toml:"LogaliSLS"`
	producer *producer.Producer
}

type SLSConfig struct {
	// https://github.com/aliyun/aliyun-log-go-sdk/tree/master/producer
	ProducerConfig *producer.ProducerConfig `toml:"ProducerConfig"`
	// producer safe close
	Safe      bool	`toml:"Safe"`
	TimeoutMs int64	`toml:"TimeoutMs"`
	// Logging level
	RecordLevel     Level	`toml:"RecordLevel"`
	Endpoint        string	`toml:"Endpoint"`
	AccessKeyID     string	`toml:"AccessKeyID"`
	AccessKeySecret string	`toml:"AccessKeySecret"`
	ProjectName     string	`toml:"ProjectName"`
	LogtorName      string	`toml:"LogtorName"`
	Topic           string	`toml:"Topic"`
}


func NewSLSCfg()(cfs *SLSConfig){
	cfg := &aliSLS{}
	paladin.Get("application.toml").UnmarshalTOML(&cfg)
	return cfg.Config
}


// NewAliSLS create a sls log handler
func NewAliSLS(cfg *SLSConfig) (*aliSLS, error) {
	if cfg == nil {
		return nil, errors.New("Configuration cannot be nil ")
	}
	if cfg.ProducerConfig == nil {
		cfg.ProducerConfig = producer.GetDefaultProducerConfig()
	}
	a := &aliSLS{}
	if cfg.Endpoint == "" || cfg.AccessKeySecret == "" || cfg.AccessKeyID == "" {
		return nil, errors.New("Endpoint, AccessKeySecret, AccessKeyID cannot be empty ")
	}
	if cfg.ProjectName == "" || cfg.LogtorName == "" {
		return nil, errors.New("ProjectName, LogtorName cannot be empty ")
	}
	cfg.ProducerConfig.Endpoint = cfg.Endpoint
	cfg.ProducerConfig.AccessKeyID = cfg.AccessKeyID
	cfg.ProducerConfig.AccessKeySecret = cfg.AccessKeySecret
	cfg.ProducerConfig.AllowLogLevel = "error"
	a.Config = cfg
	producerInstance := producer.InitProducer(cfg.ProducerConfig)
	producerInstance.Start() // 启动producer实例
	a.producer = producerInstance
	return a, nil
}

// Log logging to aliyun sls
func (a *aliSLS) Log(ctx context.Context, lv Level, args ...D) {
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
	var log *sls.Log
	var content []*sls.LogContent
	// create log
	for k, v := range d {
		logKV := &sls.LogContent{
			Key:   proto.String(k),
			Value: proto.String(fmt.Sprintf("%v", v)),
		}
		content = append(content, logKV)
	}
	log = &sls.Log{
		Time:     proto.Uint32(uint32(time.Now().Unix())),
		Contents: content,
	}

	source := d[_source].(string)
	err := a.producer.SendLog(a.Config.ProjectName, a.Config.LogtorName, a.Config.Topic, source, log)
	if err != nil {
		sysLog.Printf("log producer error(%v)", err)
	}
}

func (a *aliSLS) SetFormat(string) {

}

// Close log handler
func (a *aliSLS) Close() error {
	if a.Config.Safe {
		a.producer.SafeClose()
	} else {
		err := a.producer.Close(a.Config.TimeoutMs)
		if err != nil {
			return err
		}
	}
	return nil
}

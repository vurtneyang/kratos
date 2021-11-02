package log

import (
	"context"
	"testing"

	"kratos/pkg/conf/env"
)

func TestMain(m *testing.M) {
	conf := &SLSConfig{
		ProducerConfig:  nil,
		Safe:            true,
		TimeoutMs:       0,
		RecordLevel:     1,
		Endpoint:        "cn-hangzhou.log.aliyuncs.com",
		AccessKeyID:     "",
		AccessKeySecret: "",
		ProjectName:     "",
		LogtorName:      "",
		Topic:           "",
	}
	a, err := NewAliSLS(conf)
	if err != nil {
		panic(err)
	}
	env.AppID = "test.service"
	logConf := &Config{
		Host:   "192.168.1.1",
		Stdout: false,
	}
	Init(logConf, a)
	m.Run()
}

func TestAliSLS_Log(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		l := Warnv
		l(ctx, KVString("source", "test1"), KVInt("int", i))
	}
	_ = Close()
}

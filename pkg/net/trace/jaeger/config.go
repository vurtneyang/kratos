package jaeger

import (
	"flag"
	"fmt"
	"os"

	"kratos/pkg/conf/env"
	"kratos/pkg/net/trace"
)

var (
	_jaegerAppID    = env.AppID
	_jaegerEndpoint = "http://127.0.0.1:9191"
)

func init() {
	if v := os.Getenv("JAEGER_ENDPOINT"); v != "" {
		_jaegerEndpoint = v
	}

	if v := os.Getenv("APP_NAME"); v != "" {
		group := os.Getenv("APP_GROUP")
		de := os.Getenv("DEPLOY_ENV")
		_jaegerAppID = fmt.Sprintf("%s-%s-%s", group, v, de)
	}

	flag.StringVar(&_jaegerEndpoint, "jaeger_endpoint", _jaegerEndpoint, "jaeger report endpoint, or use JAEGER_ENDPOINT env.")
	flag.StringVar(&_jaegerAppID, "jaeger_appid", _jaegerAppID, "jaeger report appid, or use JAEGER_APPID env.")
}

// Init Init
func Init() {
	c := &Config{Endpoint: _jaegerEndpoint, BatchSize: 200}
	trace.SetGlobalTracer(trace.NewTracer(_jaegerAppID, newReport(c), true))
}

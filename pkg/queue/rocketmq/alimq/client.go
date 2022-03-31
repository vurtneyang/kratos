package alimq

import mqHttpSdk "github.com/aliyunmq/mq-http-go-sdk"

type Config struct {
	Endpoint    string
	AccessKey   string
	SecretKey   string
	SecretToken string
}
func newClient(c *Config) mqHttpSdk.MQClient {
	return mqHttpSdk.NewAliyunMQClient(c.Endpoint, c.AccessKey, c.SecretKey, c.SecretToken)
}

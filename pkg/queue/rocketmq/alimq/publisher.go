package alimq

import (
	mqHttpSdk "github.com/aliyunmq/mq-http-go-sdk"
)

type Publisher struct {
	config   *PublisherConfig
	producer mqHttpSdk.MQProducer
}

type PublisherConfig struct {
	*Config
	Topic      string
	InstanceId string
}

func NewPublisher(c *PublisherConfig) (p *Publisher) {
	client := newClient(c.Config)
	producer := client.GetProducer(c.InstanceId, c.Topic)

	return &Publisher{
		config:   c,
		producer: producer,
	}
}

func (p *Publisher) Publish(req mqHttpSdk.PublishMessageRequest) (resp mqHttpSdk.PublishMessageResponse, err error) {
	return p.producer.PublishMessage(req)
}

func (p *Publisher) Close() {}

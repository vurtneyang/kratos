package txmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"kratos/pkg/log"
)

type Publisher struct {
	config   *PublisherConfig
	producer rocketmq.Producer
}

type Config struct {
	ServerAddress string
	AccessKey     string
	SecretKey     string
	NameSpace     string
}

type PublisherConfig struct {
	*Config
	Retries int // 重试次数
}

type PublishMessageRequest struct {
	Topic          string
	MessageBody    string
	DelayTimeLevel int
	MessageTag     string
	MessageKeys    []string
	ShardingKey    string
}

func NewPublisher(c *PublisherConfig) (p *Publisher, err error) {
	pr, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{c.ServerAddress})),
		producer.WithCredentials(primitive.Credentials{
			SecretKey: c.SecretKey,
			AccessKey: c.AccessKey,
		}),
		producer.WithNamespace(c.NameSpace),
		producer.WithRetry(c.Retries),
	)

	if err != nil {
		return
	}

	err = pr.Start()
	if err != nil {
		return
	}

	p = &Publisher{
		config:   c,
		producer: pr,
	}


	return
}

func (p *Publisher) Publish(ctx context.Context, req *PublishMessageRequest) (resp *primitive.SendResult, err error) {
	msg := primitive.NewMessage(req.Topic, []byte(req.MessageBody))
	msg.WithDelayTimeLevel(req.DelayTimeLevel)
	msg.WithTag(req.MessageTag)
	msg.WithKeys(req.MessageKeys)
	msg.WithShardingKey(req.ShardingKey)

	return p.producer.SendSync(ctx, msg)
}

func (p *Publisher) Close() {
	err := p.producer.Shutdown()
	if err != nil {
		log.Error("[Publisher]close err:%+v", err)
	} else {
		log.Info("[Publisher]close")
	}
}

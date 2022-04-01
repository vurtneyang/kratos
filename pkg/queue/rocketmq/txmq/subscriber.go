package txmq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"kratos/pkg/log"
	"sync/atomic"
)

type Subscriber struct {
	config   *SubscriberConfig
	consumer rocketmq.PushConsumer
	status   int32
}

type SubscriberConfig struct {
	*Config
	Topic     string
	GroupName string
}

func NewSubscriber(c *SubscriberConfig) (s *Subscriber, err error) {
	cs, err := rocketmq.NewPushConsumer(
		// 设置消费者组
		consumer.WithGroupName(c.GroupName),
		// 设置服务地址
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{c.ServerAddress})),
		// 设置acl权限
		consumer.WithCredentials(primitive.Credentials{
			SecretKey: c.SecretKey,
			AccessKey: c.AccessKey,
		}),
		// 设置命名空间名称
		consumer.WithNamespace(c.NameSpace),
		// 设置从起始位置开始消费
		//consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
		// 设置消费模式（默认集群模式）
		//consumer.WithConsumerModel(consumer.Clustering),
	)

	if err != nil {
		return
	}

	s = &Subscriber{
		config:   c,
		consumer: cs,
	}

	return
}

func (s *Subscriber) Subscribe(handler func(context.Context, ...*primitive.MessageExt) (consumer.ConsumeResult, error)) error {
	selector := consumer.MessageSelector{}

	err := s.consumer.Subscribe(s.config.Topic, selector, handler)
	if err != nil {
		return err
	}

	err = s.consumer.Start()
	if err != nil {
		return err
	}

	atomic.StoreInt32(&s.status, stateRunning)

	return nil
}

func (s *Subscriber) Close() {
	var err error

	if atomic.LoadInt32(&s.status) == stateRunning {
		err = s.consumer.Shutdown()
	}

	if err != nil {
		log.Error("[Subscriber]topic:%s, groupName:%s close err:%v", s.config.Topic, s.config.GroupName, err)
	} else {
		log.Info("[Subscriber]topic:%s, groupName:%s close", s.config.Topic, s.config.GroupName)
	}
}

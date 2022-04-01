package alimq

import (
	"context"
	mqHttpSdk "github.com/aliyunmq/mq-http-go-sdk"
	"github.com/gogap/errors"
	"kratos/pkg/log"
	"kratos/pkg/sync/errgroup"
	"strings"
	"sync/atomic"
)

type Subscriber struct {
	config      *SubscriberConfig
	consumer    mqHttpSdk.MQConsumer
	messageChan chan mqHttpSdk.ConsumeMessageResponse
	errChan     chan error
	stopChan    chan bool
	handlerFunc func([]mqHttpSdk.ConsumeMessageEntry)
	status      int32
	wg          *errgroup.Group
}

type SubscriberConfig struct {
	*Config
	InstanceId    string
	Topic         string
	GroupId       string
	MessageTag    string
	NumOfMessages int32 // 一次最多消费x条（最多可设置为16条）
	WaitSeconds   int64 // 长轮询时间xs（最多可设置为30s）
}

func NewSubscriber(c *SubscriberConfig) (s *Subscriber) {
	client := newClient(c.Config)
	consumer := client.GetConsumer(c.InstanceId, c.Topic, c.GroupId, c.MessageTag)

	return &Subscriber{
		config:      c,
		consumer:    consumer,
		messageChan: make(chan mqHttpSdk.ConsumeMessageResponse),
		errChan:     make(chan error),
		stopChan:    make(chan bool),
		wg:          errgroup.WithContext(context.Background()),
	}
}

func (s *Subscriber) Subscribe(handler func(messages []mqHttpSdk.ConsumeMessageEntry)) {
	s.handlerFunc = handler

	s.wg.Go(func(ctx context.Context) (err error) {
		s.master()
		return
	})

	s.wg.Go(func(ctx context.Context) (err error) {
		s.worker()
		return
	})

	atomic.StoreInt32(&s.status, stateRunning)
}

func (s *Subscriber) Ack(receiptHandles []string) error {
	return s.consumer.AckMessage(receiptHandles)
}

func (s *Subscriber) Close() {
	var err error

	if atomic.LoadInt32(&s.status) == stateRunning {
		close(s.stopChan)
		err = s.wg.Wait()
	}

	if err != nil {
		log.Error("[Subscriber]topic:%s, groupId:%s close err:%v", s.config.Topic, s.config.GroupId, err)
	} else {
		log.Info("[Subscriber]topic:%s, groupId:%s close", s.config.Topic, s.config.GroupId)
	}
}

func (s *Subscriber) master() {
	defer func() {
		log.Info("[Subscriber]topic:%s, groupId:%s master close", s.config.Topic, s.config.GroupId)
	}()

	log.Info("[Subscriber]topic:%s, groupId:%s master start", s.config.Topic, s.config.GroupId)

	for {
		select {
		case <-s.stopChan:
			{
				close(s.messageChan)
				close(s.errChan)
				return
			}
		default:
			s.consumer.ConsumeMessage(s.messageChan, s.errChan,
				s.config.NumOfMessages,
				s.config.WaitSeconds,
			)
		}
	}
}

func (s *Subscriber) worker() {
	defer func() {
		log.Info("[Subscriber]topic:%s, groupId:%s worker close", s.config.Topic, s.config.GroupId)
	}()

	log.Info("[Subscriber]topic:%s, groupId:%s worker start", s.config.Topic, s.config.GroupId)

	for {
		select {
		case resp, ok := <-s.messageChan:
			{
				if !ok {
					log.Info("[Subscriber]topic:%s, groupId:%s messageChan close", s.config.Topic, s.config.GroupId)
					return
				}

				s.handlerFunc(resp.Messages)
			}
		case err, ok := <-s.errChan:
			{
				if !ok {
					log.Info("[Subscriber]topic:%s, groupId:%s errChan close", s.config.Topic, s.config.GroupId)
					return
				}
				if strings.Contains(err.(errors.ErrCode).Error(), "MessageNotExist") {
					continue
				} else {
					log.Error("[Subscriber]topic:%s, groupId:%s worker err:%+v", s.config.Topic, s.config.GroupId, err)
				}
			}
		}
	}
}

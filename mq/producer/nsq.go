package producer

import (
	"github.com/nsqio/go-nsq"
	"time"
)

type NsqProducer struct {
	address  string
	opt      Poptions
	producer *nsq.Producer
}

func NewNsqProducer(addr string, opt Poptions) *NsqProducer {
	config := nsq.NewConfig()
	if opt.AuthSecret != "" {
		config.AuthSecret = opt.AuthSecret
	}
	q, err := nsq.NewProducer(addr, config)
	if err != nil {
		return nil
	}
	producer := NsqProducer{
		address:  addr,
		opt:      opt,
		producer: q,
	}
	return &producer
}
func (nsqp NsqProducer) Close() {}
func (nsqp NsqProducer) Publish(msg PMessage, route PRoute) error {
	err := nsqp.producer.Publish(route.Topic, msg.Body)
	return err
}
func (nsqp NsqProducer) DelayPublish(delay time.Duration, msg PMessage, route PRoute) error {
	err := nsqp.producer.DeferredPublish(route.Topic, delay, msg.Body)
	return err
}

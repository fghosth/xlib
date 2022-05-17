package producer

import (
	"github.com/fghosth/xlib/mq"
	"time"
)

// MqProducer 发送消息接口
type MqProducer interface {
	// Publish 发送消息，有错误就返回,发送成功error为nil
	Publish(PMessage, PRoute) error
	// DelayPublish 延迟发送消息，有错误就返回,发送成功error为nil
	DelayPublish(time.Duration, PMessage, PRoute) error
	Close()
}

// PRoute 发送消息的路由 ,不必都填写，但是要知道路由规则，每个mq都不太一样。
type PRoute struct {
	Exchange  string                 //AMQP
	RouteKey  string                 //AMQP
	QueueName string                 //AMQP
	Args      map[string]interface{} //AMQP
	Topic     string                 //kafka topic, nsq topic
}

// PMessage 发送的消息
type PMessage struct {
	Time time.Time //kafka 指定消息发送的时间
	Key  string    //kafka消息有key，如果是其他mq没有key则不用填写
	Body []byte    //消息内容
}

// Poptions producer的option
type Poptions struct {
	Vhost            string        //AMQP
	Timeout          time.Duration //秒,0不超时
	Username         string
	Password         string
	AuthSecret       string
	AMQPQueueDeclare mq.AMQPQueueDeclareOPT
	AMQPPulish       mq.AMQPPubiishOPT
	AMQPQueueBind    mq.AMQPQueueBindOPT
	GenQueueExpires  int //amq.gen--XXXXXX 自动清除时间 ms
	PRoute           PRoute
}

// Option ProducerOption producer option
type Option func(opt *Poptions)

// WithProducerTimeout 设置超时时间
func WithProducerTimeout(num time.Duration) Option {
	return func(option *Poptions) {
		option.Timeout = num
	}
}

// WithProducerVhost 设置vhost
func WithProducerVhost(vhost string) Option {
	return func(option *Poptions) {
		option.Vhost = vhost
	}
}

// WithProducerQueueBind 设置绑定queue信息
func WithProducerQueueBind(route PRoute) Option {
	return func(option *Poptions) {
		option.PRoute = route
	}
}

// WithProducerUserAndPwd 设置用户名密码
func WithProducerUserAndPwd(user, pwd string) Option {
	return func(option *Poptions) {
		option.Username = user
		option.Password = pwd
	}
}

func NewProducer(mqType mq.MQType, addr string, opt ...Option) MqProducer {
	option := Poptions{
		Timeout: 0, //秒,0不超时
	}
	for _, fn := range opt {
		fn(&option)
	}
	var mqProducer MqProducer
	switch mqType {
	case mq.AMQP:
		queueDeclare := mq.AMQPQueueDeclareOPT{
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			Nowait:     false,
			Args:       nil,
		}
		queueBind := mq.AMQPQueueBindOPT{
			Nowait: false,
			Args:   nil,
		}
		publish := mq.AMQPPubiishOPT{
			Mandatory: false,
			Immediate: false,
		}
		option.AMQPPulish = publish
		option.AMQPQueueBind = queueBind
		option.AMQPQueueDeclare = queueDeclare
		option.GenQueueExpires = 3000
		mqProducer = NewAMQPProducer(addr, option)
	case mq.Kafka:
		mqProducer = NewKafkaProducer(addr, option)
	case mq.NSQ:
		mqProducer = NewNsqProducer(addr, option)
	}
	return mqProducer
}

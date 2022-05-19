package consumer

import (
	"time"

	"github.com/Shopify/sarama"
	mapset "github.com/deckarep/golang-set"
	"github.com/fghosth/xlib/mq"
)

// MqConsumer 消费者接口
type MqConsumer interface {
	// RegisterReceiver 注册接受者
	RegisterReceiver(Receiver)
	// Start 开始接受消息
	Start()
	// Stop 结束
	Stop()
}

// Receiver 用于接收消息到来的数据
type Receiver interface {
	OnError(err error)           // 处理遇到的错误，当MQ对象发生了错误，他需要告诉接收者处理错误
	OnReceive(msg Cmessage) bool // 处理收到的消息, 这里需要告知MQ对象消息是否处理成功
	GetTopic() Topic             //消息的topic，queue等信息
}

// Cmessage 消息格式
type Cmessage struct {
	Key  []byte
	Body []byte
	Info map[string]interface{}
}

// Topic 监听的topic
type Topic struct {
	Topic     string
	Topics    []string
	Queue     string
	RouterKey string
	Channel   string
	Args      map[string]interface{}
}

// Coptions consumer的option
type Coptions struct {
	Kafka          KafkaOpt
	Nsq            NsqOpt
	AMQP           AMQPOpt
	LookupInterval time.Duration //重连时间
	Address        string
}

type KafkaOpt struct {
	Version    *sarama.KafkaVersion
	Topic      []string
	GroupID    string               //offset
	Offsets    []OffsetGroup        //不同group，topic的offset信息
	Partition  mapset.Set           //接受的partition
	TimeFilter map[string]time.Time //时间范围，string只有『After』[Befter](不包含，开区间)
	UserName   string               //鉴权 用户名
	Password   string               //鉴权 密码
}

type OffsetGroup struct {
	Topic     string
	Partition int32
	Offset    int64
	Metadata  string
}

type NsqOpt struct {
	Topic      string
	Channel    string
	AuthSecret string //鉴权 secret
	MaxTaskNum int
}

type AMQPOpt struct {
	Exchange            string
	ExchangeType        string
	Heartbeat           time.Duration //心跳检查间隔 s
	Vhost               string
	QueueName           string
	RouterKey           string
	Username            string //鉴权 用户名
	Password            string //鉴权 密码
	AMQPQueueDeclare    mq.AMQPQueueDeclareOPT
	AMQPExchangeDeclare mq.AMQPExchangeDeclareOPT
	AMQPQueueBind       mq.AMQPQueueBindOPT
	AMQPConsumer        mq.AMQPConsumeOPT
}

// Option Consumer option
type Option func(opt *Coptions)

func NewConsumer(mqType mq.MQType, address string, opt ...Option) MqConsumer {
	option := Coptions{
		Address:        address,
		LookupInterval: 3, //默认3秒
	}
	for _, fn := range opt {
		fn(&option)
	}
	var mqConsumer MqConsumer
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
		exchangeDeclare := mq.AMQPExchangeDeclareOPT{
			NoWait:     false,
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Args:       nil,
		}
		consumers := mq.AMQPConsumeOPT{
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			Nowait:    false,
			Args:      nil,
		}
		option.AMQP.AMQPQueueBind = queueBind
		option.AMQP.AMQPQueueDeclare = queueDeclare
		option.AMQP.AMQPConsumer = consumers
		option.AMQP.AMQPExchangeDeclare = exchangeDeclare
		mqConsumer = NewAMQPConsumer(option.Address, option)
	case mq.Kafka:
		mqConsumer = NewKafkaConsumer(address, option)
	case mq.NSQ:
		mqConsumer = NewNsqConsumer(address, option)
	}
	return mqConsumer
}

// WithNsqTopic 设置nsqTopic
func WithNsqTopic(topic string) Option {
	return func(option *Coptions) {
		option.Nsq.Topic = topic
	}
}

// WithAddress 设置address
func WithAddress(addr string) Option {
	return func(option *Coptions) {
		option.Address = addr
	}
}

// WithLookupInterval 设置重连间隔
func WithLookupInterval(t int) Option {
	return func(option *Coptions) {
		option.LookupInterval = time.Duration(t)
	}
}

// WithKafka 设置kafka配置
func WithKafka(kOpt KafkaOpt) Option {
	return func(option *Coptions) {
		option.Kafka = kOpt
	}
}

// WithNsq 设置nsq配置
func WithNsq(nOpt NsqOpt) Option {
	return func(option *Coptions) {
		option.Nsq = nOpt
	}
}

// WithAMQP 设置rabbitmq配置
func WithAMQP(rOpt AMQPOpt) Option {
	return func(option *Coptions) {
		option.AMQP = rOpt
	}
}

package mq

import "github.com/streadway/amqp"

type MQType int16

var (
	AMQP          MQType = 1
	NSQ           MQType = 2
	Kafka         MQType = 3
	DefaultUser          = "guest"
	DefaultPWD           = "guest"
	RetryInterval        = 3 //断线重连间隔
)

// AMQPExchangeDeclareOPT name:交换器的名称，对应图中exchangeName。
//kind:也叫作type，表示交换器的类型。有四种常用类型：direct、fanout、topic、headers。
//durable:是否持久化，true表示是。持久化表示会把交换器的配置存盘，当RMQ Server重启后，会自动加载交换器。
//autoDelete:是否自动删除，true表示是。至少有一条绑定才可以触发自动删除，当所有绑定都与交换器解绑后，会自动删除此交换器。
//internal:是否为内部，true表示是。客户端无法直接发送msg到内部交换器，只有交换器可以发送msg到内部交换器。
//noWait:是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
//args:参数，各AMQP实现功能不同。
type AMQPExchangeDeclareOPT struct {
	Name       string
	Kind       string
	NoWait     bool
	Durable    bool
	AutoDelete bool
	Internal   bool
	Args       amqp.Table
}

// AMQPExchangeBindOPT destination：目的交换器，通常是内部交换器。
//key：对应图中BandingKey，表示要绑定的键。
//source：源交换器。
//nowait：是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
//args:参数，各AMQP实现功能不同。
type AMQPExchangeBindOPT struct {
	Destination string
	Key         string
	Source      string
	Nowait      bool
	Args        amqp.Table
}

// AMQPQueueDeclareOPT name：队列名称
//durable：是否持久化，true为是。持久化会把队列存盘，服务器重启后，不会丢失队列以及队列内的信息。（注：1、不丢失是相对的，如果宕机时有消息没来得及存盘，还是会丢失的。2、存盘影响性能。）
//autoDelete：是否自动删除，true为是。至少有一个消费者连接到队列时才可以触发。当所有消费者都断开时，队列会自动删除。
//exclusive：是否设置排他，true为是。如果设置为排他，则队列仅对首次声明他的连接可见，并在连接断开时自动删除。（注意，这里说的是连接不是信道，相同连接不同信道是可见的）。
//nowait：是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
//args：参数，各AMQP实现功能不同。
type AMQPQueueDeclareOPT struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	Nowait     bool
	Args       amqp.Table
}

// AMQPQueueBindOPT name：队列名称
//key：对应图中BandingKey，表示要绑定的键。
//exchange：交换器名称
//nowait：是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
//args：参数，各AMQP实现功能不同。
type AMQPQueueBindOPT struct {
	Name     string
	Key      string
	Exchange string
	Nowait   bool
	Args     amqp.Table
}

// AMQPPubiishOPT exchange：要发送到的交换机名称，对应图中exchangeName。
//key：路由键，对应图中RoutingKey。
//mandatory：直接false，不建议使用，后面有专门章节讲解。
//immediate ：直接false，不建议使用，后面有专门章节讲解。
type AMQPPubiishOPT struct {
	Exchange  string
	Key       string
	Mandatory bool //至少将消息route到一个队列中，否则就将消息return给发送者
	Immediate bool //至少将消息route到一个队列中，否则就将消息return给发送者 rabbitmq 3.0不支持
}

// AMQPConsumeOPT queue:队列名称。
//consumer:消费者标签，用于区分不同的消费者。
//autoAck:是否自动回复ACK，true为是，回复ACK表示高速服务器我收到消息了。建议为false，手动回复，这样可控性强。
//exclusive:设置是否排他，排他表示当前队列只能给一个消费者使用。
//noLocal:如果为true，表示生产者和消费者不能是同一个connect。
//nowait：是否非阻塞，true表示是。阻塞：表示创建交换器的请求发送后，阻塞等待RMQ Server返回信息。非阻塞：不会阻塞等待RMQ Server的返回信息，而RMQ Server也不会返回信息。（不推荐使用）
//args：参数，各AMQP实现功能不同。
type AMQPConsumeOPT struct {
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	Nowait    bool
	Args      amqp.Table
}

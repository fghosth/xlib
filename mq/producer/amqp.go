package producer

import (
	"fmt"
	"github.com/fghosth/xlib/mq"
	"github.com/streadway/amqp"
	"log"
	"strconv"
	"time"
)

type AMQP struct {
	uri      string
	opt      Poptions
	amqpconf amqp.Config
	source   *amqp.Connection
	channel  *amqp.Channel
	Online   bool //是否在线
}

// NewAMQPProducer New 创建一个新的操作RabbitMQ的对象
func NewAMQPProducer(addr string, opt Poptions) *AMQP {
	var err error
	rbmq := &AMQP{
		Online: false,
		opt:    opt,
	}
	if opt.Username != "" && opt.Password != "" { //是否用帐号密码
		rbmq.uri = fmt.Sprintf("amqp://%s:%s@%s/", opt.Username, opt.Password, addr)
	} else { //使用默认帐号
		rbmq.uri = fmt.Sprintf("amqp://%s:%s@%s/", mq.DefaultUser, mq.DefaultPWD, addr)
	}
	rbmq.amqpconf = amqp.Config{
		Heartbeat: opt.Timeout,
		Locale:    "en_US",
	}
	if opt.Vhost != "" {
		rbmq.amqpconf.Vhost = opt.Vhost
	}
	conn, err := amqp.DialConfig(rbmq.uri, rbmq.amqpconf)
	if err != nil {
		log.Println(err)
		return nil
	} else {
		rbmq.source = conn
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil
	}

	rbmq.channel = ch
	return rbmq
}

func (am *AMQP) Close() {
	err := am.channel.Close()
	if err != nil {
		log.Println("channel.Close error", err)
	}
	err = am.source.Close()
	if err != nil {
		log.Println("source.Close error", err)
	}
}

func (am *AMQP) Publish(msg PMessage, route PRoute) error {
	//发送消息
	err := am.channel.Publish(
		route.Exchange,              // exchange     交换器名称，使用默认
		route.RouteKey,              // routing key    路由键，这里为队列名称
		am.opt.AMQPPulish.Mandatory, // mandatory
		am.opt.AMQPPulish.Immediate, //至少将消息route到一个队列中，否则就将消息return给发送者;
		amqp.Publishing{
			ContentType: "text/plain", //消息类型，文本消息
			Body:        msg.Body,
		})
	return err
}
func (am AMQP) DelayPublish(delay time.Duration, msg PMessage, route PRoute) error {
	args := amqp.Table(route.Args)
	if args == nil {
		args = make(amqp.Table)
	}
	args["delay"] = int(delay.Milliseconds())   // ali amqp
	args["x-delay"] = int(delay.Milliseconds()) // rabbitmq
	//在使用rabbitMQ做消息传递了一段时间后，固定了队列传输消息时，
	//每次发送消息会生成一个amq.gen--XXXXXX的随机队列，不会自动清除
	//设置3秒后自动清除
	args["x-expires"] = am.opt.GenQueueExpires

	queueName := "delay" + strconv.FormatInt(delay.Milliseconds(), 10)
	conn, err := amqp.DialConfig(am.uri, am.amqpconf)
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println("conn.Close error", err)
		}
	}()
	if err != nil {
		log.Println(err)
		return nil
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return nil
	}
	_, err = ch.QueueDeclare(
		queueName,                         // name
		am.opt.AMQPQueueDeclare.Durable,   // durable
		true,                              // delete when usused
		am.opt.AMQPQueueDeclare.Exclusive, // exclusive(排他性队列)
		am.opt.AMQPQueueDeclare.Nowait,    // no-wait
		args,                              // arguments
		//nil,
	)
	if nil != err {
		// 当队列初始化失败的时候，需要告诉这个接收者相应的错误
		log.Println("初始化队列  失败:", route.QueueName, err.Error())
		return nil
	}
	// 将Queue绑定到Exchange上去
	err = ch.QueueBind(
		queueName,                   // queue name
		route.RouteKey,              // routing key
		route.Exchange,              // exchange
		am.opt.AMQPQueueBind.Nowait, // no-wait
		am.opt.AMQPQueueBind.Args,
	)
	if nil != err {
		log.Println("绑定队列  到Exchanges失败:", route.QueueName, route.RouteKey, err.Error())
		return nil
	}
	//发送消息
	err = ch.Publish(
		route.Exchange,              // exchange     交换器名称，使用默认
		route.RouteKey,              // routing key    路由键，这里为队列名称
		am.opt.AMQPPulish.Mandatory, // mandatory
		am.opt.AMQPPulish.Immediate, //至少将消息route到一个队列中，否则就将消息return给发送者;
		amqp.Publishing{
			ContentType: "text/plain", //消息类型，文本消息
			Body:        msg.Body,
		})
	return nil
}

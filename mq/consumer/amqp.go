package consumer

import (
	"fmt"
	"github.com/fghosth/xlib/mq"
	"github.com/fghosth/xlib/mq/utils"
	"github.com/streadway/amqp"
	"log"
	"strconv"
	"sync"
	"time"
)

type AMQPConsumer struct {
	address   string
	opt       Coptions
	receivers []Receiver
	source    *amqp.Connection
	channel   *amqp.Channel
	wg        *sync.WaitGroup
	config    amqp.Config
	uri       string //连接字符串
	stopflag  bool   //停止标签
	lock      *sync.RWMutex
}

func NewAMQPConsumer(addr string, opt Coptions) *AMQPConsumer {
	amqpconf := amqp.Config{
		Heartbeat: opt.AMQP.Heartbeat,
		Locale:    "en_US",
	}
	if opt.AMQP.Vhost != "" {
		amqpconf.Vhost = opt.AMQP.Vhost
	}
	var uri string
	if opt.AMQP.Username != "" && opt.AMQP.Password != "" { //是否用帐号密码
		uri = fmt.Sprintf("amqp://%s:%s@%s/", opt.AMQP.Username, opt.AMQP.Password, addr)
	} else { //使用默认帐号
		uri = fmt.Sprintf("amqp://%s:%s@%s/", mq.DefaultUser, mq.DefaultPWD, addr)
	}
	consumer := AMQPConsumer{
		address:   addr,
		opt:       opt,
		wg:        new(sync.WaitGroup),
		config:    amqpconf,
		uri:       uri,
		stopflag:  false,
		receivers: make([]Receiver, 0),
		lock:      new(sync.RWMutex),
	}
	conn, err := amqp.DialConfig(uri, amqpconf)
	if err != nil {
		log.Println(err)
		return nil
	} else {
		consumer.source = conn
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return nil
	} else {
		consumer.channel = ch
	}
	return &consumer
}

func (amqpc *AMQPConsumer) RegisterReceiver(r Receiver) {
	amqpc.receivers = append(amqpc.receivers, r)
}
func (amqpc *AMQPConsumer) Start() {
	for range time.Tick(time.Duration(mq.RetryInterval) * time.Second) { //任何时候断开连接秒后重连
		amqpc.run()
		//结束信号
		if amqpc.stopflag {
			return
		}
	}
}
func (amqpc *AMQPConsumer) Stop() {
	amqpc.distory()
}

//重新连接channel
func (amqpc *AMQPConsumer) connect() *amqp.Channel {
	conn, err := amqp.DialConfig(amqpc.uri, amqpc.config)
	if err != nil {
		return nil
	} else {
		amqpc.source = conn
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil
	}
	return ch
}

// Listen 监听指定路由发来的消息
// 这里需要针对每一个接收者启动一个goroutine来执行listen
// 该方法负责从每一个接收者监听的队列中获取数据，并负责重试
func (amqpc *AMQPConsumer) listen(receiver Receiver) {
	defer amqpc.wg.Done()
	// 这里获取每个接收者需要监听的队列和路由
	amqpc.lock.Lock()
	queueName := receiver.GetTopic().Queue
	routerKey := receiver.GetTopic().RouterKey
	// 申明Queue
	args := amqp.Table(receiver.GetTopic().Args)
	_, err := amqpc.channel.QueueDeclare(
		queueName,                                  // name
		amqpc.opt.AMQP.AMQPQueueDeclare.Durable,    // durable
		amqpc.opt.AMQP.AMQPQueueDeclare.AutoDelete, // delete when usused
		amqpc.opt.AMQP.AMQPQueueDeclare.Exclusive,  // exclusive(排他性队列)
		amqpc.opt.AMQP.AMQPQueueDeclare.Nowait,     // no-wait
		args,                                       // arguments
	)
	if nil != err {
		// 当队列初始化失败的时候，需要告诉这个接收者相应的错误
		receiver.OnError(fmt.Errorf("初始化队列 %s 失败: %s", queueName, err.Error()))
	}
	// 将Queue绑定到Exchange上去
	err = amqpc.channel.QueueBind(
		queueName,                           // queue name
		routerKey,                           // routing key
		amqpc.opt.AMQP.Exchange,             // exchange
		amqpc.opt.AMQP.AMQPQueueBind.Nowait, // no-wait
		amqpc.opt.AMQP.AMQPQueueBind.Args,
	)
	amqpc.lock.Unlock()
	if nil != err {
		receiver.OnError(fmt.Errorf("绑定队列 [%s - %s] 到Exchanges失败: %s", queueName, routerKey, err.Error()))
	}

	// 获取消费通道
	//prefetchSize：0 prefetchSize maximum amount of content (measured in* octets) that the server will deliver, 0 if unlimited
	//prefetchCount：会告诉RabbitMQ不要同时给一个消费者推送多于N个消息，即一旦有N个消息还没有ack，则该consumer将block掉，直到有消息ack
	//global：true\false 是否将上面设置应用于channel，简单点说，就是上面限制是channel级别的还是consumer级别
	//备注：据说prefetchSize 和global这两项，rabbitmq没有实现，暂且不研究
	err = amqpc.channel.Qos(10, 0, false) // 确保rabbitmq会一个一个发消息
	if err != nil {
		log.Println(err)
	}
	cid, err := utils.GetUUID()
	if err != nil {
		log.Println(err)
	}
	msgs, err := amqpc.channel.Consume(
		queueName,                             // queue
		strconv.FormatUint(cid, 10),           // consumer
		amqpc.opt.AMQP.AMQPConsumer.AutoAck,   // auto-ack
		amqpc.opt.AMQP.AMQPConsumer.Exclusive, // exclusive
		amqpc.opt.AMQP.AMQPConsumer.NoLocal,   // no-local
		amqpc.opt.AMQP.AMQPConsumer.Nowait,    // no-wait
		amqpc.opt.AMQP.AMQPConsumer.Args,      // args
	)
	if nil != err {
		receiver.OnError(fmt.Errorf("获取队列 %s 的消费通道失败: %s", queueName, err.Error()))
	}

	// 使用callback消费数据
	for msg := range msgs {
		bmsg := Cmessage{
			Body: msg.Body,
		}
		// 当接收者消息处理失败的时候，
		// 比如网络问题导致的数据库连接失败，连接失败等等这种
		// 通过重试可以成功的操作，那么这个时候是需要重试的
		// 直到数据处理成功后再返回，然后才会回复rabbitmq ack
		for !receiver.OnReceive(bmsg) {
			log.Println("receiver 数据处理失败，将要重试")
			time.Sleep(1 * time.Second)
		}

		// 确认收到本条消息, multiple必须为false
		err = msg.Ack(false)
		if err != nil {
			log.Println(err)
		}
	}
}

// prepareExchange 准备rabbitmq的Exchange
func (amqpc *AMQPConsumer) prepareExchange() error {

	// 申明Exchange
	err := amqpc.channel.ExchangeDeclare(
		amqpc.opt.AMQP.Exchange,                       // exchange
		amqpc.opt.AMQP.ExchangeType,                   // type
		amqpc.opt.AMQP.AMQPExchangeDeclare.Durable,    // durable
		amqpc.opt.AMQP.AMQPExchangeDeclare.AutoDelete, // autoDelete
		amqpc.opt.AMQP.AMQPExchangeDeclare.Internal,   // internal
		amqpc.opt.AMQP.AMQPExchangeDeclare.NoWait,     // noWait
		amqpc.opt.AMQP.AMQPExchangeDeclare.Args,       // args
	)

	if nil != err {

		return err
	}
	return nil
}

//检查连接是否正常
func (amqpc *AMQPConsumer) refresh() bool {
	_, err := amqp.DialConfig(amqpc.uri, amqpc.config)
	if err != nil {
		return false
	} else {
		return true
	}
}

//销毁连接对象
func (amqpc *AMQPConsumer) distory() {
	err := amqpc.channel.Close()
	if err != nil {
		log.Println("channel.Close", err)
	}
	err = amqpc.source.Close()
	if err != nil {
		log.Println("source.Close", err)
	}
}

// run 开始获取连接并初始化相关操作
func (amqpc *AMQPConsumer) run() {
	defer func() {
		amqpc.distory()
		if err := recover(); err != nil {
			log.Println("rabbitmq run() 进程  崩了", err)
		}
	}()
	if !amqpc.refresh() {
		log.Println("rabbit刷新连接失败，将要重连")
		return
	}

	// 获取新的channel对象
	amqpc.channel = amqpc.connect()
	// pp.Println(mq.channel)
	// 初始化Exchange
	err := amqpc.prepareExchange()
	if err != nil {
		log.Println(err)
	}
	for _, receiver := range amqpc.receivers {
		amqpc.wg.Add(1)
		time.Sleep(100 * time.Millisecond)
		go amqpc.listen(receiver) // 每个接收者单独启动一个goroutine用来初始化queue并接收消息
	}

	amqpc.wg.Wait()

	log.Println("所有处理queue的任务都意外退出了")
	// 理论上mq.run()在程序的执行过程中是不会结束的
	// 一旦结束就说明所有的接收者都退出了，那么意味着程序与rabbitmq的连接断开
	// 那么则需要重新连接，这里尝试销毁当前连接
}

package consumer

import (
	"github.com/fghosth/xlib/mq"
	"github.com/fghosth/xlib/mq/utils"
	"log"
	"testing"
)

func TestAliAMQP(t *testing.T) {
	ak := ""
	sk := ""
	instanceId := "" // 请替换成您阿里云AMQP控制台首页instanceId
	vhost := "xxx1"

	addr := "XXXXXX:5672"
	username := utils.GetUserName(ak, instanceId)
	password := utils.GetPassword(sk)
	opt := AMQPOpt{
		Vhost:        vhost,
		Username:     username,
		Password:     password,
		ExchangeType: "direct",
		Exchange:     "xxx_test",
	}
	c := NewConsumer(mq.AMQP, addr, WithAMQP(opt))
	c.RegisterReceiver(&AMQPC{})
	c.RegisterReceiver(&AMQPC2{})
	c.Start()
}

func TestRabbitMQ(t *testing.T) {
	addr := "localhost:5672"
	opt := AMQPOpt{
		Username:     "guest",
		Password:     "guest",
		ExchangeType: "fanout",
		Exchange:     "report.xxx",
	}
	c := NewConsumer(mq.AMQP, addr, WithAMQP(opt))
	for i := 0; i < 10; i++ {
		c.RegisterReceiver(&RabbitMQC{})
		c.RegisterReceiver(&RabbitMQC2{})

	}
	c.Start()
}

type AMQPC struct {
	//topic string
}

func (amqp AMQPC) OnReceive(msg Cmessage) bool {
	log.Println("AMQPC", msg)
	return true
}

func (amqp AMQPC) OnError(err error) {
	log.Println(err)
}

func (amqp AMQPC) GetTopic() Topic {
	args := map[string]interface{}{"x-dead-letter-exchange": "xxxDead", "x-dead-letter-routing-key": "report"}
	return Topic{Queue: "report", RouterKey: "task", Args: args}
}

type AMQPC2 struct {
	//topic string
}

func (amqp AMQPC2) OnReceive(msg Cmessage) bool {
	log.Println("AMQPC2", msg)
	return true
}

func (amqp AMQPC2) OnError(err error) {
	log.Println(err)
}

func (amqp AMQPC2) GetTopic() Topic {
	args := map[string]interface{}{"x-dead-letter-exchange": "xxxDead", "x-dead-letter-routing-key": "report"}
	return Topic{Queue: "reportstask", RouterKey: "mtask", Args: args}
}

type RabbitMQC struct {
	//topic string
}

func (ampq RabbitMQC) OnReceive(msg Cmessage) bool {
	log.Println("AMQPC", msg)
	return true
}

func (ampq RabbitMQC) OnError(err error) {
	log.Println(err)
}

func (ampq RabbitMQC) GetTopic() Topic {
	//args := map[string]interface{}{"x-dead-letter-exchange": "xxxDead", "x-dead-letter-routing-key": "report"}
	return Topic{Queue: "report2", RouterKey: "rrr", Args: nil}
}

type RabbitMQC2 struct {
	//topic string
}

func (rbmq RabbitMQC2) OnReceive(msg Cmessage) bool {
	log.Println("AMQPC2", msg)
	return true
}

func (rbmq RabbitMQC2) OnError(err error) {
	log.Println(err)
}

func (rbmq RabbitMQC2) GetTopic() Topic {
	//args:=map[string]interface{}{"x-dead-letter-exchange": "xxxDead", "x-dead-letter-routing-key": "report"}
	return Topic{Queue: "reportstask", RouterKey: "mtask", Args: nil}
}

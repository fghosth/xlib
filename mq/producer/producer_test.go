package producer

import (
	"github.com/fghosth/xlib/mq"
	"github.com/fghosth/xlib/mq/utils"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestAliQMQP(t *testing.T) {
	ak := "XXXXXXX"
	sk := "XXXXXXXXX"
	instanceId := "1551953348333975" // 请替换成您阿里云AMQP控制台首页instanceId
	vhost := "xxx1"

	addr := "XXXXXXXXXXXXXXXXX"
	username := utils.GetUserName(ak, instanceId)
	password := utils.GetPassword(sk)
	msg := PMessage{
		Body: []byte("sssss"),
	}
	route := PRoute{
		Exchange: "xxx_test",
		RouteKey: "task",
	}
	p := NewProducer(mq.AMQP, addr, WithProducerUserAndPwd(username, password), WithProducerVhost(vhost))
	var err error
	//err = p.DelayPublish(8*time.Second, msg, route)
	//if err != nil {
	//	log.Println(err)
	//}
	for i := 1; i < 3; i++ {
		//err = p.Publish(msg, route)
		err = p.DelayPublish(time.Duration(i)*time.Second, msg, route)
		if err != nil {
			log.Println(err)
		}
	}

}

func TestRabbitMQ(t *testing.T) {
	addr := "localhost:5672"
	username := "guest"
	password := "guest"
	vhost := "devops"
	msg := PMessage{
		Body: []byte("sssss"),
	}
	route := PRoute{
		Exchange: "cmdb.cfb",
		RouteKey: "mtask",
	}
	p := NewProducer(mq.AMQP, addr, WithProducerUserAndPwd(username, password), WithProducerVhost(vhost))
	err := p.Publish(msg, route)
	if err != nil {
		log.Println(err)
	}
	//route = PRoute{
	//	Exchange: "report.xxx",
	//	RouteKey: "rrr",
	//}
	//err = p.Publish(msg, route)
	//if err != nil {
	//	log.Println(err)
	//}
	//var wg sync.WaitGroup
	//for i := 0; i < 1; i++ {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		err := p.DelayPublish(time.Duration(3)*time.Second, msg, route)
	//		if err != nil {
	//			log.Println(err)
	//		}
	//	}()
	//}
	//wg.Wait()
	//time.Sleep(5 * time.Second)
}

func TestNsq(t *testing.T) {
	addr := "localhost:4150"
	p := NewProducer(mq.NSQ, addr)
	msg := PMessage{
		Body: []byte("sssss"),
	}
	route := PRoute{
		Topic: "test11111111",
	}
	err := p.Publish(msg, route)
	if err != nil {
		log.Println(err)
	}
	route = PRoute{
		Topic: "test222222222",
	}
	err = p.Publish(msg, route)
	if err != nil {
		log.Println(err)
	}
}

func TestKafka(t *testing.T) {
	//kafka-topics --create --zookeeper localhost:2181 --replication-factor 1 --partitions 1 --topic test
	//kafka-console-consumer  --bootstrap-server localhost:9092 --from-beginning --topic test
	//kafka-consumer-groups --bootstrap-server localhost:9092 --list
	addr := "localhost:9092"
	p := NewProducer(mq.Kafka, addr)
	msg := PMessage{
		Time: time.Now(),
		Key:  "234123",
		Body: []byte("111"),
	}
	route := PRoute{
		Topic: "test1",
	}
	for i := 0; i < 2; i++ {
		msg.Key = strconv.Itoa(i)
		for j := 0; j < 10; j++ {
			msg.Body = []byte("abc" + strconv.Itoa(j))
			err := p.Publish(msg, route)
			if err != nil {
				log.Println(err)
			}
		}

	}
}

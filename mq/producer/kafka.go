package producer

import (
	"github.com/Shopify/sarama"
	"log"
	"time"
)

type kafkaProducer struct {
	prodSync sarama.SyncProducer
	address  string
	opt      Poptions
}

func NewKafkaProducer(addr string, opt Poptions) *kafkaProducer {
	config := sarama.NewConfig()
	if opt.Username != "" && opt.Password != "" { //使用sasl认证
		config.Net.SASL.Enable = true
		config.Net.SASL.User = opt.Username
		config.Net.SASL.Password = opt.Password
	}
	//等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	//根据key hash的分区类型
	config.Producer.Partitioner = sarama.NewHashPartitioner
	//根据key的值取hash
	// config.Producer.Partitioner = sarama.NewHashPartitioner
	//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	config.Version = sarama.V3_1_0_0
	prodSync, e := sarama.NewSyncProducer([]string{addr}, config)
	if e != nil {
		log.Println(e)
		return nil
	}
	pd := &kafkaProducer{
		opt:      opt,
		prodSync: prodSync,
		address:  addr,
	}
	return pd
}

func (kapd *kafkaProducer) Close() {
	err := kapd.prodSync.Close()
	if err != nil {
		log.Println("prodSync.Close error", err)
	}
}

// Publish 发送消息
func (kapd kafkaProducer) Publish(msg PMessage, route PRoute) error {
	//发送的消息,主题,key
	kmsg := &sarama.ProducerMessage{
		Topic:     route.Topic,
		Key:       sarama.StringEncoder(msg.Key),
		Value:     sarama.ByteEncoder(msg.Body),
		Timestamp: msg.Time,
	}
	if msg.Time.IsZero() {
		msg.Time = time.Now()
	}
	_, _, err := kapd.prodSync.SendMessage(kmsg)
	// part, offset, err := pd.prodSync.SendMessage(msg)
	if err != nil {
		return err
		// log.Printf("send message(%s) err=%s \n", message, err)
	} else {
		// log.Printf("发送成功，partition=%d, offset=%d \n", part, offset)
		return nil
	}
}
func (kapd kafkaProducer) DelayPublish(delay time.Duration, msg PMessage, route PRoute) error {
	time.Sleep(delay)
	err := kapd.Publish(msg, route)
	if err != nil {
		return err
	}
	return nil
}

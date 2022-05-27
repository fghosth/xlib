package consumer

import (
	"github.com/Shopify/sarama"
	"github.com/fghosth/xlib/mq"
	"log"
	"testing"
)

func TestKafka(t *testing.T) {
	addr := "localhost:9092"
	opt := KafkaOpt{
		GroupID: "derek",
		//Offsets: []OffsetGroup{
		//	{
		//		Topic:     "test1",
		//		Partition: 0,
		//		Offset:    1,
		//	},
		//},
	}
	c := NewConsumer(mq.Kafka, addr, WithKafka(opt))
	c.RegisterReceiver(&Kafka{})
	c.RegisterReceiver(&Kafka2{})
	c.RegisterReceiver(&Kafka3{})
	c.Start()
}

type Kafka struct {
	//topic string
}

func (ka Kafka) OnReceive(msg Cmessage) bool {
	log.Println("Kafka", string(msg.Key), string(msg.Body), msg.Info)
	return true
}

func (ka Kafka) OnError(err error) {
	log.Println(err)
}

func (ka Kafka) GetTopic() Topic {
	return Topic{Topics: []string{"test1"}}
}

type Kafka2 struct {
	//topic string
}

func (ka Kafka2) OnReceive(msg Cmessage) bool {
	log.Println("Kafka2", string(msg.Key), string(msg.Body), msg.Info)
	return true
}

func (ka Kafka2) OnError(err error) {
	log.Println(err)
}

func (ka Kafka2) GetTopic() Topic {
	return Topic{Topics: []string{"test1"}}
}

type Kafka3 struct {
	//topic string
}

func (ka Kafka3) OnReceive(msg Cmessage) bool {
	log.Println("Kafka3", string(msg.Key), string(msg.Body), msg.Info)
	return true
}

func (ka Kafka3) OnError(err error) {
	log.Println(err)
}

func (ka Kafka3) GetTopic() Topic {
	return Topic{Topics: []string{"test1"}}
}

func TestGetTopicInfo(t *testing.T) {
	opt := Coptions{
		Kafka: KafkaOpt{
			Version:  &sarama.V3_1_0_0,
			UserName: "",
			Password: "",
		},
	}
	addr := "localhost:9092"
	include := []string{"__consumer_offsets"}
	topics, err := GetTopicInfoInclude(addr, opt, include)
	if err != nil {
		log.Println(err)
	}
	log.Println(topics)
}

func TestGetTopicInfoExclude(t *testing.T) {
	addr := "localhost:9092"
	exclude := []string{"__consumer_offsets"}
	opt := Coptions{
		Kafka: KafkaOpt{
			Version:  &sarama.V3_1_0_0,
			UserName: "",
			Password: "",
		},
	}
	topics, err := GetTopicInfoExclude(addr, opt, exclude)
	if err != nil {
		log.Println(err)
	}
	log.Println(topics)
}

func TestCreateTopic1(t *testing.T) {
	addr := "localhost:9092"
	topicName := "test"
	opt := Coptions{
		Kafka: KafkaOpt{
			Version:  &sarama.V3_1_0_0,
			UserName: "",
			Password: "",
		},
	}
	topicDetail := sarama.TopicDetail{
		NumPartitions:     2,
		ReplicationFactor: 1,
	}
	err := CreateTopic(addr, topicName, topicDetail, opt)
	if err != nil {
		log.Println(err)
	}
}

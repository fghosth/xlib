package consumer

import (
	"github.com/fghosth/xlib/mq"
	"log"
	"testing"
)

func TestNsq(t *testing.T) {
	addr := "localhost:4150"
	c := NewConsumer(mq.NSQ, addr)
	c.RegisterReceiver(&Nsqc{})
	c.RegisterReceiver(&Nsqc2{})
	c.Start()
}

type Nsqc struct {
	//topic string
}

func (nsp Nsqc) OnReceive(msg Cmessage) bool {
	log.Println("Nsqc", msg)
	return true
}

func (nsp Nsqc) OnError(err error) {
	log.Println(err)
}

func (nsp Nsqc) GetTopic() Topic {
	return Topic{Topic: "test11111111", Channel: "channel11111111"}
}

type Nsqc2 struct {
	//topic string
}

func (nsq Nsqc2) OnReceive(msg Cmessage) bool {
	log.Println("Nsqc2", msg)
	return true
}

func (nsq Nsqc2) OnError(err error) {
	log.Println(err)
}

func (nsq Nsqc2) GetTopic() Topic {
	return Topic{Topic: "test222222222", Channel: "channe2222222222"}
}

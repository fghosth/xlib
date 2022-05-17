package consumer

import (
	"errors"
	"fmt"
	"github.com/nsqio/go-nsq"
	"log"
	"runtime/debug"
	"time"
)

type NsqConsumer struct {
	address      string
	opt          Coptions
	receivers    []Receiver
	stop         chan int
	consumers    []*nsq.Consumer
	config       *nsq.Config
	MaxWorkerNum int
}

func NewNsqConsumer(addr string, opt Coptions) *NsqConsumer {
	config := nsq.NewConfig()
	if opt.Nsq.AuthSecret != "" {
		config.AuthSecret = opt.Nsq.AuthSecret
	}

	consumer := &NsqConsumer{
		address:      addr,
		opt:          opt,
		stop:         make(chan int),
		config:       config,
		receivers:    make([]Receiver, 0),
		MaxWorkerNum: opt.Nsq.MaxTaskNum,
	}
	return consumer
}

func (nsqc *NsqConsumer) RegisterReceiver(r Receiver) {
	nsqc.receivers = append(nsqc.receivers, r)
}

// Start 开始接受消息
func (nsqc *NsqConsumer) Start() {
	for range time.Tick(3 * time.Second) { //任何时候断开连接3秒后重连
		nsqc.run()
		//结束信号
		_, ok := <-nsqc.stop
		if ok {
			for _, v := range nsqc.consumers {
				v.Stop()
			}
			return
		}
	}
}

// Stop 停止接受消息
func (nsqc NsqConsumer) Stop() {
	nsqc.stop <- 1
}

func (nsqc *NsqConsumer) run() {

	if nsqc.MaxWorkerNum <= 0 {
		nsqc.MaxWorkerNum = 1
	}

	log.Println(fmt.Sprintf("Nsq MaxWorkerNum: %d", nsqc.MaxWorkerNum))
	for _, v := range nsqc.receivers { //添加监听者

		consumer, err := nsq.NewConsumer(v.GetTopic().Topic, v.GetTopic().Channel, nsqc.config)
		if err != nil {
			log.Println(err)
		}
		nsqc.consumers = append(nsqc.consumers, consumer)
		consumer.ChangeMaxInFlight(nsqc.MaxWorkerNum * 2)
		//延迟调用所以需要新建变量
		receive := v
		consumer.AddConcurrentHandlers(nsq.HandlerFunc(func(message *nsq.Message) error {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("nsq handle func panic: %v", r)
					log.Printf("stacktrace from panic: \n%s" + string(debug.Stack()))
				}
			}()
			msg := Cmessage{
				Body: message.Body,
			}
			res := receive.OnReceive(msg)
			if res {
				return nil
			} else {
				return errors.New("client return error")
			}
		}), nsqc.MaxWorkerNum)
		err = consumer.ConnectToNSQD(nsqc.address)
		if err != nil {
			log.Println("Could not connect")
		}
	}

}

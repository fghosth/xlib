package consumer

import (
	"errors"
	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	mapset "github.com/deckarep/golang-set"
	"github.com/fghosth/xlib/mq"
	"github.com/fghosth/xlib/mq/utils"
	"log"
	"sync"
	"time"
)

type KafkaConsumer struct {
	opt       Coptions
	receivers []Receiver
	stop      []chan int
	config    *cluster.Config
	address   []string
	wg        *sync.WaitGroup
}

func NewKafkaConsumer(addr string, opt Coptions) *KafkaConsumer {
	consumer := KafkaConsumer{
		receivers: make([]Receiver, 0),
		config:    cluster.NewConfig(),
		opt:       opt,
		wg:        new(sync.WaitGroup),
	}
	if opt.Kafka.UserName != "" && opt.Kafka.Password != "" { //使用sasl认证
		consumer.config.Net.SASL.Enable = true
		consumer.config.Net.SASL.User = opt.Kafka.UserName
		consumer.config.Net.SASL.Password = opt.Kafka.Password
	}
	consumer.opt.Kafka.Partition = mapset.NewSet()
	consumer.address = []string{addr}
	consumer.config.Consumer.Offsets.Initial = sarama.OffsetNewest
	consumer.config.Consumer.Offsets.CommitInterval = 1 * time.Second
	consumer.config.Group.Mode = cluster.ConsumerModePartitions
	consumer.config.Consumer.Return.Errors = true
	consumer.config.Group.Return.Notifications = true
	// config.Group.Mode = cluster.ConsumerModePartitions
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	//consumer.config.Version = sarama.V0_11_0_0
	consumer.config.Version = sarama.V3_1_0_0

	return &consumer
}

func (kac *KafkaConsumer) RegisterReceiver(r Receiver) {
	kac.receivers = append(kac.receivers, r)
}
func (kac *KafkaConsumer) Start() {
	for range time.Tick(time.Duration(mq.RetryInterval) * time.Second) { //任何时候断开连接秒后重连
		kac.run()
	}
}
func (kac *KafkaConsumer) Stop() {
	for _, v := range kac.stop {
		v <- 1
	}
}

// run 开始获取连接并初始化相关操作
func (kac *KafkaConsumer) run() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("karuka run() 进程  崩了", err)
		}
	}()
	for _, receiver := range kac.receivers {
		kac.wg.Add(1)
		go kac.listen(receiver) // 每个接收者单独启动一个goroutine用来初始化queue并接收消息
	}

	kac.wg.Wait()
	// defer cs.Close()
}

func (kac *KafkaConsumer) setOffset(offsets []OffsetGroup, groupid string, topics []string) (err error) {
	config := cluster.NewConfig()
	if kac.opt.Kafka.UserName != "" && kac.opt.Kafka.Password != "" { //使用sasl认证
		config.Net.SASL.Enable = true
		config.Net.SASL.User = kac.opt.Kafka.UserName
		config.Net.SASL.Password = kac.opt.Kafka.Password
	}
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Group.Mode = cluster.ConsumerModeMultiplex
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second
	config.Version = kac.config.Version
	c, err := cluster.NewConsumer(kac.address, groupid, topics, config)
	defer func() {
		err = c.Close()
		if err != nil {
			log.Println("consumer.Close()", err)
		}
	}()
	if err != nil {
		// log.Printf("%s: sarama.NewSyncProducer err, message=%s \n", groupid, err)
		return err
	}
	// consume errors
	go func() {
		for err := range c.Errors() {
			log.Printf("setoffset---%s:Error: %s\n", groupid, err.Error())
		}
	}()
	// consume notifications
	go func() {
		for ntf := range c.Notifications() {
			log.Printf("setoffset---%s:Rebalanced: %+v \n", groupid, ntf)
		}
	}()
	//for _, v := range offsets {
	//	c.ResetPartitionOffset(v.Topic, v.Partition, v.Offset, v.Metadata)
	//}

	_, ok := <-c.Messages()
	if ok {
		for _, v := range offsets {
			c.ResetPartitionOffset(v.Topic, v.Partition, v.Offset, v.Metadata)
		}
	} else {
		err = errors.New("setoffset error")
	}
	return
}

// Listen 监听指定topic发来的消息
// 该方法负责从每一个接收者监听的队列中获取数据
func (kac *KafkaConsumer) listen(receiver Receiver) {
	kac.stop = append(kac.stop, make(chan int))
	defer kac.wg.Done()
	// 这里获取每个接收者需要监听的topic,Partition,Offset
	topics := receiver.GetTopic().Topics
	// partition := receiver.Partition()
	// offset := receiver.Offset()
	groupid := kac.opt.Kafka.GroupID
	offsets := kac.opt.Kafka.Offsets
	partition := kac.opt.Kafka.Partition
	timeFilter := kac.opt.Kafka.TimeFilter
	if len(offsets) > 0 { //如果需要设置offset
		err := kac.setOffset(offsets, groupid, topics)
		if err != nil {
			log.Println(err)
		}
	}
	kac.config.Group.Mode = cluster.ConsumerModePartitions
	//根据消费者获取指定的主题分区的消费者,Offset这里指定为获取最新的消息.
	c, err := cluster.NewConsumer(kac.address, groupid, topics, kac.config)
	if err != nil {
		log.Printf("%s: sarama.NewSyncProducer err, message=%s \n", groupid, err)
		return
	}
	defer func() {
		err = c.Close()
		if err != nil {
			log.Println("consumer.Close()", err)
		}
	}()

	// consume errors
	go func() {
		for err := range c.Errors() {
			receiver.OnError(err)
			// log.Printf("%s:Error: %s\n", groupId, err.Error())
		}
	}()
	// consume notifications
	go func() {
		for ntf := range c.Notifications() {
			log.Printf("%s:Rebalanced: %+v \n", groupid, ntf)
		}
	}()

	//循环等待接受消息.
	for {
		select {
		//接收消息通道和错误通道的内容.
		case part, ok := <-c.Partitions():
			if ok {
				// pp.Println(part.Partition(), partition.Cardinality())
				//start a separate goroutine to consume messages
				if !partition.Contains(part.Partition()) && partition.Cardinality() > 0 {
					continue
				}
				go func(pc cluster.PartitionConsumer) {
					for msg := range pc.Messages() {
						//过滤时间条件
						if !kac.timeFilter(timeFilter, msg.Timestamp) && len(timeFilter) > 0 { //如果设置了timefilter，并且过滤失败不返回此消息
							continue
						}
						bmsg := Cmessage{
							Key:  msg.Key,
							Body: msg.Value,
							Info: map[string]interface{}{"partition": msg.Partition},
						}
						res := receiver.OnReceive(bmsg)
						if res {
							c.MarkOffset(msg, "") // mark message as processed
						}
						// fmt.Fprintf(os.Stdout, "%s/%d/%d\t%s\t%s\n", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)

					}
				}(part)
			}
		case _, ok := <-kac.stop[len(kac.stop)-1]:
			if ok {
				err = c.Close()
				if err != nil {
					log.Println("consumer.Close()", err)
				}
				return
			}
		}
	}
}

//判断t是否tf条件范围内
func (kac *KafkaConsumer) timeFilter(tf map[string]time.Time, t time.Time) (res bool) {
	res = true
	for k, v := range tf {
		switch k {
		case "After":
			if !t.After(v) {
				res = false
				return
			}
		case "Before":
			if !t.Before(v) {
				res = false
				return
			}
		}
	}
	return
}

// GetTopicInfo
// @Description: 获得topic 信息，包含partitions信息
// @param addr
// @param version
// @param include 过滤topic 包含这些的topic
// @return topicsInfo
// @return err
func GetTopicInfo(addr string, version sarama.KafkaVersion, include []string) (topicsInfo map[string][]int32, err error) {
	topicsInfo = make(map[string][]int32)
	conf := sarama.NewConfig()
	conf.Version = version
	c, err := sarama.NewConsumer([]string{addr}, conf)
	if err != nil {
		return
	}
	topicArr, err := c.Topics()
	for _, v := range topicArr {
		if !utils.IsContainStrArr(include, v) {
			continue
		}
		var p []int32
		p, err = c.Partitions(v)
		if err != nil {
			return
		}
		topicsInfo[v] = p
	}
	return
}

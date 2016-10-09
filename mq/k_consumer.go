package mq

import (
	kmq "github.com/Shopify/sarama"
	bsm "github.com/bsm/sarama-cluster"
	"github.com/mous77/go/logger"
	"log"
	"strings"
	"time"
	"fmt"
)

type KMessage *kmq.ConsumerMessage

func toString(this KMessage) string {
	return fmt.Sprintf("topic:%s,part:%d,off:%d, key:[%s],val:[%s]",
		this.Topic, this.Partition, this.Offset,
		string(this.Key), string(this.Value))
}

type IKmqConsumer interface {
	IMQWorker
	Setup(_zks, _app, _group string, _receiver chan KMessage)
	Subscribe(_topic string)
}

/*
消费者对象接口
 */
type tKConsumer struct {
	tMQWorker
	topics    []string
	app       string
	group     string
	totalRecv int64
	receiver  chan KMessage
}

func NewKmqConsumer() IKmqConsumer {
	c := &tKConsumer{}
	c.init("kconsuemer", nil, c.tryOpen, nil)
	return c
}

func (this *tKConsumer)tryOpen() {
	this.lg.Info("doOpen...")
	this.setStatus(wsConning)
	brokers, e1 := GetKBrokers(this.zkAddrs)
	if nil != e1 {
		this.lg.Error("error on get broker %s", e1.Error())
		this.setStatus(wsClosed)
		return
	}
	this.lg.Info("alive brokers=%s", brokers)

	consumer, e2 := bsm.NewConsumer(brokers, this.group, this.topics, nil)
	if nil != e2 {
		this.lg.Error("error on new consumer %s", e2.Error())
		this.setStatus(wsClosed)
	} else {
		this.lg.Info("connect to broker %s ok!", brokers)
		this.setStatus(wsConneted)

		go func() {
			msgs := consumer.Messages()
			for msg := range msgs {
				this.receiver <- msg
			}
		}()

		go func(){
			errs := consumer.Errors()
			for err := range errs {
				this.lg.Error("error on %s", err.Error())
			}
		}()

		go func(){
			evts := consumer.Notifications()
			for evt := range evts {
				this.lg.Info("on event %s", evt.Claimed)
			}
		}()

	}
}

func (this *tKConsumer)Setup(_zks, _app, _group string, _receiver chan KMessage) {
	if nil == _receiver {
		panic("nil receiver")
	}

	this.tMQWorker.Setup(_zks)
	this.app = _app
	this.group = _group
	this.receiver = _receiver

	this.lg.Info("Setup(%s,%sm%s)", _zks, _app, _group)
}

/*
订阅特定主题的特定分区
 */
func (this *tKConsumer)Subscribe(_topic string) {
	this.lg.Info("Subscribe(%s)", _topic)
	_topic = strings.Trim(_topic, " ")
	if len(_topic) > 0 {
		for i := 0; i < len(this.topics); i++ {
			if strings.Compare(this.topics[i], _topic) == 0 {
				return
			}
		}

		this.topics = append(this.topics, _topic)
	}
}

func TestKConsumer() {
	lg := logger.GetLogger("recv")
	msgs := make(chan KMessage, 8192)
	all := 0
	go func() {
		for msg := range msgs {
			lg.Info("recv msg %s", toString(msg))
			all += 1
		}
	}()

	go func() {
		vv := 0
		for {
			lg.Info("spd=%d, total=%d", all - vv, all)
			vv = all
			time.Sleep(time.Second)
			break
		}
	}()

	c := NewKmqConsumer()
	c.Setup("127.0.0.1:2182", "demo", "g0", msgs)
	c.Subscribe("RAW-U")
	c.Start()

	log.Println("end of TestConsumer")
}
package mq

import (
	"github.com/mous77/go/logger"
	"time"
	"strconv"
	kmq "github.com/Shopify/sarama"
)

type IKProducer interface {
	IMQWorker
	Setup(_zk_addrs string)
	Push(_topic *string, _part int, _key, _val []byte)
}

/*
生产者对象接口
 */
type tKProducer struct {
	tMQWorker
	sender    kmq.AsyncProducer
	messages  chan *kmq.ProducerMessage
	totalSend int64
	totalSent int64
}

func NewKProducer() (IKProducer) {
	p := &tKProducer{}
	p.init("kproducer", p.doStart, p.tryOpen, p.doStop)
	return p
}

func (this *tKProducer)tryOpen() {
	this.lg.Info("tryOpen...")
	this.setStatus(wsConning)

	brokers, e1 := GetKBrokers(this.zkAddrs)
	if nil != e1 {
		this.lg.Error("GetBrokers(%s) failed :%s", this.zkAddrs, e1.Error())
		this.setStatus(wsClosed)
		return
	}

	config := kmq.NewConfig()
	config.Producer.RequiredAcks = kmq.WaitForLocal
	config.Producer.Partitioner = kmq.NewManualPartitioner
	config.Producer.Compression = kmq.CompressionSnappy

	if sender, err := kmq.NewAsyncProducer(brokers, config); nil != err {
		this.lg.Error("failed on new producer on %s : %s", brokers, err.Error())
		this.setStatus(wsClosed)
	} else {
		this.setStatus(wsConneted)
		this.lg.Info("connected %s", brokers)
		this.setAsyncSender(sender)
	}
}

func (this *tKProducer)setAsyncSender(sender kmq.AsyncProducer){
	this.sender = sender

	in := sender.Input()
	ou := sender.Successes()
	er := sender.Errors()

	go func() {
		for msg := range this.messages {
			this.totalSend += 1
			in <- msg
			this.totalSent += 1
		}
	}()

	go func() {
		for range ou {
		}
	}()

	go func(){
		for e := range er {
			this.lg.Error("error on push %s", e.Err.Error())
		}
	}()
}

func (this *tKProducer)doStart() {
	this.lg.Info("doStart")
	this.messages = make(chan *kmq.ProducerMessage, 1024)
}

func (this *tKProducer)doStop() {
	this.lg.Info("doStop")
	if nil != this.sender{
		this.sender.Close()
		this.sender = nil
	}
	close(this.messages)
}

func (this *tKProducer)Push(_topic *string, _part int, _key, _val []byte) {
	//this.lg.Debug("push %s %d %s=%s", *_topic, _part, string(_key), string(_val))

	if !this.isActive() {
		this.lg.Warn("push on inactive !")
	} else {
		msg := &kmq.ProducerMessage{
			Topic:*_topic,
			Partition:int32(_part),
			Key:kmq.ByteEncoder(_key),
			Value:kmq.ByteEncoder(_val)}

		this.messages <- msg
	}
}

func TestKProducer() {
	lg := logger.GetLogger("send")
	p := NewKProducer().(*tKProducer)
	p.Setup("127.0.0.1:2182")
	p.Start()
	var total int64
	go func() {
		sz := 32

		topic := "RAW-U"
		val := make([]byte, sz)[0:0]

		for i := 0; i < sz; i++ {
			val = append(val, byte('A'))
		}

		seq := 0
		for {
			ss := strconv.Itoa(seq)
			key := "key."+ss; seq+=1
			lg.Info("push %s %s", topic, key)

			p.Push(&topic, seq % 4, []byte(key), []byte(string(val)+ss))
			//time.Sleep(time.Second)
			total++
		}
	}()

	go func() {
		var last int64
		for {
			lg.Info("speed=%d total=%d, qlen=%d, send=%d, sent=%d",
				p.totalSent - last,
				total, total - p.totalSend,
				p.totalSend, p.totalSent)

			last = p.totalSent

			time.Sleep(time.Second)
			break
		}
	}()

}
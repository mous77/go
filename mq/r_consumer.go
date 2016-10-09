package mq

import (
	rmq "github.com/didapinchegit/go_rocket_mq"
)

type RMessage *rmq.MessageExt

type IRmqConsumer interface {
	IMQWorker
	Setup(_name_svr, _app, _group string, _receiver chan RMessage)
	Subscribe(_topic string, _need bool)
}

type tRmqConsumer struct {
	tMQWorker
	consumer rmq.Consumer
	receiver chan RMessage
}

func NewRmqConsumer() (IRmqConsumer) {
	c := &tRmqConsumer{}
	c.init("rconsumer", c.doStart, nil, c.doStop)
	return c
}

func (this *tRmqConsumer)Setup( _name_svr, _app, _group string, _receiver chan RMessage) {
	this.tMQWorker.Setup( _name_svr)
	config := rmq.Config{Nameserver:_name_svr, InstanceName: _app}
	var err error
	if this.consumer, err = rmq.NewDefaultConsumer(_group, &config); nil != err {
		panic("error on setup %s" + err.Error())
	}

	this.receiver = _receiver
	this.consumer.RegisterMessageListener(this.onMessages)
	this.setStatus(wsClosed)
}

func (this *tRmqConsumer)Subscribe(_topic string, _need bool) {
	if _need {
		this.consumer.Subscribe(_topic, "*")
	} else {
		this.consumer.UnSubcribe(_topic)
	}
}

func (this *tRmqConsumer)onMessages(_msgs []*rmq.MessageExt) (error) {
	for _, msg := range _msgs {
		//this.lg.Debug("recv rmsg %d %s", i, string(msg.Body))
		this.receiver <- RMessage(msg)
	}
	return nil
}

func (this *tRmqConsumer)doStart() {
	this.setStatus(wsConning)
	this.consumer.Start()
	this.setStatus(wsConneted)
}

func (this *tRmqConsumer)doStop() {
	this.consumer.Shutdown()
	this.setStatus(wsClosed)
}

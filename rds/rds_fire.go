package rds

import (
	"fmt"
	"lib/utils"
	"sync/atomic"
	"lib/log"
)

type RdsFire struct {
	lg     log.ILogger
	active int32
	name   string
	chData chan *RdsData
	onData OnRdsData

	killSig *RdsData
}

func NewRdsFire(_name string, _on_data OnRdsData) (*RdsFire) {
	name := fmt.Sprintf("rdsFire.%s", _name)

	return &RdsFire{
		active: 0,
		name:   name,
		lg:     log.GetLogger(name),
		onData: _on_data,

		killSig: &RdsData{},
	}
}

func (this *RdsFire) doFire(_data *RdsData) {
	defer func() {
		if v := recover(); nil != v {
			this.lg.Error("error on doFire(%s) %v", _data.String(), v)
		}
	}()
}

func (this *RdsFire) loopFire(_arg interface{}) {
	defer close(this.chData)

	for data := range this.chData {
		if data == this.killSig {
			break
		} else {
			this.doFire(data)
		}
	}
}

func (this *RdsFire) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.chData = make(chan *RdsData, 1024)
		utils.GoRunAdd(this.loopFire, this.name)
	}
}

func (this *RdsFire) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		this.chData <- this.killSig
	}
}

func (this *RdsFire) Offer(_topic string, _body string) {
	this.chData <- &RdsData{_topic, _body}
}

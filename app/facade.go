package app

import (
	"sync/atomic"
	"time"
	"os"
	"os/signal"
	"syscall"
	"lib/utils"
)

type AFacade struct {
	conf    *AConfig
	active  int32
	needTip bool

	fOnTip   func()
	fOnStart func()
	fOnStop  func()
}

func NewAFacade(_conf *AConfig) (*AFacade) {
	return &AFacade{conf: _conf}
}

func (this *AFacade) Setup(_on_tip func(), _on_start func(), _on_stop func()) {
	this.fOnTip = _on_tip
	this.fOnStart = _on_start
	this.fOnStop = _on_stop
}

func (this *AFacade) IsActive() (bool) {
	return atomic.LoadInt32(&this.active) > 0
}

func (this *AFacade) tipAlive()  {
	if nil != this.fOnTip {
		this.fOnTip()
	}
}

func (this *AFacade) waitForExit() {
	exitSignals := make(chan os.Signal)
	defer close(exitSignals)

	signal.Notify(exitSignals,
		syscall.SIGINT,  // 2
		syscall.SIGTERM, // 15
		syscall.SIGQUIT, // 3
		syscall.SIGKILL, // 9
	)

	<-exitSignals
	this.Stop()
}

func (this *AFacade) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.needTip = true

		itv := this.conf.GetInt("app/alive", 10)
		utils.GoSchedule(this.tipAlive, time.Second*time.Duration(itv), "app.tipAlive")

		if nil != this.fOnStart {
			this.fOnStart()
		}

		this.waitForExit()
	}
}

func (this *AFacade) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		if nil != this.fOnStop {
			this.fOnStop()
		}
		this.needTip = false

		utils.GoRunTerm()
	}
}

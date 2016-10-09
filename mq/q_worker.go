package mq

import (
	"sync/atomic"
	"time"
	"strings"
	"github.com/mous77/go/logger"
)

type IMQWorker interface {
	Start()
	Stop()
}

type tMQWorker struct {
	name    string
	active  int32
	status  int32
	zkAddrs []string
	lg      logger.ILogger
	closing chan bool
	onStart func()
	onOpen  func()
	onStop  func()
}

type wsStatusType int32

const (
	wsClosed wsStatusType = iota
	wsConning
	wsConneted
)

func (this *tMQWorker)init(_name string, _on_start, _on_open, _on_stop func()) {
	this.name = _name
	this.lg = logger.GetLogger(_name)

	this.setStatus(wsClosed)
	this.onStart = _on_start
	this.onOpen = _on_open
	this.onStop = _on_stop
}

func (this *tMQWorker)getStatus() (wsStatusType) {
	return wsStatusType(atomic.LoadInt32(&this.status))
}

func (this *tMQWorker)setStatus(_value wsStatusType) {
	atomic.StoreInt32(&this.status, int32(_value))
}

func (this *tMQWorker)Setup(_zk_addrs string) {
	this.zkAddrs = strings.Split(_zk_addrs, ",")
}

func (this *tMQWorker)Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.lg.Info("start")
		this.closing = make(chan bool)

		if nil != this.onStart {
			this.lg.Info("onStart")
			this.onStart()
		}

		if nil != this.onOpen {
			this.lg.Debug("check open loop")
			go func() {
				for this.isActive() {
					switch this.getStatus() {
					case wsClosed:
						if nil != this.onOpen {
							this.onOpen()
						}
					case wsConning:
						this.lg.Info("connecting to %s ...", this.zkAddrs)
					case wsConneted:
						//this.lg.Debug("connected[%s] ok, total pkgs: %d", this.name, this.allPkgs)
					}
					time.Sleep(time.Second * 5)
				}
			}()
		}

	}
}

func (this *tMQWorker)Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		this.lg.Info("stop")
		close(this.closing)
		if nil != this.onStop {
			this.onStop()
		}
	}
}

func (this *tMQWorker)isActive() (bool) {
	return atomic.LoadInt32(&this.active) > 0
}
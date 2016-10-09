package net

import (
	"fmt"
	"net"
	"time"
	"sync/atomic"
	"github.com/mous77/go/logger"
)

type INetClient interface {
	Type() (TunnelType)
	GetLogger()(logger.ILogger)
	GetAddr() (string)
	SetAttr(_attr interface{})
	GetAttr() (interface{})
	Write(_data []byte)
	Setup(_rhost string, _rport int)
	SetCallback(_callback *TClientCallback)
	IsConnected() (bool)
	IsActive() (bool)
	Start()
	Stop()
}

type tNetClient struct {
	log      logger.ILogger
	proto    string
	tunnel   TunnelType
	conn     net.Conn
	linked   int32
	attr     interface{}
	active   int32
	rHost    string
	rPort    int
	callback *TClientCallback
	chSend   chan []byte
	chRecv   chan []byte
}

func NewTcpClient() (INetClient) {
	return makeNetClient(TUN_TCP)
}

func NewUdpClient() (INetClient) {
	return makeNetClient(TUN_UDP)
}

func makeNetClient(_tunnel TunnelType) (*tNetClient) {
	var proto string
	if _tunnel == TUN_TCP {
		proto = "tcp"
	} else {
		proto = "udp"
	}

	return &tNetClient{tunnel:_tunnel,
		proto:proto,
		log:logger.GetLogger(proto + "-client"),
		callback:newDefCallbackC(),
		active:0}
}

func (this *tNetClient)SetCallback(_callback *TClientCallback) {
	if nil != _callback {
		this.callback.copyFrom(_callback)
	}
}

func (this *tNetClient)Type() (TunnelType) {
	return this.tunnel
}

func (this *tNetClient)IsConnected() (bool) {
	return 1 == atomic.LoadInt32(&this.linked)
}

func (this *tNetClient)IsActive() (bool) {
	return atomic.LoadInt32(&this.active) == 1
}

func (this *tNetClient)GetLogger() (logger.ILogger) {
	return this.log
}

func (this *tNetClient)GetAttr() (interface{}) {
	return this.attr
}

func (this *tNetClient)SetAttr(_attr interface{}) {
	this.attr = _attr
}

func (this *tNetClient)GetAddr() (string) {
	return fmt.Sprintf("%s:%d", this.rHost, this.rPort)
}

func (this *tNetClient)Setup(_rhost string, _rport int) {
	this.log.Info("setup(%s:%d", _rhost, _rport)
	if atomic.LoadInt32(&this.active) == 0 {
		this.rHost = _rhost
		this.rPort = _rport
	}
}

func (this *tNetClient)Start() {
	this.log.Info("Start()")
	if len(this.rHost) < 7 || this.rPort < 1 {
		panic("no remote info!")
	}

	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.chRecv = make(chan []byte, 32)
		go this.loopParse()
		go this.checkConn()
	}
}

func (this *tNetClient)Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		close(this.chRecv)
		this.close()
	}
}

func (this *tNetClient)Write(_data []byte) {
	if atomic.LoadInt32(&this.linked) == 1 {
		this.chSend <- _data
	} else {
		this.log.Warn("not linked")
	}
}

func (this *tNetClient)tryConn() {
	raddr := fmt.Sprintf("%s:%d", this.rHost, this.rPort)
	this.log.Info("try connect to %s", raddr)
	if conn, err := net.Dial(this.proto, raddr); nil != err {
		this.log.Error("error %s", err.Error())
	} else {
		atomic.StoreInt32(&this.linked, 1)
		this.log.Info("connect to %s OK!", raddr)
		this.callback.OnConn(this)
		this.conn = conn
		this.chSend = make(chan []byte, 32)
		go this.loopRead()
		go this.loopSend()
	}
}

func (this *tNetClient)checkConn() {
	tm := time.NewTimer(time.Second)
	defer func() {
		tm.Stop()
	}()
	for this.IsActive() {
		select {
		case <-tm.C:
			if this.IsActive() {
				if atomic.LoadInt32(&this.linked) == 0 {
					this.tryConn()
				}

				if atomic.LoadInt32(&this.linked) == 1 {
					//this.log.Debug("linked...")
					tm.Reset(time.Second * 3)
				} else {
					tm.Reset(time.Second)
				}
			}
		}
	}
}

func (this *tNetClient)close() {
	if atomic.CompareAndSwapInt32(&this.linked, 1, 0) {
		this.callback.OnBrok(this)
		this.conn.Close()
		close(this.chSend)
	}
}

func (this *tNetClient)loopParse() {
	for atomic.LoadInt32(&this.active) == 1 {
		for buf := range this.chRecv {
			this.callback.OnData(this, buf)
		}
	}
}

func (this *tNetClient)loopRead() {
	buf := make([]byte, 8192)
	for atomic.LoadInt32(&this.linked) > 0 {
		if nr, er := this.conn.Read(buf[0:]); nil != er {
			this.log.Error("err on read %s", er.Error())
			this.close()
			break
		} else if nr > 0 {
			data := make([]byte, nr)
			copy(data, buf[0:nr])
			this.chRecv <- data
		} else {
			this.log.Debug("read zero")
			break
		}
	}
}

func (this *tNetClient)loopSend() {
	for atomic.LoadInt32(&this.linked) > 0 {
		for buf := range this.chSend {
			for len(buf) > 0 {
				if nw, er := this.conn.Write(buf); nil != er {
					this.log.Error("error on write %s", er.Error())
					this.close()
					break
				} else {
					buf = buf[nw:]
				}
			}
		}
	}
}

func TestTcpClient() {
	logger.Setup(&logger.TConfig{Source:true, Console:true})
	c := makeNetClient(TUN_TCP)
	c.Setup("127.0.0.1", 6000)
	c.Start()

	i := 0
	for {
		msg := fmt.Sprintf("hello-%d", i); i += 1
		c.Write([]byte(msg))
		time.Sleep(time.Second * 3)
	}
}
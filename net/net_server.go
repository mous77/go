package net

import (
	"sync"
	"sync/atomic"
	"github.com/mous77/go/logger"
	"github.com/mous77/go/utils"
)

type INetServer interface {
	Setup(_host string, _ports []int, _option *TOption)
	Start()
	Stop()
	IsActive() (bool)
	GetTunnelType() (TunnelType)
	SetCallback(_callback *TServerCallback)
	GetSession(*TNetAddr) (INetSession)
	GetSessionCount() (int32)
	GetAllRecvBytes() (int64)
	GetAllSentBytes() (int64)
	WriteTo(_dst *TNetAddr, _msg []byte) (bool)
	GetLogger() (logger.ILogger)
}

type tNetServer struct {
	active       int32
	timeout      int64
	bufPool      *utils.TBufPool
	waitGroup    *sync.WaitGroup
	servLog      logger.ILogger
	lsnrLog      logger.ILogger
	sessLog      logger.ILogger
	sessionSeq   int64
	sessionCount int32
	allRecvBytes int64
	allSentBytes int64
	tunnelType   TunnelType
	host         string
	option       *TOption
	onPorts      func([]int)
	onStart      func()
	onStop       func()
	callback     *TServerCallback
}

func (this *tNetServer)GetTunnelType() (TunnelType) {
	return this.tunnelType
}

func (this *tNetServer)init(_type TunnelType, _do_ports func([]int), _do_start func(), _do_stop func()) {
	this.tunnelType = _type

	name := string(_type)

	this.servLog = logger.GetLogger(name + ".serv")
	this.lsnrLog = logger.GetLogger(name + ".lsnr")
	this.sessLog = logger.GetLogger(name + ".sess")

	this.waitGroup = &sync.WaitGroup{}
	this.callback = &TServerCallback{OnBrok:defOnBrokS, OnConn:defOnConnS, OnData:defOnDataS}

	this.onPorts = _do_ports
	this.onStart = _do_start
	this.onStop = _do_stop
}

func (this *tNetServer)GetLogger() (logger.ILogger) {
	return this.servLog
}

func (this *tNetServer)SetCallback(_callback *TServerCallback) {
	this.callback.copyFrom(_callback)
}

func (this *tNetServer)GetSessionCount() (int32) {
	return atomic.LoadInt32(&this.sessionCount)
}

func (this *tNetServer)GetAllSentBytes() (int64) {
	return atomic.LoadInt64(&this.allSentBytes)
}

func (this *tNetServer)GetAllRecvBytes() (int64) {
	return atomic.LoadInt64(&this.allRecvBytes)
}

func (this *tNetServer)IsActive() (bool) {
	return atomic.LoadInt32(&this.active) == 1
}

func (this *tNetServer)onConn(_session INetSession) {
	atomic.AddInt32(&this.sessionCount, 1)
	this.callback.OnConn(_session)
}

func (this *tNetServer)onBrok(_session INetSession) {
	this.callback.OnBrok(_session)
	atomic.AddInt32(&this.sessionCount, -1)
}

func (this *tNetServer)onData(_session INetSession, _data []byte) {
	this.callback.OnData(_session, _data)
}

func (this *tNetServer)addRecvBytes(_bytes int) {
	atomic.AddInt64(&this.allRecvBytes, int64(_bytes))
}

func (this *tNetServer)addSentBytes(_bytes int) {
	atomic.AddInt64(&this.allSentBytes, int64(_bytes))
}

func (this *tNetServer)genSessionSeq() (int64) {
	return atomic.AddInt64(&this.sessionSeq, 1) & 0xFFFF
}

func (this *tNetServer)Setup(_host string, _ports []int, _option *TOption) {
	if this.IsActive() {
		panic("can not setup on active status")
	} else {
		this.host = _host

		this.option = NewOption()
		this.option.copyFrom(_option)

		this.bufPool = this.option.GetBufPool()

		this.onPorts(_ports)
	}
}

func (this *tNetServer)Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.onStart()
	} else {
		this.servLog.Warn("start already")
	}
}

func (this *tNetServer)Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		this.onStop()
	} else {
		this.servLog.Warn("no active")
	}
}

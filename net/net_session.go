package net

import (
	"sync/atomic"
	"github.com/mous77/go/logger"
)

type INetSession interface {
	GetLog()(logger.ILogger)
	GetAddr() *TNetAddr
	GetAttr() interface{}
	SetAttr(interface{})
	Write([]byte)
	String()(string)
}

type tNetSession struct {
	server   *tNetServer
	key      *TNetAddr
	attr     interface{}

	closed   int32
	lastRecv int64
	lastSent int64
}

const NIL = ""

func (this *tNetSession)setup(_server *tNetServer, _key *TNetAddr) {
	atomic.StoreInt32(&this.closed, 0)
	this.server = _server
	this.key = _key
}

func (this *tNetSession)clean() {
	this.key.Clear()
	this.lastRecv = 0
	this.lastSent = 0
	this.attr = nil

}

func (this *tNetSession)GetLog() (logger.ILogger) {
	return this.server.sessLog
}

func (this *tNetSession)String() (string) {
	return this.key.String()
}

func (this *tNetSession)GetAttr() (interface{}) {
	return this.attr
}

func (this *tNetSession)SetAttr(_attr interface{}) {
	this.attr = _attr
}

func (this *tNetSession)GetAddr() (*TNetAddr) {
	return this.key
}

func (this *tNetSession)GetLastRecv() (int64) {
	return this.lastRecv
}

func (this *tNetSession)GetLastSent() (int64) {
	return this.lastSent
}
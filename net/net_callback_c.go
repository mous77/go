package net

import (
	"log"
	"sync/atomic"
)

type TClientCallback struct {
	OnConn func(INetClient)
	OnBrok func(INetClient)
	OnData func(INetClient, []byte)
}

func newDefCallbackC()(*TClientCallback){
	return &TClientCallback{OnConn:defOnConnC, OnBrok:defOnBrokC, OnData:defOnDataC}
}

func (this *TClientCallback)copyFrom(_src *TClientCallback){
	if nil!=_src {
		if nil != _src.OnConn {
			this.OnConn = _src.OnConn
		}
		if nil != _src.OnBrok {
			this.OnBrok = _src.OnBrok
		}
		if nil != _src.OnData {
			this.OnData = _src.OnData
		}
	}
}

var defLinkCountC int64

func defOnConnC(_sender INetClient) {
	log.Printf("====defOnConnC %s %d \r\n", _sender.GetAddr(), atomic.AddInt64(&defLinkCountC, 1))
}

func defOnBrokC(_sender INetClient) {
	log.Printf("====defOnBrokC %s\r\n", _sender.GetAddr())
}

func defOnDataC(_sender INetClient, _data []byte) {
	log.Printf("====defOnDataC %s, %s \r\n", _sender.GetAddr(), string(_data))
	_sender.Write(_data)
}
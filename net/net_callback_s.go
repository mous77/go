package net

import (
	"log"
	"sync/atomic"
)

type TServerCallback struct {
	OnConn func(INetSession)
	OnBrok func(INetSession)
	OnData func(INetSession, []byte)
}

func NewServerCallback(_on_conn, _on_brok func(INetSession), _on_data func(INetSession, []byte))(*TServerCallback) {
	return &TServerCallback{OnConn:_on_conn, OnBrok:_on_brok, OnData: _on_data}
}

func (this *TServerCallback)copyFrom(_src *TServerCallback) {
	if nil != _src {
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

var defLinkCountS int64

func defOnConnS(_session INetSession) {
	log.Printf("====defOnConn %s %d \r\n", _session.GetAddr(), atomic.AddInt64(&defLinkCountS, 1))
}

func defOnBrokS(_session INetSession) {
	log.Printf("====defOnBrok %s\r\n", _session.GetAddr())
}

func defOnDataS(_session INetSession, _data []byte) {
	//log.Printf("====defOnData %s, %s \r\n", _session.GetKey(), string(_data))
	_session.Write(_data)
}
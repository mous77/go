package net

import (
	"github.com/mous77/go/logger"
)

type tUdpServer struct {
	tNetServer
	servants map[int]*tUdpServant
}

func NewUdpServer() (INetServer) {
	server := &tUdpServer{servants:make(map[int]*tUdpServant)}
	server.init(TUN_UDP, server.doPorts, server.doStart, server.doStop)
	return server
}

func (this *tUdpServer)WriteTo(_addr *TNetAddr, _buf []byte)(bool){
	session := this.GetSession(_addr)
	if nil!=session {
		session.Write(_buf)
	}
	return true
}

func (this *tUdpServer)GetSession(_addr *TNetAddr) (INetSession) {
	panic("not support")
}

func (this *tUdpServer)doPorts(_ports []int) {
	for _, port := range _ports {
		this.servants[port] = newUdpServant(this, port)
	}
}

func (this *tUdpServer)doStop() {
	for _, servant := range this.servants {
		servant.doStop()
	}
}

func (this *tUdpServer)doStart() {
	for _, servant := range this.servants {
		servant.doStart()
	}
}

func TestUdpServer() {
	logger.Setup(&logger.TConfig{Source:true, Console:true})
	s := NewUdpServer()
	s.Setup("127.0.0.1", []int{7878}, &TOption{BufSize:1024})
	s.Start()

	genUdpClient(32)

	<-make(chan bool)
}

func genUdpClient(_cnt int) {
	for i := 0; i < _cnt; i++ {

	}
}
package net

import (
	"github.com/mous77/go/logger"
	"time"
)

type TTcpServer struct {
	tNetServer
	servants map[int]*tTcpServant
}

func NewTcpServer() (INetServer) {
	server := &TTcpServer{servants:make(map[int]*tTcpServant)}
	server.init(TUN_TCP, server.doPorts, server.doStart, server.doStop)
	return server
}

func (this *TTcpServer)WriteTo(_addr *TNetAddr, _buf []byte)(bool){
	session := this.GetSession(_addr)
	if nil!=session{
		session.Write(_buf)
	}
	return nil!=session
}

func (this *TTcpServer)GetSession(_addr *TNetAddr) (INetSession) {
	servant := this.servants[_addr.GetLocalPort()]
	if nil == servant {
		return nil
	} else {
		key := _addr.String()
		return servant.getSession(&key)
	}
}

func (this *TTcpServer)doPorts(_ports []int){
	for _, port := range _ports {
		this.servants[port] = newTcpServant(this, port)
	}
}

func (this *TTcpServer)doStop() {
	for _, servant := range this.servants {
		servant.doStop()
	}
}

func (this *TTcpServer)doStart() {
	this.servLog.Info("doStart")
	for _, servant := range this.servants {
		servant.doStart()
	}
}

func TestTcpServer() {
	logger.Setup(&logger.TConfig{App:"demo", Root:"c:/tmp/", Source:true, Console:true})

	s := NewTcpServer()

	option := &TOption{BufSize:1024}
	s.Setup("127.0.0.1", []int{6000}, option)

	s.Start()

	log := logger.GetLogger("tm")
	for {
		log.Info("links=%d, recvs=%d, sents=%d", s.GetSessionCount(), s.GetAllRecvBytes(), s.GetAllSentBytes())
		time.Sleep(time.Second * 3)
	}
}

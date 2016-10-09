package net

import (
	"net"
	"sync"
)

type tTcpServant struct {
	tNetServant
	server   *TTcpServer
	lsnr     *net.TCPListener
	sessLock sync.RWMutex
	sessions map[string]INetSession
}

func newTcpServant(_server *TTcpServer, _port int) (servant *tTcpServant) {
	_server.servLog.Info("newTcpServant %s:%d", _server.host, _port)
	servant = &tTcpServant{server:_server, sessLock:sync.RWMutex{}, sessions:make(map[string]INetSession)}
	servant.setup(&_server.tNetServer, _server.host, _port)
	return
}

func (this *tTcpServant)getSession(_key *string) (session INetSession) {
	this.sessLock.RLock()
	session = this.sessions[*_key]
	this.sessLock.RUnlock()
	return
}

func (this *tTcpServant)doStop() {
	log := this.server.lsnrLog

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on clean %v", v)
		}
		this.sessLock.RUnlock()
	}()

	this.lsnr.Close()

	this.sessLock.RLock()
	for _, ref := range this.sessions {
		sess := ref.(*tTcpSession)
		sess.conn.Close()
	}
}

/**
	构造监听地址，并发起监听
 */
func (this *tTcpServant)doStart() {
	log := this.server.lsnrLog

	addr := this.getLocal()
	log.Debug("doStart %s", addr)

	if laddr, ea := net.ResolveTCPAddr("tcp4", addr); ea != nil {
		panic(ea)
	} else {
		log.Info("try listen on %s", this.getKey())
		if lsnr, err := net.ListenTCP("tcp", laddr); nil != err {
			panic(err)
		} else {
			this.lsnr = lsnr
			go this.loopAccept()
			log.Info("listen on %s ok", this.getKey())
		}
	}
}

/**
	负责接收来自客户端的连接请求，并构造连接事件
 */
func (this *tTcpServant)loopAccept() {
	log := this.server.lsnrLog

	for this.server.IsActive() {
		if conn, err := this.lsnr.AcceptTCP(); nil != err {
			panic(err)
		} else {
			go func(_conn *net.TCPConn) {
				sess := newTcpSession(&this.tNetServant, _conn)
				go this.runSession(sess)
			}(conn)
		}
	}

	log.Info("-------loopAccept.end")
}

func (this *tTcpServant)runSession(_session *tTcpSession) {
	key := _session.key.String()

	this.sessLock.Lock()
	this.sessions[key] = _session
	this.sessLock.Unlock()

	_session.run()

	this.sessLock.Lock()
	delete(this.sessions, key)
	this.sessLock.Unlock()

	_session.clean()
}
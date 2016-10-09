package net

import (
	"net"
	"sync"
	"time"
	"sync/atomic"
)

type tUdpServant struct {
	tNetServant
	server   *tUdpServer
	conn     *net.UDPConn
	closed   int32
	sessLock *sync.Mutex
	sessions map[*net.UDPAddr]*tUdpSession
	chSend   chan *tUdpData
	chRecv   chan *tUdpData
}

func newUdpServant(_server *tUdpServer, _port int) (servant *tUdpServant) {
	sess := make(map[*net.UDPAddr]*tUdpSession)
	servant = &tUdpServant{server:_server, sessLock:&sync.Mutex{}, sessions:sess}
	servant.setup(&_server.tNetServer, _server.host, _port)
	return
}

func (this *tUdpServant)doStop() {
	log := this.server.lsnrLog

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on stop %v", v)
		}
		this.sessLock.Unlock()
	}()

	if nil != this.chSend {
		close(this.chSend)
		this.chSend = nil
	}

	if nil != this.chRecv {
		close(this.chRecv)
		this.chRecv = nil
	}

	if nil != this.conn {
		this.conn.Close()
		this.conn = nil
	}

	this.sessLock.Lock()
	for _, session := range this.sessions {
		session.clean()
	}
}

func (this *tUdpServant)checkTimout() {
	defer this.sessLock.Unlock()
	this.sessLock.Lock()

	utime := time.Now().Unix()
	for key, sess := range this.sessions {
		if int(utime - sess.utime) > this.server.option.Timeout {
			this.server.callback.OnBrok(sess)
			delete(this.sessions, key)
		}
	}
}

func (this *tUdpServant)loopIdle() {
	itv := time.Second * 5
	tm := time.NewTimer(itv)
	defer func() {
		if v := recover(); nil != v {
			this.server.servLog.Error("error on loopIdle %v", v)
		}
		tm.Stop()
	}()

	for this.server.IsActive() {
		select {
		case <-tm.C:
			this.checkTimout()
			tm.Reset(itv)
		}
	}
}

func (this *tUdpServant)doStart() {
	log := this.server.lsnrLog

	addr := this.getLocal()
	if uaddr, ea := net.ResolveUDPAddr("udp", addr); nil != ea {
		panic(ea)
	} else {
		log.Info("try listen on udp://%s", addr)

		if conn, el := net.ListenUDP("udp", uaddr); nil != el {
			panic(el)
		} else {
			this.conn = conn
			opt := this.server.option
			this.chSend = make(chan *tUdpData, opt.SndQLen)
			this.chRecv = make(chan *tUdpData, opt.RcvQLen)
			go this.loopSend()
			go this.loopIdle()
		}
	}
}

func (this *tUdpServant)close() {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {

	}
}

func getUdpAddr(p []byte) (addr *net.UDPAddr) {
	return &net.UDPAddr{IP:p[0:4], Port:makeWord(p[4], p[5])}
}

func (this *tUdpServant)loopSend() {
	log := this.server.lsnrLog

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on loopSend %v", v)
		}
		this.close()
	}()

	for data := range this.chSend {
		buf := data.buf
		for {
			if nw, e := this.conn.WriteToUDP(buf, data.addr); nil != e {
				log.Error("error on send %s", e.Error())
				break
			} else if nw > 0 {
				this.server.addSentBytes(nw)
				buf := buf[nw:]
				if len(buf) == 0 {
					break
				}
			}
		}
	}
}

func (this *tUdpServant)loopRead() {
	log := this.server.lsnrLog

	obj := this.server.bufPool.Pop()
	defer func() {
		this.server.bufPool.Push(obj)
		this.close()
	}()

	mem := obj.Data()
	for this.server.IsActive() {
		if nr, addr, er := this.conn.ReadFromUDP(mem); nil != er {
			log.Error("error on read %s", er.Error())
			break
		} else if nr > 0 {
			buf := make([]byte, nr)
			copy(buf, mem[:nr])
			this.chRecv <- &tUdpData{addr:addr, buf:buf}
		} else {
			log.Error("error on read %d", nr)
		}
	}
}

func (this *tUdpServant)loopData() {
	log := this.server.servLog

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on loopData %v", v)
		}
		this.close()
	}()

	for d := range this.chRecv {

		this.sessLock.Lock()
		session := this.sessions[d.addr]
		this.sessLock.Unlock()

		if nil == d.buf {
			if nil != session {
				this.server.onBrok(session)
			}
		} else {
			if nil == session {
				session = newUdpSession(this, d.addr)

				this.sessLock.Lock()
				this.sessions[d.addr] = session
				this.sessLock.Unlock()

				this.server.onConn(session)
			}

			session.utime = time.Now().Unix()
			this.server.onData(session, d.buf)
		}
	}
}

func (this *tUdpServant)sendTo(_addr *net.UDPAddr, _buf []byte) {
	this.chSend <- &tUdpData{addr:_addr, buf:_buf}
}

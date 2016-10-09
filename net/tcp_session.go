package net

import (
	"net"
	"sync/atomic"
	"sync"
	"fmt"
)

type tTcpSession struct {
	tNetSession
	owner  *tNetServant
	conn   net.Conn
	chSend chan []byte
	chRecv chan []byte
}

func newTcpSession(_owner *tNetServant, _conn *net.TCPConn) (*tTcpSession) {
	server := _owner.server
	opt := server.option

	session := &tTcpSession{
		owner:_owner,
		chSend:make(chan []byte, opt.SndQLen),
		chRecv:make(chan []byte, opt.RcvQLen),
		conn:_conn}

	key_str := fmt.Sprintf("T%.4x-%s/%s",
		server.genSessionSeq(),
		_owner.getLocal(),
		_conn.RemoteAddr().String())

	session.setup(_owner.server, MakeAddrByString(&key_str))

	return session
}

func (this *tTcpSession)run() {
	this.server.onConn(this)

	wg := &sync.WaitGroup{}
	wg.Add(3)

	go this.loopSend(wg)
	go this.loopRecv(wg)
	go this.loopData(wg)

	wg.Wait()

	this.server.onBrok(this)
}

func (this *tTcpSession)close(){
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		close(this.chSend)
		close(this.chRecv)
		this.conn = nil
	}
}

func (this *tTcpSession)loopData(_wg *sync.WaitGroup) {
	log := this.server.sessLog

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on data (%v)", v)
		}
		_wg.Add(-1)
		this.close()
	}()

	for data := range this.chRecv {
		this.server.onData(this, data)
	}
}

func (this *tTcpSession)loopSend(_wg *sync.WaitGroup) {
	log := this.server.sessLog

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on send (%v)", v)
		}
		_wg.Add(-1)
		this.close()
	}()

	for buf := range this.chSend {
		this.doSend(buf)
	}
}

func (this *tTcpSession)doSend(buf []byte){
	for {
		if nw, e := this.conn.Write(buf); nil != e {
			this.server.sessLog.Error("%s error on send %s", this.key, e.Error())
			return
		} else if nw == 0 {
			this.server.sessLog.Warn("%s error on send ZERO", this.key)
			return
		} else {
			this.server.addSentBytes(nw)

			buf = buf[nw:]
			if len(buf) == 0 {
				break
			}
		}
	}
}

//todo 这个函数可以优化以提升性能
func (this *tTcpSession)loopRecv(_wg *sync.WaitGroup) {
	log := this.server.sessLog

	server := this.server
	buf_obj := server.bufPool.Pop()

	defer func() {
		if v := recover(); nil != v {
			log.Error("error on recv (%v)", v)
		}
		server.bufPool.Push(buf_obj)
		this.close()
		_wg.Add(-1)
	}()

	buf := buf_obj.Data()
	for {
		if nr, e := this.conn.Read(buf); nil != e {
			log.Error("read error: (%s)", e.Error())
			break
		} else if 0 == nr {
			log.Warn("[%s] read zero", this.key)
			break
		} else {
			this.server.addRecvBytes(nr)

			data := make([]byte, nr)
			copy(data, buf[:nr])
			this.chRecv <- data
		}
	}
}

func (this *tTcpSession)Write(_buf []byte) {
	this.chSend <- _buf
}

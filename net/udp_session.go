package net

import (
	"fmt"
	"net"
	"time"
)

type tUdpSession struct {
	tNetSession
	owner *tUdpServant
	addr  *net.UDPAddr
	utime int64
}

type tUdpData struct {
	addr *net.UDPAddr
	buf  []byte // nil means brok
}

func newUdpSession(_owner *tUdpServant, _addr *net.UDPAddr) (session *tUdpSession) {
	session = &tUdpSession{owner:_owner, addr:_addr, utime:time.Now().Unix()}
	key_str := fmt.Sprintf("")
	key := MakeAddrByString(&key_str)
	session.setup(&_owner.server.tNetServer, key)
	return
}

func (this *tUdpSession)Write(_buf []byte) {
	this.owner.sendTo(this.addr, _buf)
}

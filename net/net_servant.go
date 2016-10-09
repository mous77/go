package net

import(
	"fmt"
)

type tNetServant struct {
	server   *tNetServer
	host     string
	port     int
}

func (this *tNetServant)setup(_server *tNetServer, _host string, _port int) {
	this.server = _server
	this.host = _host
	this.port = _port
}

func (this *tNetServant)getLocal() (string) {
	return fmt.Sprintf("%s:%d", this.host, this.port)
}

func (this *tNetServant)getKey() (string) {
	return fmt.Sprintf("tcp://%s:%d", this.host, this.port)
}

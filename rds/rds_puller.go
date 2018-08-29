package rds

import (
	"sync"
	"sync/atomic"
	"time"
	"lib/utils"
	"fmt"
)

type RdsPuller struct {
	active int32
	fire   *RdsFire
	client *RdsClient
	topic  string
	wg     *sync.WaitGroup
}

func NewRdsPuller(_conf *RdsConfig, _topic string, _on_data OnRdsData) (*RdsPuller) {
	return &RdsPuller{
		topic:  _topic,
		fire:   NewRdsFire("Rdspuller", _on_data),
		client: NewRdsClient(_conf),
	}
}

func (p *RdsPuller) IsActive() (bool) {
	return atomic.LoadInt32(&p.active) == 1
}

func (p *RdsPuller) loop(_arg interface{}) {
	defer func() {
		p.wg.Done()
	}()

	for p.IsActive() {
		ary := p.client.BLpop(time.Second, p.topic)
		if len(ary) < 2 {
			continue
		}

		p.fire.Offer(ary[0], ary[1])
	}
}

func (p *RdsPuller) Start() {
	if atomic.CompareAndSwapInt32(&p.active, 0, 1) {
		p.fire.Start()
		p.client.Start()

		p.wg = &sync.WaitGroup{}
		p.wg.Add(1)
		utils.GoRunAdd(p.loop, "RdsPuller.loop")
	}
}

func (p *RdsPuller) Stop() {
	if atomic.CompareAndSwapInt32(&p.active, 1, 0) {
		p.wg.Wait()

		p.client.Stop()
		p.fire.Stop()
	}
}

func TestPuller() {
	conf := NewRdsConf()
	subs := NewRdsPuller(conf, "CMDRT", func(_data *RdsData) {
		fmt.Printf("%s\r\n", _data.String())
	})
	subs.Start()
	time.Sleep(time.Hour)
}

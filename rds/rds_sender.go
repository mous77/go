package rds

import (
	"fmt"
	"lib/log"
	"sync/atomic"
	"sync"
	"lib/utils"
)

type TaskType string

const (
	rttNON     = "NON"
	rttSET     = "SET"
	rttHMSET   = "HMSET"
	rttPUSH    = "PUSH"
	rttPUBLISH = "PUBLISH"
)

type RdsTask struct {
	typ  TaskType
	key  string
	data string
	maps map[string]interface{}
}

func (t *RdsTask) String() (string) {
	return fmt.Sprintf("%this:%this=%this|%v", t.typ, t.key, t.data, t.maps)
}

func newRdsTask(_typ TaskType, _key string, _data string, _map map[string]interface{}) (*RdsTask) {
	return &RdsTask{typ: _typ, key: _key, data: _data, maps: _map}
}

type RdsSender struct {
	lg      log.ILogger
	name    string
	active  int32
	killSig *RdsTask
	wg      *sync.WaitGroup
	client  *RdsClient
	chTask  chan *RdsTask
}

func NewRdsSender(_conf *RdsConfig) (*RdsSender) {
	return &RdsSender{
		active:  0,
		lg:      log.GetLogger("rdsSender"),
		name:    fmt.Sprintf("rdsSender:%s", _conf.String()),
		killSig: &RdsTask{typ: rttNON},
		client:  NewRdsClient(_conf),
		wg:      &sync.WaitGroup{},
	}
}

func (this *RdsSender) exec(_task *RdsTask) {
	defer func() {
		if v := recover(); nil != v {
			this.lg.Error("error on exec(%s) %v", _task.String(), v)
		}
	}()

	switch _task.typ {
	case rttPUSH:
		this.client.RPush(_task.key, _task.data)
	case rttPUBLISH:
		this.client.Publish(_task.key, _task.data)
	case rttSET:
		this.client.Set(_task.key, _task.data, -1)
	case rttHMSET:
		this.client.HMset(_task.key, _task.maps)
	default:
	}
}

func (this *RdsSender) Set(_key string, _val string) {
	this.chTask <- newRdsTask(rttSET, _key, _val, nil)
}

func (this *RdsSender) Push(_topic string, _msg string) {
	this.chTask <- newRdsTask(rttPUSH, _topic, _msg, nil)
}

func (this *RdsSender) Publish(_topic string, _msg string) {
	this.chTask <- newRdsTask(rttPUBLISH, _topic, _msg, nil)
}

func (this *RdsSender) Hset(_key string, _fld string, _val string) {
	m := map[string]interface{}{_fld: _val}
	this.chTask <- newRdsTask(rttHMSET, _key, "", m)
}

func (this *RdsSender) Hmset(_key string, _map map[string]interface{}) {
	this.chTask <- newRdsTask(rttHMSET, _key, "", _map)
}

func (this *RdsSender) loop(_arg interface{}) {
	defer this.wg.Done()

	for task := range this.chTask {
		if task == this.killSig {
			break
		} else {
			this.exec(task)
		}
	}
}

func (this *RdsSender) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.chTask = make(chan *RdsTask, 1024)
		this.wg.Add(1)
		this.client.Start()
		utils.GoRunAdd(this.loop, this.name)
	}
}

func (this *RdsSender) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		this.chTask <- this.killSig
		this.wg.Wait()
		close(this.chTask)
		this.client.Stop()
	}
}

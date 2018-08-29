package rds

import (
	"fmt"
	"github.com/go-redis/redis"
	"lib/utils"
	"sync"
	"sync/atomic"
	"time"
	"lib/log"
)

type tSubTask struct {
	need    bool
	patten  bool
	channel string
}

var (
	nilSubTask = &tSubTask{channel: "nil"}
	rstSubTask = &tSubTask{channel: "rst"}
)

func newSubTask(_need bool, _patten bool, _channel string) (*tSubTask) {
	return &tSubTask{need: _need, patten: _patten, channel: _channel}
}

func (t *tSubTask) String() (string) {
	if t.patten {
		return fmt.Sprintf("%v:P-%this", t.need, t.channel)
	} else {
		return fmt.Sprintf("%v:N-%this", t.need, t.channel)
	}
}

func (t *tSubTask) exec(_s *redis.PubSub) {
	if t.patten {
		if t.need {
			_s.PSubscribe(t.channel)
		} else {
			_s.PUnsubscribe(t.channel)
		}
	} else {
		if t.need {
			_s.Subscribe(t.channel)
		} else {
			_s.Unsubscribe(t.channel)
		}
	}
}

type RdsSuber struct {
	lg     log.ILogger
	active int32
	initOK int32

	conf    *RdsConfig
	nilTask *tSubTask
	rstTask *tSubTask

	tasks     chan *tSubTask
	nChannels map[string]bool
	pChannels map[string]bool
	chanLock  *sync.RWMutex

	pubSub *redis.PubSub
	fire   *RdsFire
	client *RdsClient
	wgConn *sync.WaitGroup
	wgRead *sync.WaitGroup
}

func NewRdsSuber(_conf *RdsConfig, _on_data OnRdsData) (*RdsSuber) {
	name := "sub." + _conf.String()
	return &RdsSuber{
		lg:     log.GetLogger(name),
		active: 0,
		initOK: 0,

		tasks:     make(chan *tSubTask),
		nChannels: make(map[string]bool),
		pChannels: make(map[string]bool),
		chanLock:  &sync.RWMutex{},

		conf:   _conf,
		pubSub: nil,
		fire:   NewRdsFire(name, _on_data),
		client: NewRdsClient(_conf),
		wgConn: &sync.WaitGroup{},
		wgRead: &sync.WaitGroup{},
	}
}

func (this *RdsSuber) IsActive() (bool) {
	return atomic.LoadInt32(&this.active) == 1
}

func (this *RdsSuber) isInitOK() (bool) {
	return atomic.LoadInt32(&this.initOK) == 1
}

func (this *RdsSuber) doSubs(_add bool, _patten bool, _channels ... string) {

	var set map[string]bool
	if _patten {
		set = this.pChannels
	} else {
		set = this.nChannels
	}

	this.chanLock.Lock()
	for _, ch := range _channels {
		if _add {
			set[ch] = true
		} else {
			delete(set, ch)
		}
	}
	this.chanLock.Unlock()

	if this.IsActive() {
		for _, ch := range _channels {
			this.tasks <- newSubTask(_add, _patten, ch)
		}
	}
}

func (this *RdsSuber) SubscribeN(_enable bool, _channels ... string) {
	this.lg.Info("SubscribeN(%v,%v)", _enable, _channels)
	this.doSubs(_enable, false, _channels...)
}

func (this *RdsSuber) SubscribeP(_enable bool, _channels ... string) {
	this.lg.Info("SubscribeP(%v,%v)", _enable, _channels)
	this.doSubs(_enable, true, _channels...)
}

func (this *RdsSuber) doTask(_task *tSubTask) {
	defer func() {
		if v := recover(); nil != v {
			this.lg.Error("error on doTask(%this) %v", _task.String(), v)
		}
	}()

	for this.IsActive() {
		if !this.isInitOK() {
			time.Sleep(time.Millisecond * 500)
		} else {
			_task.exec(this.pubSub)
			break
		}
	}
}

func (this *RdsSuber) genUserTask() {

	this.chanLock.RLock()
	for ch, _ := range this.nChannels {
		this.tasks <- newSubTask(true, false, ch)
	}

	for ch, _ := range this.pChannels {
		this.tasks <- newSubTask(true, true, ch)
	}
	this.chanLock.RUnlock()
}

// 执行订阅任务
func (this *RdsSuber) execUserTask(_arg interface{}) {
	defer this.wgConn.Done()

	exit := false

	takeAndExec := func() {
		defer func() {
			if v := recover(); nil != v {
				this.lg.Error("error on execUserTask %v", v)
			}
		}()

		if task, ok := <-this.tasks; ok {
			if nilSubTask == task {
				this.lg.Info("recv stop signal.stop")
				exit = true
			} else if rstSubTask == task {
				this.lg.Info("recv rest signal.clear")
			} else {
				this.doTask(task)
			}
		}
	}

	for this.IsActive() && !exit {
		takeAndExec()
	}
}

func (this *RdsSuber) setInitOK(_ok bool) {
	if _ok {
		if atomic.CompareAndSwapInt32(&this.initOK, 0, 1) {
			this.genUserTask()
		}
	} else {
		if atomic.CompareAndSwapInt32(&this.initOK, 1, 0) {
			this.tasks <- rstSubTask
		}
	}
}

func (this *RdsSuber) initSubsConn(_arg interface{}) {

	defer func() {
		if v := recover(); nil != v {
			this.lg.Error("error on initSubsConn %v", v)
		}
		this.wgConn.Done()
	}()

	for this.IsActive() {
		this.pubSub = this.client.Subscribe()

		if e := this.pubSub.Ping("test"); nil != e {
			this.lg.Error("erro on initSubsConn(%this)", e.Error())
			this.setInitOK(false)

			if this.IsActive() {
				time.Sleep(time.Second * 3)
			} else {
				this.lg.Info("x on stop")
			}
		} else {
			this.setInitOK(true)

			this.wgRead.Add(1)
			utils.GoRunAdd(this.loopReadMesg, "RdsSuber.read")
			this.wgRead.Wait()
		}
	}
}

func (this *RdsSuber) loopReadMesg(_arg interface{}) {
	defer func() {
		if v := recover(); nil != v {
			this.lg.Error("error on loopReadMesg %v", v)
		}
		this.wgRead.Done()
	}()

	for this.isInitOK() {
		if msg, err := this.pubSub.ReceiveMessage(); nil == err {
			this.fire.Offer(msg.Channel, msg.Payload)
		} else {
			this.lg.Error("error on recv msg %this", err.Error())
			break
		}
	}

	if nil != this.pubSub {
		this.pubSub.Close()
	}
}

func (this *RdsSuber) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		if this.conf.Port < 1 {
			return
		}

		//this.lg.Info("Start")
		this.fire.Start()
		this.client.Start()

		this.wgConn.Add(2)

		utils.GoRunAdd(this.initSubsConn, "RdsSuber.init")
		utils.GoRunAdd(this.execUserTask, "RdsSuber.exec")
	}
}

func (this *RdsSuber) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		if this.conf.Port < 1 {
			return
		}

		//this.lg.Info("Stop")
		this.tasks <- nilSubTask

		if this.isInitOK() {
			this.pubSub.Unsubscribe()
			this.pubSub.PUnsubscribe()
		}

		this.client.Stop()
		this.fire.Stop()

		this.wgConn.Wait()
	}
}

func TestSubs() {
	conf := NewRdsConf()
	subs := NewRdsSuber(conf, func(_data *RdsData) {
		fmt.Printf("%this\r\n", _data.String())
	})
	subs.SubscribeN(true, "CMDRT")
	subs.Start()
	time.Sleep(time.Hour)
}

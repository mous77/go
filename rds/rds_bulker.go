package rds

import (
	"github.com/go-redis/redis"
	"time"
	"fmt"
	"lib/log"
	"lib/utils"
	"sync/atomic"
)

type RdsBulkTask interface {
	Exec(pipeline redis.Pipeliner)
}

type tNoop struct {
}

type RdsBulker struct {
	lg        log.ILogger
	name      string
	active    int32
	batchSize int
	delayTime time.Duration
	qTasks    *utils.TQueue
	chNoop    chan *tNoop
	rdsClient *RdsClient
	noopSig   *tNoop
}

func NewRdsBulker(_cfg *RdsConfig, _batch_size int, _delay_time time.Duration) (*RdsBulker) {
	if 100 > _batch_size {
		_batch_size = 100
	}

	return &RdsBulker{
		lg:        log.GetLogger("RdsBulker"),
		name:      fmt.Sprintf("RdsBulker:%s", _cfg.String()),
		active:    0,
		batchSize: _batch_size,
		delayTime: _delay_time,
		qTasks:    utils.NewQueue(0),
		rdsClient: NewRdsClient(_cfg),
		noopSig:   &tNoop{},
	}
}

func (this *RdsBulker) genNoop() {
	if atomic.LoadInt32(&this.active) > 0 {
		this.chNoop <- this.noopSig
	}
}

func (this *RdsBulker) doBulk(_items []RdsBulkTask) {
	defer func() {
		if v := recover(); nil != v {
			this.lg.Error("error on doBulk %v", v)
		}
	}()

	pp := this.rdsClient.Pipeline()
	for _, task := range _items {
		task.Exec(pp)
	}
	pp.Exec()

}

func (this *RdsBulker) Offer(_task RdsBulkTask) {
	this.qTasks.Offer(_task)
}

func (this *RdsBulker) loop(_arg interface{}) {

	buffer := make([]RdsBulkTask, this.batchSize)
	items := buffer[:0]

	execTask := func() {
		for !this.qTasks.IsEmpty() {
			if ref, ok := this.qTasks.Poll(); ok {
				items = append(items, ref.(RdsBulkTask))
			}

			if len(items) == this.batchSize {
				this.doBulk(items)
				items = buffer[:0]
			}
		}

		if len(items) > 0 {
			this.doBulk(items)
			items = buffer[:0]
		}
	}

	for atomic.LoadInt32(&this.active) > 0 {
		execTask()
		<-this.chNoop
	}

	execTask()
}

func (this *RdsBulker) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		this.rdsClient.Start()
		utils.GoSchedule(this.genNoop, this.delayTime, "rdsBulker.genNoop")
	}
}

func (this *RdsBulker) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		this.lg.Info("stop")
		this.rdsClient.Stop()
		this.lg.Info("stop.done")
	}
}

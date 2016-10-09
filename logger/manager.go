package logger

import (
	"time"
	"sync"
	"github.com/mous77/go/utils"
)

type tLogManager struct {
	writers *tLogWriters
	locks   *sync.Mutex
	insts   map[string]ILogger
	items   *utils.TQueue
	config  *TConfig
	active  int32
}

func newManager() *tLogManager {
	return &tLogManager{
		active:0,
		locks: &sync.Mutex{},
		insts: make(map[string]ILogger),
		config:newDefConfig(),
		items: utils.NewQueue()}
}

func (this *tLogManager)run() {

	defer func() {
		writers.close()
		for k, _ := range this.insts {
			delete(this.insts, k)
		}
	}()

	terminated := false
	need_flush := false
	last_flush := time.Now()

	var item *tLogItem
	const itv = time.Second * 1

	for !terminated {
		if ref, ok := this.items.Poll(itv); ok {
			need_flush = true
			item = ref.(*tLogItem)
			switch item.iType {
			case itDATA:
				writers.recv(item)
			case itDONE:
				terminated = true
			}
		} else {
			if need_flush && last_flush.Add(itv).Before(time.Now()) {
				writers.flush()
				last_flush = time.Now()
				need_flush = false
			}
		}
	}
	println("end of loop on write log")
}

func (this *tLogManager)getLogger(_key string) ILogger {
	defer this.locks.Unlock()
	this.locks.Lock()

	lg := this.insts[_key]
	if nil == lg {
		cfg := newDefConfig()
		cfg.copyFrom(this.config)

		lg = &tLogInstance{config:cfg, key:_key, items: this.items}
		this.insts[_key] = lg
	}
	return lg
}


package logger

import (
	"sync/atomic"
	"time"
	"os"
	"fmt"
)

var (
	manager *tLogManager = newManager()
	writers *tLogWriters = newWriters()
)

type ILogWriter interface {
	Open()
	Recv(_line string)
	Flush()
	Close()
}

type ILogger interface {
	Printf(string, ...interface{})
	Fine(string, ...interface{})
	Debug(string, ...interface{})
	Trace(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Critical(string, ...interface{})
}

func Setup(_cfg *TConfig) {
	//fmt.Printf("setup(%s)\r\n", _cfg)
	if atomic.CompareAndSwapInt32(&manager.active, 0, 1) {
		manager.config.copyFrom(_cfg)
		writers.setup()
		go manager.run()
	} else {
		fmt.Println("logger setup already!")
	}
}

func Close() {
	if atomic.CompareAndSwapInt32(&manager.active, 1, 0) {
		done := &tLogItem{iType:itDONE}
		manager.items.Offer(done)
	}
}

func GetLogger(_name string) ILogger {
	if atomic.LoadInt32(&manager.active) < 1 {
		println("please call logger.Setup(config) first!")
		os.Exit(0)
	}
	//fmt.Printf("GetLogger %s", _name)
	return manager.getLogger(_name)
}

func TestLog() {
	Setup(&TConfig{App:"demo", Root:"c:/tmp/", Source:true, Console:false})
	logger := GetLogger("demo")
	var seq int = 0
	for {
		logger.Info("Hello %d, %s", seq, "wrz")
		seq++
		time.Sleep(time.Nanosecond * 1)
	}
	Close()
}
package log

import (
	"sync/atomic"
	"time"
	"os"
	"fmt"
)

var (
	manager = newManager()
)

type ILogHandler interface {
	Recv(_msg string)
}

type ILogWriter interface {
	Open()
	Recv(_msg string)
	Flush()
	Close()
}

type ILogger interface {
	IsDebug() bool
	IsInfo() bool
	IsTrace() bool
	IsWarn() bool
	IsError() bool
	IsCritical() bool
	IsPanic() bool

	Printf(string, ...interface{})

	Fine(string, ...interface{})
	Debug(string, ...interface{})
	Trace(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{})
	Error(string, ...interface{})
	Critical(string, ...interface{})
	Panic(string, ...interface{})
}

func Setup(_cfg *TConfig) {
	//fmt.Printf("setup(%s)\r\n", _cfg)
	if atomic.CompareAndSwapInt32(&manager.active, 0, 1) {
		manager.config.copyFrom(_cfg)

		if nil != _cfg {
			manager.setup(_cfg.Custom)
		}

		go manager.run()
	} else {
		fmt.Println("logger setup already!")
	}
}

func Close() {
	if atomic.CompareAndSwapInt32(&manager.active, 1, 0) {
		manager.items <- &tLogItem{iType: itDONE}
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

func TestLog(_handler ILogHandler) {
	Setup(&TConfig{App: "demo", Root: "c:/tmp/", Console: true, Custom: _handler})
	logger := GetLogger("demo")
	var seq int = 0
	for {
		logger.Info("Hello %d, %s", seq, "wrz")
		seq++
		time.Sleep(time.Millisecond * 1000)
	}
	Close()
}

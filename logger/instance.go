package logger

import (
	"runtime"
	"strings"
	"fmt"
	"github.com/mous77/go/utils"
)

type tLogInstance struct {
	key    string
	items  *utils.TQueue
	config *TConfig
}

func (this *tLogInstance)doLog(_level TLogLevel, _fmt string, _args ...interface{}) {
	if _level < this.config.Level {
		println("logger.skip", _level, _fmt, _args)
		return
	}

	var (
		file string = ""
		line int = 0
		ok bool = false
	)

	if this.config.Source {
		_, file, line, ok = runtime.Caller(2)
		if ok {
			if idx := strings.LastIndex(file, "/"); idx > 0 {
				file = file[idx + 1:]
			}
		}
	}

	msg := _fmt
	if len(_args) > 0 {
		msg = fmt.Sprintf(_fmt, _args...)
	}

	item := newLogItem(this, _level, msg, file, line)
	this.items.Offer(item)
}

func (this *tLogInstance)Printf(_fmt string, _args ...interface{}) {
	this.doLog(LLINFO, _fmt, _args...)
}

func (this *tLogInstance)Fine(_fmt string, _args ...interface{}) {
	this.doLog(LLFINE, _fmt, _args...)
}

func (this *tLogInstance)Trace(_fmt string, _args ...interface{}) {
	this.doLog(LLTRACE, _fmt, _args...)
}

func (this *tLogInstance)Debug(_fmt string, _args ...interface{}) {
	this.doLog(LLDEBUG, _fmt, _args...)
}

func (this *tLogInstance)Info(_fmt string, _args ...interface{}) {
	this.doLog(LLINFO, _fmt, _args...)
}

func (this *tLogInstance)Warn(_fmt string, _args ...interface{}) {
	this.doLog(LLWARN, _fmt, _args...)
}

func (this *tLogInstance)Error(_fmt string, _args ...interface{}) {
	this.doLog(LLERROR, _fmt, _args...)
}

func (this *tLogInstance)Critical(_fmt string, _args ...interface{}) {
	this.doLog(LLCRITICAL, _fmt, _args...)
}

package log

import (
	"runtime"
	"strings"
	"fmt"
)

type tLogInstance struct {
	key    string
	items  chan<- *tLogItem
	config *TConfig
}

func (this *tLogInstance) doLog(_level TLogLevel, _fmt string, _args ...interface{}) {
	if _level < this.config.Level {
		println("logger.skip", _level, _fmt, _args)
		return
	}

	var (
		file = ""
		line = 0
		ok   = false
	)

	if this.config.Source {
		_, file, line, ok = runtime.Caller(2)
		if ok {
			if idx := strings.LastIndex(file, "/"); idx > 0 {
				file = file[idx+1:]
			}
		}
	}

	msg := _fmt
	if len(_args) > 0 {
		msg = fmt.Sprintf(_fmt, _args...)
	}

	this.items <- newLogItem(this, _level, msg, file, line)
}

func (this *tLogInstance) Printf(_fmt string, _args ...interface{}) {
	if this.config.Level <= LLINFO {
		this.doLog(LLINFO, _fmt, _args...)
	}
}

func (this *tLogInstance) IsFine() bool {
	return this.config.Level <= LLFINE
}

func (this *tLogInstance) Fine(_fmt string, _args ...interface{}) {
	if this.IsFine() {
		this.doLog(LLFINE, _fmt, _args...)
	}
}

func (this *tLogInstance) IsTrace() bool {
	return this.config.Level <= LLTRACE
}

func (this *tLogInstance) Trace(_fmt string, _args ...interface{}) {
	if this.IsTrace() {
		this.doLog(LLTRACE, _fmt, _args...)
	}
}

func (this *tLogInstance) IsDebug() bool {
	return this.config.Level <= LLDEBUG
}

func (this *tLogInstance) Debug(_fmt string, _args ...interface{}) {
	if this.IsDebug() {
		this.doLog(LLDEBUG, _fmt, _args...)
	}
}

func (this *tLogInstance) IsInfo() bool {
	return this.config.Level <= LLINFO
}
func (this *tLogInstance) Info(_fmt string, _args ...interface{}) {
	if this.IsInfo() {
		this.doLog(LLINFO, _fmt, _args...)
	}
}

func (this *tLogInstance) IsWarn() bool {
	return this.config.Level <= LLWARN
}

func (this *tLogInstance) Warn(_fmt string, _args ...interface{}) {
	if this.IsWarn() {
		this.doLog(LLWARN, _fmt, _args...)
	}
}

func (this *tLogInstance) IsError() bool {
	return this.config.Level <= LLERROR
}

func (this *tLogInstance) Error(_fmt string, _args ...interface{}) {
	if this.IsError() {
		this.doLog(LLERROR, _fmt, _args...)
	}
}

func (this *tLogInstance) IsCritical() bool {
	return this.config.Level <= LLCRITICAL
}

func (this *tLogInstance) Critical(_fmt string, _args ...interface{}) {
	if this.IsCritical() {
		this.doLog(LLCRITICAL, _fmt, _args...)
	}
}

func (this *tLogInstance) IsPanic() bool {
	return this.config.Level <= LLPANIC
}

func (this *tLogInstance) Panic(_fmt string, _args ...interface{}) {
	if this.IsPanic() {
		this.doLog(LLPANIC, _fmt, _args...)
	}
}

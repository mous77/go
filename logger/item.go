package logger

import (
	"time"
	"fmt"
	"sync/atomic"
)

type tItemType int

const (
	itDATA tItemType = iota
	itDONE
)

type tLogItem struct {
	seq   uint64
	from  *tLogInstance
	iType tItemType
	key   string
	level TLogLevel
	time  time.Time
	msg   string
	file  string
	line  int
}

// EF FF FF FF
// EF FF FF FF
var g_log_seq uint64

func newLogItem(_from *tLogInstance, _level TLogLevel, _msg, _file string, _line int) (*tLogItem) {
	atomic.CompareAndSwapUint64(&g_log_seq, 0xEFFFFFFF, 0)
	return &tLogItem{
		from:_from,
		iType:itDATA,
		key:_from.key,
		level:_level,
		seq:atomic.AddUint64(&g_log_seq, 1),
		time:time.Now(),
		msg:_msg,
		file:_file,
		line:_line}
}

const (
	dateTimeFormat = "2006-01-02 15:04:05"
)

func (this *tLogItem)String() string {
	time_str := this.time.Format(dateTimeFormat)
	if this.from.config.Source {
		return fmt.Sprintf("%.8X %s [%s] %s (%s:%d) : %s\r\n",
			this.seq, time_str, c_logLevelNames[this.level], this.key,
			this.file, this.line,
			this.msg)
	} else {
		return fmt.Sprintf("%.8X %s [%s] %s : %s\r\n",
			this.seq, time_str, c_logLevelNames[this.level], this.key,
			this.msg)
	}
}
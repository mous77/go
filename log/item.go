package log

import (
	"time"
	"fmt"
)

type tItemType int

const (
	itDATA tItemType = iota
	itDONE
)

type tLogItem struct {
	from  *tLogInstance
	iType tItemType
	key   string
	level TLogLevel
	time  time.Time
	msg   string
	file  string
	line  int
}

func newLogItem(_from *tLogInstance, _level TLogLevel, _msg, _file string, _line int) (*tLogItem) {
	return &tLogItem{
		from:_from,
		iType:itDATA,
		key:_from.key,
		level:_level,
		time:time.Now(),
		msg:_msg,
		file:_file,
		line:_line}
}

const (
	dateTimeFormat = "2006-01-02 15:04:05.000"
)

func (item *tLogItem)String() string {
	time_str := item.time.Format(dateTimeFormat)
	if item.from.config.Source {
		return fmt.Sprintf("%s [%s] %s (%s:%d) : %s",
			time_str, c_logLevelNames[item.level], item.key,
			item.file, item.line,
			item.msg)
	} else {
		return fmt.Sprintf("%s [%s] %s : %s",
			time_str, c_logLevelNames[item.level], item.key,
			item.msg)
	}
}
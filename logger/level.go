package logger

type TLogLevel int

const (
	LLFINE TLogLevel = iota
	LLDEBUG
	LLTRACE
	LLINFO
	LLWARN
	LLERROR
	LLCRITICAL
)

var (
	c_logLevelNames = [...]string{"FINE", "DEBG", "TRAC", "INFO", "WARN", "EROR", "CRIT"}
)

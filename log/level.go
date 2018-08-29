package log

import "strings"

type TLogLevel int

// F D T I W E C P
const (
	LLFINE     TLogLevel = iota
	LLTRACE
	LLDEBUG
	LLINFO
	LLWARN
	LLERROR
	LLCRITICAL
	LLPANIC
)

var (
	c_logLevelNames = [...]string{"FINE", "TRAC", "DEBG", "INFO", "WARN", "EROR", "CRIT", "PANI"}
)

func LevelByName(_name string) (TLogLevel) {
	name := strings.ToUpper(_name)
	for i, n := range c_logLevelNames {
		if strings.Index(name, n) == 0 {
			return TLogLevel(i)
		}
	}
	return LLINFO
}

package logger

import (
	"fmt"
	"os"
)

type TConfig struct {
	App     string
	Root    string
	Level   TLogLevel
	Source  bool
	Console bool
	GenFile bool
}

func newDefConfig() (*TConfig) {
	return &TConfig{App: "app", Root:".", Source:true, Level:LLINFO, Console:true}
}
/*
log file name format is : ./logs/app.ymd.log
 */
func (this *TConfig)getLogFileName(_ymd, _hms *string) (string) {
	return fmt.Sprintf("%s/%s/%s.%s.log", this.Root, *_ymd, this.App, *_hms)
}

func (this *TConfig)copyFrom(_c *TConfig) {
	if nil == _c {
		println("*logger.TConfig.copyFrom : warn : config is null")
		return
	}

	if len(_c.App) > 0 {
		this.App = _c.App
	}

	root := _c.Root;
	cnt := len(root)
	if cnt > 0 {
		if root[cnt - 1:] == "/" {
			root = root[0:cnt - 1]
		}
	} else {
		root, _ = os.Getwd()
	}
	this.Root = root

	this.Level = _c.Level
	this.Source = _c.Source
	this.Console = _c.Console
	this.GenFile = _c.GenFile
}

package app

import (
	"lib/log"
	"os"
	"strings"
	"fmt"
	"errors"
	"os/signal"
	"syscall"
	"sync/atomic"
	"bufio"
	"time"
	"os/exec"
)

func hasEcho() bool {
	for _, arg := range os.Args {
		if strings.Compare(arg, "-echo") == 0 {
			return true
		}
	}
	return false
}

func IsMaster() (bool) {
	for _, arg := range os.Args {
		if strings.Compare(arg, "-master") == 0 {
			return true
		}
	}
	return false
}

func fileExists(_file_name string) (bool) {
	exists := true
	if _, err := os.Stat(_file_name); os.IsNotExist(err) {
		exists = false
	}
	return exists
}

func checkPidsDir() {
	if info, err := os.Stat("pids"); os.IsNotExist(err) {
		os.MkdirAll("pids", 0777)
	} else if !info.IsDir() {
		panic(errors.New("can not found pids dir"))
	}
}

func customLogfile(_app string) log.ILogger {
	cfg := log.NewDefConfig()
	cfg.App = _app + ".master"
	cfg.Level = log.LLDEBUG
	log.Setup(cfg)
	return log.GetLogger("master")
}

func RunMaster(_app string) {
	checkPidsDir()

	pwd, _ := os.Getwd()
	ctrlFileName := fmt.Sprintf("%s/pids/.pid.%s.master", pwd, _app)

	if fileExists(ctrlFileName) {
		panic(errors.New(_app + " master has run already!"))
	}

	if fileExists(fmt.Sprintf("pids/.pid.%s.slave", _app)) {
		panic(errors.New(_app + " slave has run already!"))
	}

	if fileExists(fmt.Sprintf("pids/.pid.%s.alone", _app)) {
		panic(errors.New(_app + " has run as alone already!"))
	}

	active := int32(1)
	echo := hasEcho()
	ctrlPID := fmt.Sprintf("%d", os.Getpid())
	exe, _ := os.Executable()
	lg := customLogfile(_app)

	lg.Info("ctrlFile is %s", ctrlFileName)

	waitForExit := func() {
		exitSignals := make(chan os.Signal)
		defer close(exitSignals)

		signal.Notify(exitSignals,
			syscall.SIGINT,  // 2
			syscall.SIGTERM, // 15
			syscall.SIGQUIT, // 3
			syscall.SIGKILL, // 9
		)

		sig := <-exitSignals
		lg.Info("catch signal %v", sig)

		atomic.StoreInt32(&active, 0)
	}

	doTouchCtrlFile := func() {
		defer func() {
			if v := recover(); nil != v {
				lg.Error("error on touchCtrlFile %v", v)
			}
		}()

		lg.Info("doTouchCtrlFile(%s)", ctrlFileName)

		var (
			f *os.File
			e error
		)

		if !fileExists(ctrlFileName) {
			f, e = os.Create(ctrlFileName)
		} else {
			f, e = os.OpenFile(ctrlFileName, os.O_WRONLY, 0666)
		}

		if nil != e {
			lg.Error("error doTouchCtrlFile(%s) %s", ctrlFileName, e.Error())
		} else {
			w := bufio.NewWriter(f)
			w.WriteString(ctrlPID)
			w.Flush()
			f.Close()
		}
	}

	touchCtrlFileEverySecond := func() {
		tk := time.NewTicker(time.Second)
		defer tk.Stop()

		for atomic.LoadInt32(&active) > 0 {
			<-tk.C

			if atomic.LoadInt32(&active) > 0 {
				doTouchCtrlFile()
			}
		}
	}

	serviceForSlaveProcess := func() {
		for atomic.LoadInt32(&active) > 0 {
			lg.Info("slave process launch(%s -slave %s)", exe, ctrlPID)
			c := exec.Command(exe, "-slave", ctrlPID)
			if e := c.Run(); nil != e {
				lg.Error("error on run(%s) %s", exe, e.Error())
			}
		}
	}

	defer func() {
		if v := recover(); nil != v {
			lg.Error("error on run %v", v)
		}

		if fileExists(ctrlFileName) {
			os.Remove(ctrlFileName)
		}
	}()

	lg.Info("exec echo=%v", echo)
	go touchCtrlFileEverySecond()
	go serviceForSlaveProcess()
	waitForExit()
}

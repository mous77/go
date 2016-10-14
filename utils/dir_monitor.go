package utils

import (
	"github.com/howeyc/fsnotify"
	"path/filepath"
	"os"
	"fmt"
	"sync/atomic"
	"strings"
)

/**
 * 文件事件类型
 */
type TFileEventType int

const (
	FET_EXISTS TFileEventType = iota
	FET_CREATE
	FET_MODIFY
	FET_DELETE
)

func (this TFileEventType)String() (string) {
	switch this{
	case FET_CREATE: return "CREATE"
	case FET_MODIFY: return "MODIFY"
	case FET_DELETE: return "DELETE"
	case FET_EXISTS: return "EXISTS"
	default: return "ERROR"
	}
}

type tOnFileEvent func(_event_type TFileEventType, _file_path string, _user_data interface{})

type tWatchTask struct {
	monitor  *TDirMonitor
	maxLevel int // 监测子目录时，此字段标识最多几层子目录，0表示不监测子目录
	curLevel int // 监测子模块时，此字段表示当前在第几层
	root     string
	onEvent  tOnFileEvent
	isSubDir bool
	fileExt  string
	userData interface{}
	inWatch  int32
}

type tNotifyItem struct {
	wtask *tWatchTask
	etype TFileEventType
	fpath string
}

func (this *tWatchTask)genEvent(_type TFileEventType, _path string) {
	if this.isExtMatch(_path) {
		this.onEvent(_type, _path, this.userData)
	}
}

func (this *tWatchTask)addSubTask(_path string) {
	if this.curLevel < this.maxLevel {
		if sub_task, ok := this.monitor.addTask(_path,
			this.fileExt, this.onEvent,
			this.maxLevel, this.curLevel + 1,
			this.isSubDir, this.userData); ok {
			sub_task.start()
		}
	}
}

func (this *tWatchTask)loadExistsFiles() {
	walk_fun := func(path string, info os.FileInfo, err error) error {
		//lg.Debug("found %s", path)
		if nil == err && !info.IsDir() {
			path = strings.Replace(path, "\\", "/", -1)
			same_dir := this.root == strings.Replace(filepath.Dir(path), "\\", "/", -1)
			if same_dir && this.isExtMatch(path) {
				this.genEvent(FET_EXISTS, path)
			}
		}
		return err
	}
	filepath.Walk(this.root, walk_fun)
}

func (this *tWatchTask)isExtMatch(_path string) (bool) {
	if len(this.fileExt) > 0 {
		ext := filepath.Ext(_path)
		if ext != this.fileExt {
			return false
		}
	}
	return true
}

func isDir(_path string) (bool) {
	if finfo, err := os.Stat(_path); nil != err {
		return false
	} else {
		return finfo.IsDir()
	}
}

func (this *tWatchTask)start() {
	if atomic.CompareAndSwapInt32(&this.inWatch, 0, 1) {
		if !this.isSubDir {
			this.loadExistsFiles()
		}

		if err := this.monitor.watcher.Watch(this.root); nil != err {
			this.monitor.onError("error on watch %s %s", this.root, err.Error())
		}
	}
}

type tOnDirMonErrLogger func(string, ...interface{})

func defDirMonitorLog(_fmt string, _args ...interface{}) {
	fmt.Printf(_fmt, _args...)
}

type TDirMonitor struct {
	onError tOnDirMonErrLogger
	active  int32
	watcher *fsnotify.Watcher
	taskMap map[string]*tWatchTask
}

func NewDirMonitor() (*TDirMonitor) {
	return &TDirMonitor{
		taskMap:make(map[string]*tWatchTask),
		onError: defDirMonitorLog,
		active:0}
}

func (this *TDirMonitor)SetErrorLogger(_on_err tOnDirMonErrLogger) {
	if nil != _on_err {
		this.onError = _on_err
	}
}

func (this *TDirMonitor)isActive() (bool) {
	return atomic.LoadInt32(&this.active) > 0
}

func (this *TDirMonitor)delTask(_task *tWatchTask) {
	if _task.isSubDir {
		delete(this.taskMap, _task.root)
	}
}

func (this *TDirMonitor)addTask(_path string, _file_ext string, _listener tOnFileEvent, _max_level int, _cur_level int, _sub_dir bool, _user_data interface{}) (*tWatchTask, bool) {
	if task, ok := this.taskMap[_path]; ok {
		return task, false
	} else {
		task = &tWatchTask{root:_path,
			onEvent:_listener,
			maxLevel:_max_level,
			curLevel:_cur_level,
			isSubDir: _sub_dir,
			fileExt:_file_ext,
			userData:_user_data,
			inWatch:0,
			monitor:this}

		this.taskMap[_path] = task

		return task, true
	}
}

func (this *TDirMonitor)addSubdir(_root string, _ext string, _listener tOnFileEvent, _max_level int, _user_data interface{}) {
	walk_fun := func(_path string, info os.FileInfo, err error) error {
		if nil != err {
			this.onError("error on walkDir %s", err.Error())
			return err
		}

		if info.IsDir() {
			_path = strings.Replace(_path, "\\", "/", -1)
			dir_level := strings.Count(_path, "/") - strings.Count(_root, "/")

			if dir_level <= _max_level {
				this.addTask(_path, _ext, _listener, _max_level, dir_level, true, _user_data)
			}
		}

		return nil
	}

	filepath.Walk(_root, walk_fun)
}

func (this *TDirMonitor)Register(_path string, _file_ext string, _listener tOnFileEvent, _level int, _user_data interface{}) {
	if len(_path) == 0 {
		panic("path is null")
	}

	if nil == _listener {
		panic("listener is null")
	}

	if !isDir(_path) {
		panic("path not exists")
	}

	if len(_file_ext) > 0 && _file_ext[0] != '.' {
		_file_ext = "." + _file_ext
	}

	if 0 == _level {
		this.addTask(_path, _file_ext, _listener, _level, 0, false, _user_data)
	} else {
		this.addSubdir(_path, _file_ext, _listener, _level, _user_data)
	}
}

func (this *TDirMonitor)recvEvent(_event *fsnotify.FileEvent) {
	path := strings.Replace(_event.Name, "\\", "/", -1)
	dir := strings.Replace(filepath.Dir(path), "\\", "/", -1)

	var base *tWatchTask

	task, ok := this.taskMap[path]
	if ok {
		base = task
	} else {
		base, _ = this.taskMap[dir]
	}

	if _event.IsDelete() {
		if nil != task {
			this.delTask(task)
		} else {
			base.genEvent(FET_DELETE, path)
		}
	} else if _event.IsModify() {
		if nil == task {
			base.genEvent(FET_MODIFY, path)
		}
	} else if _event.IsCreate() {
		info, _ := os.Stat(path)
		if info.IsDir() {
			base.addSubTask(path)
		} else {
			base.genEvent(FET_CREATE, path)
		}
	} else if _event.IsRename() {
		if info, err := os.Stat(path); nil != err {
			if os.IsNotExist(err) {
				if nil != task {
					this.delTask(task)
				} else {
					base.genEvent(FET_DELETE, path)
				}
			}
		} else {
			if info.IsDir() {
				base.addSubTask(path)
			} else {
				base.genEvent(FET_CREATE, path)
			}
		}
	}
}

func (this *TDirMonitor)pullEvents() {
	for this.isActive() {
		select {
		case evt := <-watcher.Event:
			this.recvEvent(evt)
		case err := <-watcher.Error:
			this.onError("poll event error(%s)", err.Error())
		}
	}
}

func (this *TDirMonitor) Start() {
	if atomic.CompareAndSwapInt32(&this.active, 0, 1) {
		if watcher, err := fsnotify.NewWatcher(); nil != err {
			this.onError("error on newWatcher %s", err.Error())
			panic(err)
		} else {
			this.watcher = watcher

			go this.pullEvents()

			for _, task := range this.taskMap {
				task.start()
			}
		}
	}
}

func (this *TDirMonitor) Stop() {
	if atomic.CompareAndSwapInt32(&this.active, 1, 0) {
		if err := this.watcher.Close(); nil != err {
			this.onError("error on stop %s", err.Error())
		}
	}
}

func TestDirMonitor() {
	listener := func(_event_type TFileEventType, _file_path string, _user_data interface{}) {
		ud := _user_data.(string)
		fmt.Printf("=======onEvent(%s,%s,%s)\r\n", _event_type, _file_path, ud)
	}
	monitor := NewDirMonitor()
	monitor.Register("c:/tmp/2016", "txt", listener, 2, "2016")
	monitor.Register("c:/tmp/2017", "txt", listener, 2, "2017")
	monitor.Start()
}